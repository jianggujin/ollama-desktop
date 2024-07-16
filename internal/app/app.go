package app

import (
	"context"
	"ollama-desktop/internal/config"
	"ollama-desktop/internal/log"
)

// App struct
type App struct {
	ctx context.Context
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Info().Ctx(ctx).Msg("Ollama Desktop startup...")
}

func (a *App) shutdown(ctx context.Context) {
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
