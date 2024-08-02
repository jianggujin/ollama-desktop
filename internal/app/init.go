package app

import (
	"github.com/wailsapp/wails/v2"
	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"ollama-desktop/internal/config"
	"ollama-desktop/internal/log"
)

type logger struct {
}

func (l *logger) Print(message string) {
	log.Logger.Print(message)
}
func (l *logger) Trace(message string) {
	log.Trace().Msg(message)
}
func (l *logger) Debug(message string) {
	log.Debug().Msg(message)
}
func (l *logger) Info(message string) {
	log.Info().Msg(message)
}
func (l *logger) Warning(message string) {
	log.Warn().Msg(message)
}
func (l *logger) Error(message string) {
	log.Error().Msg(message)
}
func (l *logger) Fatal(message string) {
	log.Fatal().Msg(message)
}

func StartApp(server *assetserver.Options) error {
	level := config.Config.Logging.Level
	if level == "warn" {
		level = "warning"
	}
	ll, err := wailsLogger.StringToLogLevel(config.Config.Logging.Level)
	if err != nil {
		ll = wailsLogger.ERROR
	}
	var singleInstanceLock *options.SingleInstanceLock
	if config.Config.SingleInstance {
		singleInstanceLock = &options.SingleInstanceLock{
			UniqueId:               "d23a62a1-9f3d-4b9d-9c1e-8c1d0c63eafe",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		}
	}
	// Create application with options
	return wails.Run(&options.App{
		Title:                    "Ollama Desktop",
		Width:                    config.Config.Width,
		Height:                   config.Config.Height,
		MinWidth:                 config.Config.MinWidth,
		MinHeight:                config.Config.MinHeight,
		AlwaysOnTop:              config.Config.AlwaysOnTop,
		EnableDefaultContextMenu: config.Config.EnableDefaultContextMenu,
		//DisableResize: true,
		Frameless:   true,
		AssetServer: server,
		//BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:          app.startup,
		OnDomReady:         app.domReady,
		OnShutdown:         app.shutdown,
		SingleInstanceLock: singleInstanceLock,
		Bind: []interface{}{
			&app,
			&dao,
			&downloader,
			&ollama,
			&chat,
		},
		Logger:             &logger{},
		LogLevelProduction: ll,
	})
}
