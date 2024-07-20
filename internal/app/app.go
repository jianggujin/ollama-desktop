package app

import (
	"context"
	"ollama-desktop/internal/config"
	dao2 "ollama-desktop/internal/dao"
	"ollama-desktop/internal/job"
	"ollama-desktop/internal/log"
	"runtime"
)

// App struct
type App struct {
	dao *dao2.DbDao
	ctx context.Context
}

func (a *App) startup(ctx context.Context) {
	log.Info().Ctx(ctx).Msg("Ollama Desktop startup...")
	a.ctx = ctx
	a.dao = &dao2.DbDao{}
	a.dao.Startup(ctx)
	job.GetSchedule().AddFunc("0/10 * * * * ?", a.OllamaHeartbeat)
}

func (a *App) domReady(ctx context.Context) {
	log.Info().Ctx(ctx).Msg("Ollama Desktop domReady...")
	a.OllamaHeartbeat()
}

func (a *App) shutdown(ctx context.Context) {
	log.Info().Msg("Ollama Desktop shutdown...")
	a.dao.Shutdown()
	job.GetSchedule().Stop()
}

func (a *App) AppInfo() map[string]string {
	return map[string]string{
		"Version":   config.BuildVersion,
		"BuildTime": config.BuildTime,
		"Platform":  runtime.GOOS,
		"Arch":      runtime.GOARCH,
	}
}
