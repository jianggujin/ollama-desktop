package app

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/options"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"ollama-desktop/internal/config"
	"ollama-desktop/internal/job"
	"ollama-desktop/internal/log"
	"runtime"
	"strings"
)

var app App = App{}

type App struct {
	ctx context.Context
}

func (a *App) startup(ctx context.Context) {
	log.Info().Ctx(ctx).Msg("Ollama Desktop startup...")
	a.ctx = ctx
	dao.startup(ctx)
	job.GetSchedule().AddFunc("0/10 * * * * ?", ollama.Heartbeat)
}

func (a *App) domReady(ctx context.Context) {
	log.Info().Ctx(ctx).Msg("Ollama Desktop domReady...")
	ollama.Heartbeat()
}

func (a *App) shutdown(ctx context.Context) {
	log.Info().Msg("Ollama Desktop shutdown...")
	dao.shutdown()
	job.GetSchedule().Stop()
}

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	secondInstanceArgs := secondInstanceData.Args

	log.Debug().Str("Args", strings.Join(secondInstanceData.Args, ",")).Msg("user opened second instance")
	wailsruntime.WindowUnminimise(a.ctx)
	wailsruntime.Show(a.ctx)
	go wailsruntime.EventsEmit(a.ctx, "launchArgs", secondInstanceArgs)
}

func (a *App) AppInfo() map[string]string {
	return map[string]string{
		"Version":   config.BuildVersion,
		"BuildHash": config.BuildHash,
		"Platform":  runtime.GOOS,
		"Arch":      runtime.GOARCH,
	}
}
