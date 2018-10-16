package logger

import (
	"log"
	"os"
)

// Log is a wrapper on the stdlib log pkg.
type Log struct {
	*log.Logger
}

// New returns an initialized Log with defaults setup.
func New() *Log {
	return &Log{
		log.New(os.Stderr, "", log.Lshortfile|log.Lmicroseconds|log.Ldate),
	}
}
