package cuslog

import (
	"io"
	"os"
)

const (
	FmtEmptySeparate = ""
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

func initOptions(opts ...Option) *options {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	if o.output == nil {
		o.output = os.Stderr
	}
	if o.formatter == nil {
		//默认formmater为TextFormatter
		o.formatter = &TextFormatter{}
	}
	return o
}

func WithOutput(output io.Writer) Option {
	return Option(func(options2 *options) {
		options2.output = output
	})
}

func WithLevel(level Level) Option {
	return Option(func(options2 *options) {
		options2.level = level
	})
}

func WithStdLevel(level Level) Option {
	return Option(func(options2 *options) {
		options2.stdLevel = level
	})
}

func WithFormatter(f Formatter) Option {
	return Option(func(options2 *options) {
		options2.formatter = f
	})
}

func WithDisableCaller(d bool) Option {
	return Option(func(options2 *options) {
		options2.disableCaller = d
	})
}
