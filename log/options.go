package log

import (
	"io"
	"os"
)

type Level uint8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

var LevelNameMapping = map[Level]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	PanicLevel: "PANIC",
	FatalLevel: "FATAL",
}

type options struct {
	output        io.Writer
	level         Level
	stdLevel      Level
	formatter     Formatter
	disableCaller bool
}

type Option func(options2 *options)

func initOptions(opts ...Option) Option {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	if o.output == nil {
		o.output = os.Stderr
	}
	if o.formatter == nil {
		o.formatter = &Fo
	}
}
