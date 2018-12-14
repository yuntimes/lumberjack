package lumberjack_test

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuntimes/lumberjack"
)

// To use lumberjack with the standard library's log package, just pass it into
// the SetOutput function when your application starts.
func Example() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "/var/log/myapp/foo.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   // days
		Compress:   true, // disabled by default
	})
}

func ExampleBackupNaming() {
	nameFn := func(name string, local bool) string {
		dir := filepath.Dir(name)
		filename := filepath.Base(name)
		t := time.Now()
		if !local {
			t = t.UTC()
		}

		timestamp := t.Format("2006010215")
		return filepath.Join(dir, fmt.Sprintf("%s.%s", filename, timestamp))
	}
	timeFn := func(filename, prefix, ext string) (time.Time, error) {
		if strings.HasSuffix(prefix, "-") {
			prefix = prefix[:len(prefix)-1]
		}
		if !strings.HasPrefix(filename, prefix) {
			return time.Time{}, errors.New("mismatched prefix")
		}
		if !strings.HasSuffix(filename, ext) {
			return time.Time{}, errors.New("mismatched extension")
		}
		return time.Parse("2006010215", ext)
	}
	naming, err := lumberjack.NewBackupNaming(nameFn, timeFn)
	if err != nil {
		log.Fatal("backup name and time parse mismatched")
	}
	log.SetOutput(&lumberjack.Logger{
		BackupNaming: naming,
		Filename:     "/var/log/myapp/foo.log",
		MaxSize:      500, // megabytes
		MaxBackups:   3,
		MaxAge:       28,   // days
		Compress:     true, // disabled by default
	})
}
