package lumberjack

import (
	"fmt"

	"github.com/robfig/cron"
)

// Options defines the options for TimedRotatingLogger
type Options struct {
	CronSpec string
}

// Option defines the option func
type Option func(*Options)

// WithCronSpec set cron expression, refer to `https://www.godoc.org/github.com/robfig/cron`
func WithCronSpec(spec string) Option {
	return func(opts *Options) {
		opts.CronSpec = spec
	}
}

// NewTimedRotatingLogger creates new TimedRotatingLogger
func NewTimedRotatingLogger(logger *Logger, opts ...Option) *TimedRotatingLogger {
	options := &Options{
		CronSpec: "0 0 0 * * *", // midnight
	}
	for _, o := range opts {
		o(options)
	}

	tl := &TimedRotatingLogger{
		Logger:   logger,
		cronSpec: options.CronSpec,
		cron:     cron.New(),
	}
	err := tl.cron.AddFunc(tl.cronSpec, func() { tl.Rotate() })
	if err != nil {
		panic(fmt.Errorf("bad cron expreesion: %v", err))
	}
	tl.cron.Start()

	return tl
}

// TimedRotatingLogger rotates log according to cron expression
type TimedRotatingLogger struct {
	*Logger

	cronSpec string
	cron     *cron.Cron
}

// Close close the logger and stop the scheduler
func (l *TimedRotatingLogger) Close() error {
	l.cron.Stop()
	return l.Logger.Close()
}
