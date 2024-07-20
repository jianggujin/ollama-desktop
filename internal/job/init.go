package job

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"ollama-desktop/internal/log"
	"sync"
)

var _once sync.Once
var schedule *cron.Cron

type JobLogger struct {
}

func (l *JobLogger) Info(msg string, keysAndValues ...interface{}) {
	event := log.Info()
	for i := 0; i < len(keysAndValues); i += 2 {
		event.Any(fmt.Sprintf("%v", keysAndValues[i]), keysAndValues[i+1])
	}
	event.Msg(msg)
}

func (l *JobLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	event := log.Error()
	for i := 0; i < len(keysAndValues); i += 2 {
		event.Any(fmt.Sprintf("%v", keysAndValues[i]), keysAndValues[i+1])
	}
	event.Msg(msg)
}

func GetSchedule() *cron.Cron {
	_once.Do(func() {
		logger := &JobLogger{}
		schedule = cron.New(cron.WithLogger(logger), cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(logger), cron.Recover(logger)))
		schedule.Start()
	})
	return schedule
}
