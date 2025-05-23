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
	journal.Send(string(p), j.priority, nil)
	return len(p), nil
}

func Init(mode string, verbose bool) {
	flags := log.Ldate | log.Ltime
	if verbose {
		flags |= log.Lshortfile
	}

	var output io.Writer = os.Stdout

	useJournal := (mode == "production") && journal.Enabled()

	// Assign appropriate writers
	if useJournal {
		Debug = log.New(io.Discard, "[DEBUG] ", flags)
		Info = log.New(journalWriter{journal.PriInfo}, "", 0)
		Warning = log.New(journalWriter{journal.PriWarning}, "", 0)
		Error = log.New(journalWriter{journal.PriErr}, "", 0)
	} else {
		Debug = log.New(io.Discard, "[DEBUG] ", flags)
		if verbose {
			Debug.SetOutput(output)
		}
		Info = log.New(output, "[INFO] ", flags)
		Warning = log.New(output, "[WARN] ", flags)
		Error = log.New(output, "[ERROR] ", flags)
	}
}
