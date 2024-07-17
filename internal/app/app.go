package app

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"ollama-desktop/internal/config"
	dao2 "ollama-desktop/internal/dao"
	"ollama-desktop/internal/log"
)

// App struct
type App struct {
	dao       *dao2.DbDao
	ctx       context.Context
	forceQuit bool
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

func (a *App) beforeClose(ctx context.Context) bool {
	if a.forceQuit {
		return false
	}
	log.Info().Msg("Ollama Desktop beforeClose...")
	runtime.EventsEmit(ctx, "beforeClose")
	return true
}

func (a *App) AppVersion() string {
	return config.BuildVersion
}

func (a *App) AppBuildTime() string {
	return config.BuildTime
}

func (a *App) Quit() {
	a.forceQuit = true
	runtime.Quit(a.ctx)
}

func (a *App) OllamaHost() string {
	return config.Config.Ollama.Host.String()
}
