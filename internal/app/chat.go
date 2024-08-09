package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"ollama-desktop/internal/log"
	olm "ollama-desktop/internal/ollama"
	"strconv"
	"time"
)

var chat = Chat{}

type Chat struct {
}

func (c *Chat) scanSession(rows *sql.Rows) (*SessionModel, error) {
	session := &SessionModel{}
	if err := rows.Scan(&session.Id, &session.SessionName, &session.ModelName,
		&session.MessageHistoryCount, &session.KeepAlive, &session.SystemMessage, &session.Options, &session.CreatedAt, &session.UpdatedAt); err != nil {
		return nil, err
	}
	return session, nil
}

func (c *Chat) Sessions() ([]*SessionModel, error) {
	sqlStr := `select id, session_name, model_name, message_history_count, keep_alive, system_message, options, created_at, updated_at
            from t_session
            order by created_at desc`
	rows, err := dao.db().QueryContext(app.ctx, sqlStr)
	if err != nil {
		log.Error().Err(err).Msg("query session error")
		return nil, err
	}
	defer rows.Close()
	var sessions []*SessionModel
	for rows.Next() {
		session, err := c.scanSession(rows)
		if err != nil {
			log.Error().Err(err).Msg("fill session info")
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (c *Chat) CreateSession(session *SessionModel) (*SessionModel, error) {
	session.Id = uuid.NewString()
	session.CreatedAt = time.Now()
	session.UpdatedAt = session.CreatedAt

	sqlStr := `insert into t_session(id, session_name, model_name, message_history_count, keep_alive, system_message, options, created_at, updated_at)
               values (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := dao.db().ExecContext(app.ctx, sqlStr, session.Id, session.SessionName, session.ModelName,
		session.MessageHistoryCount, session.KeepAlive, session.SystemMessage, session.Options, session.CreatedAt, session.UpdatedAt)
	return session, err
}

func (c *Chat) DeleteSession(id string) (string, error) {
	return id, dao.transaction(func(tx *sql.Tx) error {
		// 删除会话
		sqlStr := "delete from t_session where id = ?"
		if _, err := tx.ExecContext(app.ctx, sqlStr, id); err != nil {
			log.Error().Err(err).Msg("delete session error")
			return err
		}
		// 删除聊天
		sqlStr = "delete from t_chat_message where session_id = ?"
		if _, err := tx.ExecContext(app.ctx, sqlStr, id); err != nil {
			log.Error().Err(err).Msg("delete session chat error")
			return err
		}
		return nil
	})
}

func (c *Chat) UpdateSession(session *SessionModel) (*SessionModel, error) {
	session.UpdatedAt = session.CreatedAt

	sqlStr := `update t_session set session_name = ?, model_name = ?, message_history_count = ?, keep_alive = ?, system_message = ?, options = ?, updated_at = ?
               where id = ?`
	_, err := dao.db().ExecContext(app.ctx, sqlStr, session.SessionName, session.ModelName,
		session.MessageHistoryCount, session.KeepAlive, session.SystemMessage, session.Options, session.UpdatedAt, session.Id)
	return session, err
}

func (c *Chat) GetSession(id string) (*SessionModel, error) {
	sqlStr := `select id, session_name, model_name, message_history_count, keep_alive, system_message, options, created_at, updated_at
            from t_session
            where id = ?`
	rows, err := dao.db().QueryContext(app.ctx, sqlStr, id)
	if err != nil {
		log.Error().Err(err).Msg("query session error")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		session, err := c.scanSession(rows)
		if err != nil {
			log.Error().Err(err).Msg("fill session error")
		}
		return session, err
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

func (c *Chat) scanChatMessage(rows *sql.Rows) (*ChatMessageModel, error) {
	chatMessage := &ChatMessageModel{}
	if err := rows.Scan(&chatMessage.Id, &chatMessage.SessionId, &chatMessage.QuestionContent, &chatMessage.AnswerContent,
		&chatMessage.TotalDuration, &chatMessage.LoadDuration, &chatMessage.PromptEvalCount,
		&chatMessage.PromptEvalDuration, &chatMessage.EvalCount, &chatMessage.EvalDuration, &chatMessage.DoneReason,
		&chatMessage.IsSuccess, &chatMessage.CreatedAt, &chatMessage.UpdatedAt); err != nil {
		return nil, err
	}
	return chatMessage, nil
}

type SessionHistoryMessageRequest struct {
	SessionId  string `json:"sessionId"`
	NextMarker string `json:"nextMarker"`
}

func (c *Chat) SessionHistoryMessages(request *SessionHistoryMessageRequest) ([]*ChatMessage, error) {
	var timeMarker time.Time
	if request.NextMarker != "" {
		// 查询历史存在有效回答的消息
		sqlStr := `select created_at from t_chat_message where session_id = ? and id = ?`
		rows, err := dao.db().QueryContext(app.ctx, sqlStr, request.SessionId, request.NextMarker)
		if err != nil {
			log.Error().Err(err).Msg("query chat message with marker error")
			return nil, err
		}
		defer rows.Close()
		if rows.Next() {
			err := rows.Scan(&timeMarker)
			if err != nil {
				log.Error().Err(err).Msg("fill time marker error")
				return nil, err
			}
		}
	}
	if timeMarker.IsZero() {
		timeMarker = time.Now()
	}

	sqlStr := `select id, session_id, question_content, answer_content, total_duration, load_duration, 
                 prompt_eval_count, prompt_eval_duration, eval_count, eval_duration, done_reason, is_success, created_at, updated_at
            from t_chat_message
            where session_id = ? and created_at < ?
            order by created_at desc
            limit ?`
	rows, err := dao.db().QueryContext(app.ctx, sqlStr, request.SessionId, timeMarker, 50)
	if err != nil {
		log.Error().Err(err).Msg("query chat message error")
		return nil, err
	}
	defer rows.Close()
	var chatMessages []*ChatMessageModel
	for rows.Next() {
		item, err := c.scanChatMessage(rows)
		if err != nil {
			log.Error().Err(err).Msg("fill chat message error")
			return nil, err
		}
		chatMessages = append(chatMessages, item)
	}
	if len(chatMessages) == 0 {
		return nil, nil
	}

	var messages []*ChatMessage
	for i := len(chatMessages) - 1; i >= 0; i-- {
		message := chatMessages[i]
		messages = append(messages, &ChatMessage{
			Id:        message.Id,
			SessionId: message.SessionId,
			Role:      messageRoleUser,
			Content:   message.QuestionContent,
			Success:   true,
			CreatedAt: message.CreatedAt,
		})
		// 回答
		messages = append(messages, &ChatMessage{
			Id:        message.Id,
			SessionId: message.SessionId,
			Role:      messageRoleAssistant,
			Content:   message.AnswerContent,
			Success:   message.IsSuccess,
			CreatedAt: message.CreatedAt,
		})
	}
	return messages, nil
}

type ConversationRequest struct {
	SessionId string `json:"sessionId"`
	Content   string `json:"content"`
}

type ConversationResponse struct {
	Id        string    `json:"id"`
	SessionId string    `json:"sessionId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func (c *Chat) Conversation(request *ConversationRequest) (*ConversationResponse, error) {
	session, err := c.GetSession(request.SessionId)
	if err != nil {
		log.Error().Err(err).Msg("get session error")
		return nil, err
	}

	message := &ChatMessageModel{
		Id:              uuid.NewString(),
		SessionId:       session.Id,
		QuestionContent: request.Content,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	go c.chat(session, message)
	return &ConversationResponse{
		Id:        message.Id,
		SessionId: message.SessionId,
		Content:   message.QuestionContent,
		CreatedAt: message.CreatedAt,
	}, nil
}

func (c *Chat) createChatMessage(message *ChatMessageModel) error {
	sqlStr := `insert into t_chat_message(id, session_id, question_content, answer_content, total_duration, load_duration, 
                   prompt_eval_count, prompt_eval_duration, eval_count, eval_duration, done_reason,
                   is_success, created_at, updated_at) 
               values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := dao.db().ExecContext(app.ctx, sqlStr, message.Id, message.SessionId, message.QuestionContent,
		message.AnswerContent, message.TotalDuration, message.LoadDuration,
		message.PromptEvalCount, message.PromptEvalDuration, message.EvalCount, message.EvalDuration, message.DoneReason,
		message.IsSuccess, message.CreatedAt, message.UpdatedAt); err != nil {
		log.Error().Err(err).Msg("create chat message error")
		return err
	}
	return nil
}

func (c *Chat) combineHistoryMessages(session *SessionModel) ([]olm.Message, error) {
	var ollamaMessages []olm.Message
	if session.SystemMessage != "" {
		ollamaMessages = append(ollamaMessages, olm.Message{
			Role:    messageRoleSystem,
			Content: session.SystemMessage,
			Images:  nil,
		})
	}
	if session.MessageHistoryCount < 1 {
		return ollamaMessages, nil
	}
	sqlStr := `select id, session_id, question_content, answer_content, total_duration, load_duration, 
                 prompt_eval_count, prompt_eval_duration, eval_count, eval_duration, done_reason,
                 is_success, created_at, updated_at
            from t_chat_message
            where session_id = ? and is_success = 1
            order by created_at desc
            limit ?`
	rows, err := dao.db().QueryContext(app.ctx, sqlStr, session.Id, session.MessageHistoryCount)
	if err != nil {
		log.Error().Err(err).Msg("query history chat message error")
		return nil, err
	}
	defer rows.Close()
	var messages []*ChatMessageModel
	for rows.Next() {
		message, err := c.scanChatMessage(rows)
		if err != nil {
			log.Error().Err(err).Msg("fill chat message error")
			return nil, err
		}
		messages = append(messages, message)
	}
	if len(messages) == 0 {
		return nil, nil
	}

	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		// 问题
		ollamaMessages = append(ollamaMessages, olm.Message{
			Role:    messageRoleUser,
			Content: message.QuestionContent,
			Images:  nil,
		})
		// 回答
		ollamaMessages = append(ollamaMessages, olm.Message{
			Role:    messageRoleAssistant,
			Content: message.AnswerContent,
			Images:  nil,
		})
	}
	return ollamaMessages, nil
}

func (c *Chat) emitChatError(message *ChatMessageModel, err error) {
	message.IsSuccess = false
	message.DoneReason = err.Error()
	message.UpdatedAt = time.Now()
	message.AnswerContent = err.Error()
	runtime.EventsEmit(app.ctx, message.Id, err.Error(), true, false)
}

func (c *Chat) chat(session *SessionModel, message *ChatMessageModel) {
	defer c.createChatMessage(message)
	messages, err := c.combineHistoryMessages(session)
	if err != nil {
		c.emitChatError(message, err)
		return
	}
	messages = append(messages, olm.Message{
		Role:    messageRoleUser,
		Content: message.QuestionContent,
		Images:  nil,
	})
	var keepAlive *olm.Duration
	if session.KeepAlive != "" {
		duration, err := time.ParseDuration(session.KeepAlive)
		if err != nil {
			c.emitChatError(message, err)
			return
		}
		keepAlive = &olm.Duration{
			Duration: duration,
		}
	}
	var options map[string]interface{}
	if session.Options != "" {
		var sessionOptions map[string]string
		if err := json.Unmarshal([]byte(session.Options), &sessionOptions); err != nil {
			c.emitChatError(message, err)
			return
		}
		options = make(map[string]interface{})
		for name, value := range sessionOptions {
			if value == "" {
				continue
			}
			switch name {
			case "seed":
				options["seed"], _ = strconv.Atoi(value)
			case "numPredict":
				options["num_predict"], _ = strconv.Atoi(value)
			case "topK":
				options["top_k"], _ = strconv.Atoi(value)
			case "topP":
				options["top_p"], _ = strconv.ParseFloat(value, 32)
			case "numCtx":
				options["num_ctx"], _ = strconv.Atoi(value)
			case "temperature":
				options["temperature"], _ = strconv.ParseFloat(value, 32)
			case "repeatPenalty":
				options["repeat_penalty"], _ = strconv.ParseFloat(value, 32)
			}
		}
	}

	var buffer bytes.Buffer

	request := &olm.ChatRequest{
		Model:     session.ModelName,
		Messages:  messages,
		KeepAlive: keepAlive,
		Options:   options,
	}
	log.Debug().Any("request", request).Msg("chat request")

	err = ollama.newApiClient().Chat(app.ctx, request, func(response olm.ChatResponse) error {
		respMessage := response.Message
		buffer.WriteString(respMessage.Content)
		fullContent := buffer.String()
		if response.Done {
			message.UpdatedAt = response.CreatedAt
			message.IsSuccess = true
			message.AnswerContent = fullContent
			message.DoneReason = response.DoneReason
			metrics := response.Metrics
			message.TotalDuration = metrics.TotalDuration
			message.LoadDuration = metrics.LoadDuration
			message.PromptEvalCount = metrics.PromptEvalCount
			message.PromptEvalDuration = metrics.PromptEvalDuration
			message.EvalCount = metrics.EvalCount
			message.EvalDuration = metrics.EvalDuration
		}
		runtime.EventsEmit(app.ctx, message.Id, fullContent, response.Done, true)
		return nil
	})
	if err != nil {
		c.emitChatError(message, err)
	}
}
