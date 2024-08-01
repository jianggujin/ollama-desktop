package app

import (
	"context"
	dao2 "ollama-desktop/internal/dao"
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
