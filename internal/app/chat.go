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
}

func (c *Chat) Conversation(requestStr string) (*ConversationResponse, error) {
	request := &ConversationRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}
	session, err := dao.FindSession(request.SessionId)
	if err != nil {
		return nil, err
	}
	isNew := true
	var question *Question
	// 存在消息编号，查找历史消息
	if request.QuestionId != "" {
		question, err = dao.FindQuestion(request.QuestionId)
		if err != nil {
			return nil, err
		}
		isNew = false
	} else {
		question = &Question{
			Id:              uuid.New().String(),
			SessionId:       session.Id,
			QuestionContent: request.Content,
			AnswerCount:     0,
			MessageType:     "chat",
			HasImage:        false,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
	}

	answer := &Answer{
		Id:         uuid.New().String(),
		SessionId:  session.Id,
		QuestionId: question.Id,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	images, err := c.convertImageDatas(request.Images)
	if err != nil {
		return nil, err
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

func (c *Chat) chat(session *Session, question *Question, answer *Answer, images []olm.ImageData, isNew bool) {
	var answerImages []olm.ImageData
	defer dao.CreateAnswer(question, answer, images, answerImages, isNew)
	var keepAlive *olm.Duration
	if session.KeepAlive > 0 {
		keepAlive = &olm.Duration{
			Duration: session.KeepAlive,
		}
	}
	messages, err := dao.CombineHistoryMessages(session, !isNew)
	if err != nil {
		runtime.EventsEmit(app.ctx, answer.Id, nil, err)
	}
	messages = append(messages, olm.Message{
		Role:    "user",
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
