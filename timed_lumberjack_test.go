package lumberjack

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTimedRotatingLogger(t *testing.T) {
	dir := makeTempDir("TestTimeWindowRotate", t)
	defer os.RemoveAll(dir)

	nameFn := func(name string, local bool) string {
		dir := filepath.Dir(name)
		filename := filepath.Base(name)
		now := currentTime()
		if !local {
			now = now.UTC()
		}

		timestamp := now.Format("20060102150304")
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
		return time.Parse("20060102150304", ext)
	}
	naming, err := NewBackupNaming(nameFn, timeFn)
	if err != nil {
		t.Fatalf("backup name and time parse mismatched error: %v", err)
	}

	filename := logFile(dir)
	tl := NewTimedRotatingLogger(&Logger{
		BackupNaming: naming,
		Filename:     filename,
		LocalTime:    true,
	}, WithCronSpec("*/5 * * * * ?"))
	defer tl.Close()
	b := []byte("boo!")
	n, err := tl.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	existsWithContent(filename, b, t)
	fileCount(dir, 1, t)

	// should not rotate
	b2 := []byte("foooooo!")
	n, err = tl.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)

	<-time.After(time.Second * 6)

	// should rotate
	b3 := []byte("b3ooo!")
	n, err = tl.Write(b3)
	isNil(err, t)
	equals(len(b3), n, t)

	// the old logfile should be moved aside and the main logfile should have
	// only the last write in it.
	existsWithContent(filename, b3, t)

	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var backupName string
	for _, f := range fs {
		if f.Name() != "foobar.log" {
			backupName = f.Name()
		}
	}
	// the backup file will use the current fake time and have the old contents.
	backupfile := filepath.Join(dir, backupName)
	existsWithContent(backupfile, []byte("boo!foooooo!"), t)
	fileCount(dir, 2, t)
}
