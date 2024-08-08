package app

import (
	"context"
	"database/sql"
	dao2 "ollama-desktop/internal/dao"
	"time"
)

const (
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

func (d *Dao) db() *sql.DB {
	return d.dao.GetDb()
}

func (d *Dao) transaction(fn func(tx *sql.Tx) error) error {
	tx, err := d.dao.GetDb().Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}

type ConfigModel struct {
	ConfigKey   string    `json:"configKey"`
	ConfigValue string    `json:"configValue"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SessionModel struct {
	Id                  string    `json:"id"`
	SessionName         string    `json:"sessionName"`
	ModelName           string    `json:"modelName"`
	MessageHistoryCount int       `json:"messageHistoryCount"`
	KeepAlive           string    `json:"keepAlive,omitempty"`
	SystemMessage       string    `json:"systemMessage,omitempty"`
	Options             string    `json:"options,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type ChatMessageModel struct {
	Id                 string        `json:"id"`
	SessionId          string        `json:"sessionId"`
	QuestionContent    string        `json:"questionContent"`
	AnswerContent      string        `json:"answerContent"`
	TotalDuration      time.Duration `json:"totalDuration"`
	LoadDuration       time.Duration `json:"loadDuration"`
	PromptEvalCount    int           `json:"promptEvalCount"`
	PromptEvalDuration time.Duration `json:"promptEvalDuration"`
	EvalCount          int           `json:"evalCount"`
	EvalDuration       time.Duration `json:"evalDuration"`
	DoneReason         string        `json:"doneReason"`
	IsSuccess          bool          `json:"isSuccess"`
	CreatedAt          time.Time     `json:"createdAt"`
	UpdatedAt          time.Time     `json:"updatedAt"`
}
