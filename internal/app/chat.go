package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	olm "ollama-desktop/internal/ollama"
	"time"
)

const (
	messageTypeChat = "chat"
)

var chat = Chat{}

type Chat struct {
}

func (c *Chat) scanSession(rows *sql.Rows) (*SessionModel, error) {
	session := &SessionModel{}
	if err := rows.Scan(&session.Id, &session.SessionName, &session.ModelName, &session.Prompts,
		&session.MessageHistoryCount, &session.UseStream, &session.ResponseFormat, &session.KeepAlive,
		&session.Options, &session.SessionType, &session.CreatedAt, &session.UpdatedAt); err != nil {
		return nil, err
	}
	return session, nil
}

func (c *Chat) Sessions() ([]*SessionModel, error) {
	sqlStr := `select id, session_name, model_name, prompts, message_history_count, use_stream, response_format, keep_alive,
                  options, session_type, created_at, updated_at
            from t_session
            order by created_at desc`
	rows, err := dao.db().QueryContext(app.ctx, sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []*SessionModel
	for rows.Next() {
		session, err := c.scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (c *Chat) CreateSession(requestStr string) (*SessionModel, error) {
	session := &SessionModel{}
	if err := json.Unmarshal([]byte(requestStr), session); err != nil {
		return nil, err
	}
	session.Id = uuid.NewString()
	session.CreatedAt = time.Now()
	session.UpdatedAt = session.CreatedAt

	sqlStr := `insert into t_session(id, session_name, model_name, prompts, message_history_count, use_stream, response_format, keep_alive,
                  options, session_type, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := dao.db().ExecContext(app.ctx, sqlStr, session.Id, session.SessionName, session.ModelName, session.Prompts,
		session.MessageHistoryCount, session.UseStream, session.ResponseFormat, session.KeepAlive,
		session.Options, session.SessionType, session.CreatedAt, session.UpdatedAt)
	return session, err
}

func (c *Chat) DeleteSession(id string) (string, error) {
	return id, dao.transaction(func(tx *sql.Tx) error {
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
		return nil
	})
}

func (c *Chat) GetSession(id string) (*SessionModel, error) {
	sqlStr := `select id, session_name, model_name, prompts, message_history_count, use_stream, response_format, keep_alive,
                  options, session_type, created_at, updated_at
            from t_session
            where id = ?`
	rows, err := dao.db().QueryContext(app.ctx, sqlStr, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		return c.scanSession(rows)
	}
	return nil, errors.New("session not exists")
}

type ChatMessage struct {
	Id        string    `json:"id"`
	SessionId string    `json:"sessionId"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"createdAt"`
}

func (c *Chat) scanQuestion(rows *sql.Rows) (*QuestionModel, error) {
	question := &QuestionModel{}
	if err := rows.Scan(&question.Id, &question.SessionId, &question.QuestionContent, &question.AnswerCount,
		&question.MessageType, &question.CreatedAt, &question.UpdatedAt); err != nil {
		return nil, err
	}
	return question, nil
}

func (c *Chat) SessionHistoryMessages(requestStr string) ([]*ChatMessage, error) {
	request := &struct {
		SessionId  string `json:"sessionId"`
		NextMarker string `json:"nextMarker"`
	}{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}

	var timeMarker time.Time
	if request.NextMarker != "" {

	} else {
		timeMarker = time.Now()
	}
	time.ParseInLocation(, request.NextMarker, time.Local)

	// 查询历史存在有效回答的消息
	sqlStr := `select id, session_id, question_content, answer_count, message_type, created_at, updated_at
            from t_question
            where session_id = ? and created_at < select
            order by created_at desc
            limit ?`
	rows, err := dao.db().QueryContext(app.ctx, sqlStr, request.SessionId, 30)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var questions []*QuestionModel
	var questionIds []string
	values := []interface{}{
		request.SessionId,
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
	return dao.sessionHistoryMessages(request.SessionId, request.NextMarker)
}

type ConversationResponse struct {
	SessionId  string `json:"sessionId"`
	QuestionId string `json:"questionId"`
	AnswerId   string `json:"messageId"`
}

func (c *Chat) Conversation(requestStr string) (*ConversationResponse, error) {
	request := &struct {
		SessionId string `json:"sessionId"`
		Content   string `json:"content"`
	}{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}
	session, err := dao.findSession(request.SessionId, sessions)
	if err != nil {
		return nil, err
	}
	isNew := true
	var question *QuestionModel
	var images []olm.ImageData
	// 存在消息编号，查找历史消息
	if request.QuestionId != "" {
		question, err = dao.findQuestion(request.QuestionId)
		if err != nil {
			return nil, err
		}
		if question.HasImage {
			images, err = dao.findImages(question.Id, refTypeQuestion)
			if err != nil {
				return nil, err
			}
		}

		isNew = false
	} else {
		if err != nil {
			return nil, err
		}
		question = &QuestionModel{
			Id:              uuid.NewString(),
			SessionId:       session.Id,
			QuestionContent: request.Content,
			AnswerCount:     0,
			MessageType:     messageTypeChat,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
	}

	answer := &AnswerModel{
		Id:         uuid.NewString(),
		SessionId:  session.Id,
		QuestionId: question.Id,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	go c.chat(session, question, answer, images, isNew)
	return &ConversationResponse{SessionId: session.Id, QuestionId: question.Id, AnswerId: answer.Id}, nil
}

func (c *Chat) chat(session *SessionModel, question *QuestionModel, answer *AnswerModel, images []olm.ImageData, isNew bool) {
	var answerImages []olm.ImageData
	defer dao.createAnswer(question, answer, images, answerImages, isNew)
	var keepAlive *olm.Duration
	if session.KeepAlive > 0 {
		keepAlive = &olm.Duration{
			Duration: session.KeepAlive,
		}
	}
	messages, err := dao.combineHistoryMessages(session, !isNew)
	if err != nil {
		runtime.EventsEmit(app.ctx, answer.Id, nil, err)
	}
	messages = append(messages, olm.Message{
		Role:    messageRoleUser,
		Content: question.QuestionContent,
		Images:  images,
	})
	var buffer bytes.Buffer
	err = ollama.newApiClient().Chat(app.ctx, &olm.ChatRequest{
		Model:     session.ModelName,
		Messages:  messages,
		Stream:    &session.UseStream,
		Format:    session.ResponseFormat,
		KeepAlive: keepAlive,
		Options:   nil,
	}, func(response olm.ChatResponse) error {
		message := response.Message
		buffer.WriteString(message.Content)
		if len(message.Images) > 0 {
			answerImages = append(answerImages, message.Images...)
		}
		if response.Done {
			answer.MessageRole = message.Role
			answer.UpdatedAt = response.CreatedAt
			answer.IsSuccess = true
			answer.DoneReason = response.DoneReason
			metrics := response.Metrics
			answer.TotalDuration = metrics.TotalDuration
			answer.LoadDuration = metrics.LoadDuration
			answer.PromptEvalCount = metrics.PromptEvalCount
			answer.PromptEvalDuration = metrics.PromptEvalDuration
			answer.EvalCount = metrics.EvalCount
			answer.EvalDuration = metrics.EvalDuration
			question.AnswerCount = question.AnswerCount + 1
		}
		runtime.EventsEmit(app.ctx, answer.Id, response)
		return nil
	})
	if err != nil {
		runtime.EventsEmit(app.ctx, answer.Id, nil, err)
	}
}
