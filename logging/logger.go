package logging

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mycelia/globals"
	"mycelia/str"
)

// -------Helpers/Formatters----------------------------------------------------

func getToday() string {
	return time.Now().Format(globals.DateLayout)
}

func getTodaysLogFileName() string {
	filename := fmt.Sprintf(
		"mycelia-log-%s.log", getToday(),
	)
	return filename
}

func getTodaysLogFilePath() string {
	return filepath.Join(globals.LogDirectory, getTodaysLogFileName())
}

func formatAction(s, uid string) string {
	now := time.Now().Format(globals.TimeLayout)
	out := fmt.Sprintf("%s: [%s][ACTION] - %s", now, uid, s)
	return out
}

func formatWarning(s, uid string) string {
	now := time.Now().Format(globals.TimeLayout)
	out := fmt.Sprintf("%s: [%s][WARNING] - %s", now, uid, s)
	return out
}

func formatError(s, uid string) string {
	now := time.Now().Format(globals.TimeLayout)
	out := fmt.Sprintf("%s: [%s][ERROR] - %s", now, uid, s)
	return out
}

// Writes out the given s when globals.Verbosity is set to or greater than the
// v level.
// Calls the corresponding write method with the matching prefix.
func formatByVerbosity(lm messageLog) (string, error) {
	if len(lm.msg) == 0 || lm.msg[len(lm.msg)-1] != '\n' {
		lm.msg += "\n"
	}

	if lm.verbosity == globals.VERB_ACT && globals.Verbosity >= lm.verbosity {
		return formatAction(lm.msg, lm.uid), nil
	}
	if lm.verbosity == globals.VERB_WRN && globals.Verbosity >= lm.verbosity {
		return formatWarning(lm.msg, lm.uid), nil
	}
	if lm.verbosity == globals.VERB_ERR && globals.Verbosity >= lm.verbosity {
		return formatError(lm.msg, lm.uid), nil
	}

	return "", fmt.Errorf("Verbosity level not active.")
}

// -------Logger----------------------------------------------------------------

type messageLog struct {
	uid       string
	msg       string
	verbosity int
}

// A Logger is a struct that writes to a dated log file.
// Use logging.GlobalLogger instead of making a new one.
type Logger struct {
	file   *os.File
	in     chan *messageLog
	date   string
	writer *bufio.Writer
}

func NewLogger() *Logger {
	if err := os.MkdirAll(globals.LogDirectory, 0755); err != nil {
		return nil
	}

	file, err := os.OpenFile(
		getTodaysLogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644,
	)
	if err != nil {
		return nil
	}

	l := &Logger{
		file:   file,
		date:   getToday(),
		in:     make(chan *messageLog, 1024),
		writer: bufio.NewWriter(file),
	}
	l.Start()

	return l
}

func (l *Logger) Start() { go l.loop() }
func (l *Logger) Stop()  { close(l.in) }

// Rotates the tracked log file to the new dated log file.
func (l *Logger) rotate() {
	_ = (l.writer).Flush()
	_ = l.file.Close()

	f, err := os.OpenFile(
		getTodaysLogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644,
	)
	if err != nil {
		str.ErrorPrint("Could not create next day log file.")
		return
	}
	l.file = f
	l.date = getToday()
	l.writer = bufio.NewWriter(l.file)
}

func (l *Logger) loop() {
	defer func() {
		_ = l.file.Close()
	}()

	writer := bufio.NewWriter(l.file)
	flush := func() {
		_ = writer.Flush()
	}
	defer flush()

	for ml := range l.in {
		if ml == nil {
			continue
		}

		if globals.LogOutput == globals.LOG_TO_FILE {
			msg, err := formatByVerbosity(*ml)
			if err != nil {
				continue
			}
			_, err = writer.WriteString(msg)
			if err != nil {
				str.ErrorPrint("Could not write to log buffer.")
			}
			flush()
		} else if globals.LogOutput == globals.LOG_TO_CONSOLE {
			str.PrintByVerbosity(ml.msg, ml.verbosity)
		}

		if getToday() != l.date {
			l.rotate()
		}
	}
}

// -------Global Singleton Logger-----------------------------------------------

var GlobalLogger = NewLogger()
