package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	dao2 "ollama-desktop/internal/dao"
	ollama2 "ollama-desktop/internal/ollama"
	"strings"
	"time"
)

const (
	refTypeQuestion      = "question"
	refTypeAnswer        = "answer"
	messageRoleUser      = "user"
	messageRoleSystem    = "system"
	messageRoleAssistant = "assistant"
)

var dao = Dao{}

type Dao struct {
	dao *dao2.DbDao
}

func (d *Dao) startup(ctx context.Context) {
	if d.dao == nil {
		d.dao = &dao2.DbDao{}
	}
	d.dao.Startup(ctx)
}

func (d *Dao) shutdown() {
	if d.dao == nil {
		return
	}
	d.dao.Shutdown()
}

func (d *Dao) sessions() ([]*SessionModel, error) {
	sqlStr := `select id, session_name, model_name, prompts, message_history_count, use_stream, response_format, keep_alive,
                  options, session_type, created_at, updated_at
            from t_session
            order by created_at desc`
	rows, err := d.dao.GetDriver().Query(app.ctx, sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []*SessionModel
	for rows.Next() {
		session := &SessionModel{}
		if err := rows.Scan(&session.Id, &session.SessionName, &session.ModelName, &session.Prompts,
			&session.MessageHistoryCount, &session.UseStream, &session.ResponseFormat, &session.KeepAlive,
			&session.Options, &session.SessionType, &session.CreatedAt, &session.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (d *Dao) createSession(session *SessionModel) error {
	sqlStr := `insert into t_session(id, session_name, model_name, prompts, message_history_count, use_stream, response_format, keep_alive,
                  options, session_type, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	return d.dao.GetDriver().Execute(app.ctx, sqlStr, session.Id, session.SessionName, session.ModelName, session.Prompts,
		session.MessageHistoryCount, session.UseStream, session.ResponseFormat, session.KeepAlive,
		session.Options, session.SessionType, session.CreatedAt, session.UpdatedAt)
}

func (d *Dao) deleteSession(id string, sessions []*SessionModel) error {
	if _, err := d.findSession(id, sessions); err != nil {
		return err
	}

	tx, err := d.dao.GetDb().Begin()
	if err == nil {
		return err
	}
	err = func() error {
		// 删除会话
		sqlStr := "delete from t_session where id = ?"
		if _, err := tx.ExecContext(app.ctx, sqlStr, id); err != nil {
			return err
		}
		// 删除问题
		sqlStr = "delete from t_question where session_id = ?"
		if _, err := tx.ExecContext(app.ctx, sqlStr, id); err != nil {
			return err
		}
		// 删除回答
		sqlStr = "delete from t_answer where session_id = ?"
		if _, err := tx.ExecContext(app.ctx, sqlStr, id); err != nil {
			return err
		}
		// 删除图片
		sqlStr = "delete from t_image where session_id = ?"
		if _, err := tx.ExecContext(app.ctx, sqlStr, id); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		err = tx.Rollback()
	} else {
		err = tx.Commit()
	}

	return err
}

func (d *Dao) findSession(id string, sessions []*SessionModel) (*SessionModel, error) {
	if len(sessions) == 0 {
		return nil, errors.New("session not exists")
	}
	for _, session := range sessions {
		if session.Id == id {
			return session, nil
		}
	}
	return nil, errors.New("session not exists")
}

func (d *Dao) scanQuestion(rows *sql.Rows) (*QuestionModel, error) {
	question := &QuestionModel{}
	if err := rows.Scan(&question.Id, &question.SessionId, &question.QuestionContent, &question.AnswerCount,
		&question.MessageType, &question.HasImage, &question.CreatedAt, &question.UpdatedAt); err != nil {
		return nil, err
	}
	return question, nil
}

func (d *Dao) findQuestion(id string) (*QuestionModel, error) {
	sqlStr := `select id, session_id, question_content, answer_count, message_type, has_image, created_at, updated_at
            from t_question where id = ?`
	rows, err := d.dao.GetDriver().Query(app.ctx, sqlStr, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		return d.scanQuestion(rows)
	}
	return nil, errors.New("question not exists")
}

func (d *Dao) combineHistoryMessages(session *SessionModel, skipLast bool) ([]ollama2.Message, error) {
	if session.MessageHistoryCount < 1 {
		return nil, nil
	}
	offset := 0
	if skipLast {
		offset = 1
	}

	// 查询历史存在有效回答的消息
	sqlStr := `select id, session_id, question_content, answer_count, message_type, has_image, created_at, updated_at
            from t_question
            where session_id = ? and exists (select 1 from t_answer where t_question.id = t_answer.question_id and t_answer.is_success = 1)
            order by created_at desc
            limit ? offset ?`
	rows, err := d.dao.GetDriver().Query(app.ctx, sqlStr, session.Id, session.MessageHistoryCount, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var questions []*QuestionModel
	var questionIds []string
	values := []interface{}{
		session.Id,
	}
	for rows.Next() {
		question, err := d.scanQuestion(rows)
		if err != nil {
			return nil, err
		}
		questionIds = append(questionIds, question.Id)
		values = append(values, question.Id)
		questions = append(questions, question)
	}
	if len(questions) == 0 {
		return nil, nil
	}

	ids := strings.Join(questionIds, ", ")
	// 查询历史回答
	sqlStr = fmt.Sprintf(`select id, session_id, question_id, answer_content, message_role, total_duration, load_duration, 
                 prompt_eval_count, prompt_eval_duration, eval_count, eval_duration, done_reason, is_last,
                 is_success, has_image, created_at, updated_at
            from t_answer
            where session_id = ? is_last = 1 and is_success = 1 and question_id in (%s)`, ids)

	rows, err = d.dao.GetDriver().Query(app.ctx, sqlStr, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	answerMap := map[string]*AnswerModel{}

	for rows.Next() {
		answer := &AnswerModel{}
		if err := rows.Scan(&answer.Id, &answer.SessionId, &answer.QuestionId, &answer.AnswerContent,
			&answer.MessageRole, &answer.TotalDuration, &answer.LoadDuration, &answer.PromptEvalCount,
			&answer.PromptEvalDuration, &answer.EvalCount, &answer.EvalDuration, &answer.DoneReason,
			&answer.IsLast, &answer.IsSuccess, &answer.HasImage, &answer.CreatedAt, &answer.UpdatedAt); err != nil {
			return nil, err
		}
		answerMap[answer.QuestionId] = answer
	}

	var ollamaMessages []ollama2.Message
	for _, message := range questions {
		var images []ollama2.ImageData
		if message.HasImage {
			images, err = d.findImages(message.Id, refTypeQuestion)
			if err != nil {
				return nil, err
			}
		}
		// 问题
		ollamaMessages = append(ollamaMessages, ollama2.Message{
			Role:    messageRoleUser,
			Content: message.QuestionContent,
			Images:  images,
		})
		// 回答
		answer := answerMap[message.Id]
		if answer == nil {
			// 原则上不存在
			continue
		}
		images = nil
		if answer.HasImage {
			images, err = d.findImages(answer.Id, refTypeAnswer)
			if err != nil {
				return nil, err
			}
		}
		ollamaMessages = append(ollamaMessages, ollama2.Message{
			Role:    answer.MessageRole,
			Content: answer.AnswerContent,
			Images:  images,
		})
	}
	return ollamaMessages, nil
}

func (d *Dao) findImages(refId, refType string) ([]ollama2.ImageData, error) {
	sqlStr := fmt.Sprintf(`select id, session_id, ref_id, ref_type, image_data, created_at, updated_at
            from t_image
            where ref_id = ? and ref_type = ?
            order by created_at`)
	rows, err := d.dao.GetDriver().Query(app.ctx, sqlStr, refId, refType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var images []ollama2.ImageData

	for rows.Next() {
		image := &ImageModel{}
		if err := rows.Scan(&image.Id, &image.SessionId, &image.RefId, &image.RefType, &image.ImageData,
			&image.CreatedAt, &image.UpdatedAt); err != nil {
			return nil, err
		}
		images = append(images, image.ImageData)
	}
	return images, nil
}

func (d *Dao) createAnswer(question *QuestionModel, answer *AnswerModel, questionImages, answerImages []ollama2.ImageData, isNew bool) error {
	tx, err := d.dao.GetDb().Begin()
	if err == nil {
		return err
	}
	err = func() error {
		if isNew {
			// 保存问题
			sqlStr := `insert into t_question(id, session_id, question_content, answer_count, message_type, has_image, created_at, updated_at) 
                       values(?, ?, ?, ?, ?, ?, ?, ?)`
			if _, err := tx.ExecContext(app.ctx, sqlStr, question.Id, question.SessionId, question.QuestionContent,
				question.AnswerCount, question.MessageType, question.HasImage, question.CreatedAt, question.UpdatedAt); err != nil {
				return err
			}
			if question.HasImage {
				sqlStr := `insert into t_image(id, session_id, ref_id, ref_type, image_data, created_at, updated_at) 
                       values(?, ?, ?, ?, ?, ?, ?, ?)`
				stm, err := tx.PrepareContext(app.ctx, sqlStr)
				if err != nil {
					return err
				}
				defer stm.Close()
				for _, image := range questionImages {
					if _, err := stm.ExecContext(app.ctx, uuid.NewString(), question.SessionId, question.Id,
						refTypeQuestion, image, question.CreatedAt, question.CreatedAt); err != nil {
						return err
					}
				}
			}

		} else {
			// 修改问题
			sqlStr := `update t_question set answer_count = ?, updated_at = ? where id = ?`
			if _, err := tx.ExecContext(app.ctx, sqlStr, question.AnswerCount, question.UpdatedAt, question.Id); err != nil {
				return err
			}
			sqlStr = `update t_answer set is_last = 0, updated_at = ? where question_id = ?`
			if _, err := tx.ExecContext(app.ctx, sqlStr, question.UpdatedAt, question.Id); err != nil {
				return err
			}
		}
		// 保存答案
		sqlStr := `insert into t_answer(id, session_id, question_id, answer_content, message_role, total_duration, load_duration, 
                 prompt_eval_count, prompt_eval_duration, eval_count, eval_duration, done_reason, is_last,
                 is_success, has_image, created_at, updated_at) 
                       values(?, ?, ?, ?, ?, ?, ?, ?)`
		if _, err := tx.ExecContext(app.ctx, sqlStr, answer.Id, answer.SessionId, answer.QuestionId,
			answer.AnswerContent, answer.MessageRole, answer.TotalDuration, answer.LoadDuration, answer.PromptEvalCount,
			answer.PromptEvalDuration, answer.EvalCount, answer.EvalDuration, answer.DoneReason, answer.IsLast,
			answer.IsSuccess, answer.HasImage, answer.CreatedAt, answer.UpdatedAt); err != nil {
			return err
		}
		if answer.HasImage {
			sqlStr := `insert into t_image(id, session_id, ref_id, ref_type, image_data, created_at, updated_at) 
                       values(?, ?, ?, ?, ?, ?, ?, ?)`
			stm, err := tx.PrepareContext(app.ctx, sqlStr)
			if err != nil {
				return err
			}
			defer stm.Close()
			for _, image := range answerImages {
				if _, err := stm.ExecContext(app.ctx, uuid.NewString(), answer.SessionId, answer.Id,
					refTypeAnswer, image, answer.CreatedAt, answer.CreatedAt); err != nil {
					return err
				}
			}
		}
		return nil
	}()
	if err != nil {
		err = tx.Rollback()
	} else {
		err = tx.Commit()
	}
	return err
}

func (d *Dao) configs() (map[string]string, error) {
	sqlStr := `select config_key, config_value from t_config`
	rows, err := d.dao.GetDriver().Query(app.ctx, sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	configs := make(map[string]string)
	for rows.Next() {
		var configKey, configValue string
		if err := rows.Scan(&configKey, &configValue); err != nil {
			return nil, err
		}
		configs[configKey] = configValue
	}
	return configs, nil
}

type ConfigModel struct {
	ConfigKey   string    `json:"configKey"`
	ConfigValue string    `json:"configValue"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SessionModel struct {
	Id                  string        `json:"id"`
	SessionName         string        `json:"sessionName"`
	ModelName           string        `json:"modelName"`
	Prompts             string        `json:"prompts,omitempty"`
	MessageHistoryCount int           `json:"messageHistoryCount"`
	UseStream           bool          `json:"stream,omitempty"`
	ResponseFormat      string        `json:"responseFormat,omitempty"`
	KeepAlive           time.Duration `json:"keepAlive,omitempty"`
	Options             string        `json:"options,omitempty"`
	SessionType         string        `json:"sessionType"`
	CreatedAt           time.Time     `json:"createdAt"`
	UpdatedAt           time.Time     `json:"updatedAt"`
}

type QuestionModel struct {
	Id              string    `json:"id"`
	SessionId       string    `json:"sessionId"`
	QuestionContent string    `json:"questionContent"`
	AnswerCount     int       `json:"answerCount"`
	MessageType     string    `json:"messageType"`
	HasImage        bool      `json:"hasImage"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type AnswerModel struct {
	Id                 string        `json:"id"`
	SessionId          string        `json:"sessionId"`
	QuestionId         string        `json:"questionId"`
	AnswerContent      string        `json:"answerContent"`
	MessageRole        string        `json:"messageRole"`
	TotalDuration      time.Duration `json:"totalDuration"`
	LoadDuration       time.Duration `json:"loadDuration"`
	PromptEvalCount    int           `json:"promptEvalCount"`
	PromptEvalDuration time.Duration `json:"promptEvalDuration"`
	EvalCount          int           `json:"evalCount"`
	EvalDuration       time.Duration `json:"evalDuration"`
	DoneReason         string        `json:"doneReason"`
	IsLast             bool          `json:"isLast"`
	IsSuccess          bool          `json:"isSuccess"`
	HasImage           bool          `json:"hasImage"`
	CreatedAt          time.Time     `json:"createdAt"`
	UpdatedAt          time.Time     `json:"updatedAt"`
}

type ImageModel struct {
	Id        string    `json:"id"`
	SessionId string    `json:"sessionId"`
	RefId     string    `json:"refId"`
	RefType   string    `json:"refType"`
	ImageData []byte    `json:"imageData"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
