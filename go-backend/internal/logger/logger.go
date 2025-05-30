package logger

import (
	"io"
	"log"
	"os"

	"github.com/coreos/go-systemd/v22/journal"
)

var (
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

type journalWriter struct {
	priority journal.Priority
}

func (j journalWriter) Write(p []byte) (n int, err error) {
	_ = journal.Send(string(p), j.priority, nil)
	return len(p), nil
}

func Init(mode string, verbose bool) {
	flags := log.Ldate | log.Ltime

	useJournal := mode == "production" && journal.Enabled()

	switch {
	case useJournal && verbose:
		// Production + VERBOSE=true
		Debug = log.New(journalWriter{journal.PriDebug}, "[DEBUG] ", flags)
		Info = log.New(journalWriter{journal.PriInfo}, "", 0)
		Warning = log.New(journalWriter{journal.PriWarning}, "", 0)
		Error = log.New(journalWriter{journal.PriErr}, "", 0)

	case useJournal && !verbose:
		// Production + VERBOSE=false
		Debug = log.New(io.Discard, "", 0)
		Info = log.New(journalWriter{journal.PriInfo}, "", 0)
		Warning = log.New(journalWriter{journal.PriWarning}, "", 0)
		Error = log.New(journalWriter{journal.PriErr}, "", 0)

	case !useJournal && verbose:
		// Development + VERBOSE=true
		output := os.Stdout
		Debug = log.New(output, "[DEBUG] ", flags|log.Lshortfile)
		Info = log.New(output, "[INFO] ", flags)
		Warning = log.New(output, "[WARN] ", flags)
		Error = log.New(output, "[ERROR] ", flags)

	default:
		// Development + VERBOSE=false
		output := os.Stdout
		Debug = log.New(io.Discard, "", 0)
		Info = log.New(output, "[INFO] ", flags)
		Warning = log.New(output, "[WARN] ", flags)
		Error = log.New(output, "[ERROR] ", flags)
	}
}
