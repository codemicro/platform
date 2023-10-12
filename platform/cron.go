package platform

import (
	"github.com/robfig/cron/v3"
	"log/slog"
)

var cronEngine *cron.Cron

func init() {
	cron.DefaultLogger = cronLogger{}
	cronEngine = cron.New()
}

func RegisterRecurringTask(spec string, f func() error) (cron.EntryID, error) {
	return cronEngine.AddFunc(spec, func() {
		if err := f(); err != nil {
			slog.Error("recurring task failure: %w", err)
		}
	})
}

func StartCron() {
	go cronEngine.Run()
	slog.Info("Cron engine alive!")
}

func GetCronEntry(id cron.EntryID) cron.Entry {
	return cronEngine.Entry(id)
}

type cronLogger struct{}

func (cronLogger) Info(msg string, keysAndValues ...any) {
	slog.Info("Cron: "+msg, keysAndValues...)
}

func (cronLogger) Error(err error, msg string, keysAndValues ...any) {
	slog.Error(msg, append(keysAndValues, []any{"error", err}))
}
