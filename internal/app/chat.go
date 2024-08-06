package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	olm "ollama-desktop/internal/ollama"
	"strings"
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
		&session.MessageHistoryCount, session.Options, &session.CreatedAt, &session.UpdatedAt); err != nil {
		return nil, err
	}
	return session, nil
}

func (c *Chat) Sessions() ([]*SessionModel, error) {
	sqlStr := `select id, session_name, model_name, prompts, message_history_count, options, created_at, updated_at
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

	sqlStr := `insert into t_session(id, session_name, model_name, prompts, message_history_count, options, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := dao.db().ExecContext(app.ctx, sqlStr, session.Id, session.SessionName, session.ModelName, session.Prompts,
		session.MessageHistoryCount, session.Options, session.CreatedAt, session.UpdatedAt)
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
	sqlStr := `select id, session_name, model_name, prompts, message_history_count, options, created_at, updated_at
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
		// 查询历史存在有效回答的消息
		sqlStr := `select created_at from t_question where session_id = ? and id = ?`
		rows, err := dao.db().QueryContext(app.ctx, sqlStr, request.SessionId, request.NextMarker)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		if rows.Next() {
			rows.Scan(&timeMarker)
		}
	}
	if timeMarker.IsZero() {
		timeMarker = time.Now()
	}

	sqlStr := `select id, session_id, question_content, answer_count, message_type, created_at, updated_at
            from t_question
            where session_id = ? and created_at < ?
            order by created_at desc
            limit ?`
	rows, err := dao.db().QueryContext(app.ctx, sqlStr, request.SessionId, timeMarker, 30)
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
		question, err := c.scanQuestion(rows)
		if err != nil {
			return nil, err
		}
		questionIds = append(questionIds, question.Id)
		values = append(values, question.Id)
		questions = append(questions, question)
	}
	if len(questionIds) == 0 {
		return nil, err
	}

	ids := strings.Join(questionIds, ", ")
	// 查询历史回答
	sqlStr = fmt.Sprintf(`select id, session_id, question_id, answer_content, message_role, total_duration, load_duration, 
                 prompt_eval_count, prompt_eval_duration, eval_count, eval_duration, done_reason, is_success, created_at, updated_at
            from t_answer
            where session_id = ? and question_id in (%s)`, ids)

	rows, err = dao.db().QueryContext(app.ctx, sqlStr, values...)
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
			&answer.IsSuccess, &answer.CreatedAt, &answer.UpdatedAt); err != nil {
			return nil, err
		}
		answerMap[answer.QuestionId] = answer
	}

	var messages []*ChatMessage
	for i := len(questions) - 1; i >= 0; i-- {
		question := questions[i]
		messages = append(messages, &ChatMessage{
			Id:        question.Id,
			SessionId: question.SessionId,
			Role:      messageRoleUser,
			Content:   question.QuestionContent,
			Success:   true,
			CreatedAt: question.CreatedAt,
		})
		// 回答
		answer := answerMap[question.Id]
		if answer == nil {
			messages = append(messages, &ChatMessage{
				Id:        uuid.NewString(),
				SessionId: question.SessionId,
				Role:      messageRoleAssistant,
				Content:   "",
				Success:   false,
				CreatedAt: question.CreatedAt,
			})
			continue
		}
		messages = append(messages, &ChatMessage{
			Id:        answer.Id,
			SessionId: question.SessionId,
			Role:      answer.MessageRole,
			Content:   answer.AnswerContent,
			Success:   answer.IsSuccess,
			CreatedAt: answer.CreatedAt,
		})
	}
	return messages, nil
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
	session, err := c.GetSession(request.SessionId)
	if err != nil {
		return nil, err
	}

	question := &QuestionModel{
		Id:              uuid.NewString(),
		SessionId:       session.Id,
		QuestionContent: request.Content,
		AnswerCount:     0,
		MessageType:     messageTypeChat,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	answer := &AnswerModel{
		Id:         uuid.NewString(),
		SessionId:  session.Id,
		QuestionId: question.Id,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	go c.chat(session, question, answer)
	return &ConversationResponse{SessionId: session.Id, QuestionId: question.Id, AnswerId: answer.Id}, nil
}

func (c *Chat) createQuestionAnswer(question *QuestionModel, answer *AnswerModel) error {
	return dao.transaction(func(tx *sql.Tx) error {
		// 保存问题
		sqlStr := `insert into t_question(id, session_id, question_content, answer_count, message_type, created_at, updated_at) 
                       values(?, ?, ?, ?, ?, ?, ?)`
		if _, err := tx.ExecContext(app.ctx, sqlStr, question.Id, question.SessionId, question.QuestionContent,
			question.AnswerCount, question.MessageType, question.CreatedAt, question.UpdatedAt); err != nil {
			return err
		}

		sqlStr = `insert into t_answer(id, session_id, question_id, answer_content, message_role, total_duration, load_duration, 
                 prompt_eval_count, prompt_eval_duration, eval_count, eval_duration, done_reason,
                 is_success, created_at, updated_at) 
                       values(?, ?, ?, ?, ?, ?, ?, ?)`
		if _, err := tx.ExecContext(app.ctx, sqlStr, answer.Id, answer.SessionId, answer.QuestionId,
			answer.AnswerContent, answer.MessageRole, answer.TotalDuration, answer.LoadDuration, answer.PromptEvalCount,
			answer.PromptEvalDuration, answer.EvalCount, answer.EvalDuration, answer.DoneReason,
			answer.IsSuccess, answer.CreatedAt, answer.UpdatedAt); err != nil {
			return err
		}
		return nil
	})
}

func (c *Chat) combineHistoryMessages(session *SessionModel) ([]olm.Message, error) {
	if session.MessageHistoryCount < 1 {
		return nil, nil
	}
	// 查询历史存在有效回答的消息
	sqlStr := `select id, session_id, question_content, answer_count, message_type, created_at, updated_at
            from t_question
            where session_id = ? and exists (select 1 from t_answer where t_question.id = t_answer.question_id and t_answer.is_success = 1)
            order by created_at desc
            limit ?`
	rows, err := dao.db().QueryContext(app.ctx, sqlStr, session.Id, session.MessageHistoryCount)
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
		question, err := c.scanQuestion(rows)
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
                 is_success, created_at, updated_at
            from t_answer
            where session_id = ? and is_success = 1 and question_id in (%s)`, ids)

	rows, err = dao.db().QueryContext(app.ctx, sqlStr, values...)
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
			&answer.IsSuccess, &answer.CreatedAt, &answer.UpdatedAt); err != nil {
			return nil, err
		}
		answerMap[answer.QuestionId] = answer
	}

	var ollamaMessages []olm.Message
	//  ) _, message := range questions
	for i := len(questions) - 1; i >= 0; i-- {
		message := questions[i]
		// 回答
		answer := answerMap[message.Id]
		if answer == nil {
			// 原则上不存在
			continue
		}
		// 问题
		ollamaMessages = append(ollamaMessages, olm.Message{
			Role:    messageRoleUser,
			Content: message.QuestionContent,
			Images:  nil,
		})
		// 回答
		ollamaMessages = append(ollamaMessages, olm.Message{
			Role:    answer.MessageRole,
			Content: answer.AnswerContent,
			Images:  nil,
		})
	}
	return ollamaMessages, nil
}

func (c *Chat) chat(session *SessionModel, question *QuestionModel, answer *AnswerModel) {
	defer c.createQuestionAnswer(question, answer)
	var keepAlive *olm.Duration
	if session.KeepAlive > 0 {
		keepAlive = &olm.Duration{
			Duration: session.KeepAlive,
		}
	}
	messages, err := c.combineHistoryMessages(session)
	if err != nil {
		answer.IsSuccess = false
		answer.DoneReason = err.Error()
		runtime.EventsEmit(app.ctx, answer.Id, nil, true, false)
	}
	messages = append(messages, olm.Message{
		Role:    messageRoleUser,
		Content: question.QuestionContent,
		Images:  nil,
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
		runtime.EventsEmit(app.ctx, answer.Id, buffer.String(), response.Done, true)
		return nil
	})
	if err != nil {
		answer.IsSuccess = false
		answer.DoneReason = err.Error()
		runtime.EventsEmit(app.ctx, answer.Id, nil, true, false)
	}
}
