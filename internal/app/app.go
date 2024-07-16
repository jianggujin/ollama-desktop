package app

import (
	"context"
	"ollama-desktop/internal/config"
	dao2 "ollama-desktop/internal/dao"
	"ollama-desktop/internal/log"
)

// App struct
type App struct {
	dao *dao2.DbDao
	ctx context.Context
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.dao = &dao2.DbDao{}
	a.dao.Startup(ctx)
	log.Info().Ctx(ctx).Msg("Ollama Desktop startup...")
}

func (a *App) shutdown(ctx context.Context) {
	a.dao.Shutdown()
	log.Info().Msg("Ollama Desktop shutdown...")
}

func (a *App) AppVersion() string {
	return config.BuildVersion
}

func (a *App) AppBuildTime() string {
	return config.BuildTime
}

func (a *App) OllamaHost() string {
	return config.Config.Ollama.Host.String()
}
