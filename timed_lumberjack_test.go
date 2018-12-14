package lumberjack

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTimeWindowRotate(t *testing.T) {
	fixedTime1 := func() time.Time {
		return time.Unix(1544781600, 0)
	}
	currentTime = fixedTime1

	dir := makeTempDir("TestTimeWindowRotate", t)
	defer os.RemoveAll(dir)

	nameFn := func(name string, local bool) string {
		dir := filepath.Dir(name)
		filename := filepath.Base(name)
		t := currentTime()
		if !local {
			t = t.UTC()
		}

		timestamp := t.Format("20060102150304")
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
	}, WithWhen("S"), WithInterval(5))
	t.Log(tl.lastWindow)
	defer tl.Close()
	b := []byte("boo!")
	n, err := tl.Write(b)
	isNil(err, t)
	equals(len(b), n, t)

	existsWithContent(filename, b, t)
	fileCount(dir, 1, t)

	fixedTime2 := func() time.Time {
		return time.Unix(1544781602, 0)
	}
	currentTime = fixedTime2

	// should not rotate
	b2 := []byte("foooooo!")
	n, err = tl.Write(b2)
	isNil(err, t)
	equals(len(b2), n, t)
	t.Log(tl.lastWindow)

	fixedTime3 := func() time.Time {
		return time.Unix(1544781608, 0)
	}
	currentTime = fixedTime3

	// should rotate
	b3 := []byte("b3ooo!")
	n, err = tl.Write(b3)
	isNil(err, t)
	equals(len(b3), n, t)
	t.Log(tl.lastWindow)

	//<-time.After(time.Millisecond * 20)

	// the old logfile should be moved aside and the main logfile should have
	// only the last write in it.
	existsWithContent(filename, b3, t)

	// the backup file will use the current fake time and have the old contents.
	backupfile := filepath.Join(dir, "foobar.log."+fixedTime3().Format("20060102150304"))
	existsWithContent(backupfile, []byte("boo!foooooo!"), t)
	fileCount(dir, 2, t)
}
