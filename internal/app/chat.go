package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	olm "ollama-desktop/internal/ollama"
	"ollama-desktop/internal/ollama/api"
	"time"
)

var chat = Chat{}

type Chat struct {
	sessions []*SessionModel
}

func (c *Chat) Sessions(forceUpdate bool) ([]*SessionModel, error) {
	if c.sessions != nil && !forceUpdate {
		return c.sessions, nil
	}
	var err error
	c.sessions, err = dao.sessions()
	return c.sessions, err
}

func (c *Chat) CreateSession(requestStr string) (*SessionModel, error) {
	session := &SessionModel{}
	if err := json.Unmarshal([]byte(requestStr), session); err != nil {
		return nil, err
	}
	session.Id = uuid.NewString()
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	if err := dao.createSession(session); err != nil {
		return nil, err
	}

	c.Sessions(true)
	return session, nil
}

func (c *Chat) DeleteSession(id string) (string, error) {
	sessions, err := c.Sessions(false)
	if err != nil {
		return id, err
	}
	if err := dao.deleteSession(id, sessions); err != nil {
		return id, err
	}
	c.Sessions(true)

	return id, nil
}

func (c *Chat) Conversation(requestStr string) (*ConversationResponse, error) {
	sessions, err := c.Sessions(false)
	if err != nil {
		return nil, err
	}
	request := &ConversationRequest{}
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
		images, err = c.convertImageDatas(request.Images)
		if err != nil {
			return nil, err
		}
		question = &QuestionModel{
			Id:              uuid.NewString(),
			SessionId:       session.Id,
			QuestionContent: request.Content,
			AnswerCount:     0,
			MessageType:     "chat",
			HasImage:        len(images) > 0,
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

func (c *Chat) convertImageDatas(images []string) ([]olm.ImageData, error) {
	if len(images) == 0 {
		return nil, nil
	}
	var list []olm.ImageData
	for _, image := range images {
		data, err := base64.StdEncoding.DecodeString(image)
		if err == nil {
			return nil, err
		}
		list = append(list, data)
	}
	return list, nil
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
	err = api.ClientFromConfig().Chat(app.ctx, &olm.ChatRequest{
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
			answer.HasImage = true
			answerImages = append(answerImages, message.Images...)
		}
		if response.Done {
			answer.MessageRole = message.Role
			answer.UpdatedAt = response.CreatedAt
			answer.IsSuccess = true
			answer.IsLast = true
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

type ConversationRequest struct {
	SessionId string `json:"sessionId"`
	// 有值表示重新生成
	QuestionId string   `json:"questionId"`
	Content    string   `json:"content"`
	Images     []string `json:"images"`
}

type ConversationResponse struct {
	SessionId  string `json:"sessionId"`
	QuestionId string `json:"questionId"`
	AnswerId   string `json:"messageId"`
}
