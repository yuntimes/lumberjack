package lumberjack

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	whens = map[string]func(inteval uint) string{
		"S":        func(interval uint) string { return timeWindow(interval, 1) },
		"M":        func(interval uint) string { return timeWindow(interval, 60) },
		"H":        func(interval uint) string { return timeWindow(interval, 3600) },
		"D":        func(interval uint) string { return dayWindow(interval) },
		"W0":       func(interval uint) string { return weekdayWindow(time.Monday) },
		"W1":       func(interval uint) string { return weekdayWindow(time.Tuesday) },
		"W2":       func(interval uint) string { return weekdayWindow(time.Wednesday) },
		"W3":       func(interval uint) string { return weekdayWindow(time.Thursday) },
		"W4":       func(interval uint) string { return weekdayWindow(time.Friday) },
		"W5":       func(interval uint) string { return weekdayWindow(time.Saturday) },
		"W6":       func(interval uint) string { return weekdayWindow(time.Sunday) },
		"midnight": func(interval uint) string { return time.Now().Format("2006-01-02") },
	}
)

func timeWindow(interval, seconds uint) string {
	ts := currentTime().Unix()
	return strconv.FormatInt(ts-ts%int64(interval*seconds), 10)
}

func dayWindow(interval uint) string {
	yearday := currentTime().YearDay()
	return strconv.Itoa(yearday - yearday%int(interval))
}

// weekdayWindow return the window by weekday
func weekdayWindow(wd time.Weekday) string {
	now := currentTime()
	year, week := now.ISOWeek()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7 // use 7 because time.Sunday == 0
	}
	if weekday < wd {
		week--
	}
	return fmt.Sprintf("%d-%d-%d", year, week, wd)
}

// Options defines the options for TimedRotatingLogger
type Options struct {
	When     string
	Interval uint
}

// Option defines the option func
type Option func(*Options)

// WithWhen set the when attr
// - "S": Seconds
// - "M": Minutes
// - "H": Hours
// - "D": Days
// - "W0" - "W6": Weekday (0=Monday)
// - "midnight": Roll over at midnight
func WithWhen(when string) Option {
	return func(opts *Options) {
		opts.When = when
	}
}

// WithInterval set the Interval attr
func WithInterval(interval uint) Option {
	return func(opts *Options) {
		opts.Interval = interval
	}
}

// NewTimedRotatingLogger creates new TimedRotatingLogger
func NewTimedRotatingLogger(logger *Logger, opts ...Option) *TimedRotatingLogger {
	options := &Options{
		When:     "D",
		Interval: 1,
	}
	for _, o := range opts {
		o(options)
	}
	when, ok := whens[options.When]
	if !ok {
		panic(fmt.Errorf("invalid when: %s", options.When))
	}
	if options.Interval == 0 {
		panic("invalid interval")
	}
	lastWindow := when(options.Interval)
	return &TimedRotatingLogger{
		Logger:     logger,
		When:       options.When,
		Interval:   options.Interval,
		lastWindow: lastWindow,
	}
}

// TimedRotatingLogger is wrapper on Logger with time-based rotation
//
// Usage:
//
// nameFn := func(name string, local bool) string {
// 	   dir := filepath.Dir(name)
// 	   filename := filepath.Base(name)
// 	   t := time.Now()
// 	   if !local {
// 	   	   t = t.UTC()
// 	   }
//
// 	   timestamp := t.Format("2006010215")
// 	   return filepath.Join(dir, fmt.Sprintf("%s.%s", filename, timestamp))
// }
// timeFn := func(filename, prefix, ext string) (time.Time, error) {
// 	   if strings.HasSuffix(prefix, "-") {
// 	   	   prefix = prefix[:len(prefix)-1]
// 	   }
// 	   if !strings.HasPrefix(filename, prefix) {
// 	   	   return time.Time{}, errors.New("mismatched prefix")
// 	   }
// 	   if !strings.HasSuffix(filename, ext) {
// 	   	   return time.Time{}, errors.New("mismatched extension")
// 	   }
// 	   return time.Parse("2006010215", ext)
// }
//
// naming, err := NewBackupNaming(nameFn, timeFn)
// if err != nil {
// 	   log.Fatal(err)
// }
//
// l := &Logger{
// 	   BackupNaming: naming,
// 	   Filename:     "app.log",
// 	   MaxSize:      10,
// 	   LocalTime:    true,
// }
// tl := NewTimedRotatingLogger(logger, WithWhen("H"))
// tl.Write("xxx")
type TimedRotatingLogger struct {
	*Logger
	When     string
	Interval uint

	lastWindow     string
	lastWindowLock sync.Mutex
}

func (l *TimedRotatingLogger) Write(p []byte) (n int, err error) {
	if when, ok := whens[l.When]; ok {
		lastWindow := when(l.Interval) // performance bottleneck?
		if lastWindow != l.lastWindow {
			l.lastWindowLock.Lock()
			l.lastWindow = lastWindow
			l.lastWindowLock.Unlock()
			l.Logger.Rotate()
		}
	}
	return l.Logger.Write(p)
}
