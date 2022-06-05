package cuszap

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

const (
	flagLevel             = "log.level"
	flagDisableCaller     = "log.disable-caller"
	flagDisableStacktrace = "log.disable-stacktrace"
	flagFormat            = "log.format"
	flagEnableColor       = "log.enable-color"
	flagOutputPath        = "log.output-paths"
	flagErrorOutputPaths  = "log.error-output-paths"
	flagDevelopment       = "log.development"
	flagName              = "log.name"

	consoleFormat = "console"
	jsonFormat    = "json"
)

//option contains configuration items related to log
type Options struct {
	OutputPaths       []string `json:"output-paths" mapstructure:"output-paths"`
	ErrorOutputPaths  []string `json:"error-output-paths" mapstructure:"error-output-paths"`
	Level             string   `json:"level" mapstructure:"level"'`
	Format            string   `json:"formart" mapstructure:"format"'`
	DisableCaller     bool     `json:"disable-caller" mapstructure:"disable-caller"`
	DisableStacktrace bool     `json:"disable-stacktrace" mapstructure:"disable-stacktrace"'`
	EnableColor       bool     `json:"enable-color" mapstructure:"enable-color"`
	Development       bool     `json:"development" mapstructure:"development"`
	Name              string   `json:"name" mapstructure:"name"`
}

// NewOptions 创建默认配置
func NewOptions() *Options {
	return &Options{
		Level:             zapcore.InfoLevel.String(),
		DisableCaller:     false,
		DisableStacktrace: false,
		Format:            consoleFormat,
		EnableColor:       false,
		Development:       false,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stdout"},
	}
}

//Validate 验证配置是否有效
func (o *Options) Validate() []error {
	var errs []error

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(o.Level)); err != nil {
		errs = append(errs, err)
	}

	format := strings.ToLower(o.Format)
	if format != consoleFormat && format != jsonFormat {
		errs = append(errs, fmt.Errorf("not a valid log format:%q", o.Format))
	}
	return errs
}

// AddFlags 添加命令行标志
func (o *Options) AddFlag(fs *pflag.FlagSet) {
	fs.StringVar(&o.Level, flagLevel, o.Level, "Minimum log output level.")
	fs.BoolVar(&o.DisableCaller, flagDisableStacktrace, o.DisableCaller, "Disable output of caller information in the log. ")
	fs.BoolVar(&o.DisableStacktrace, flagDisableStacktrace, o.DisableStacktrace, "Disable the log to record a stack trace for all messages at or above panic level ")
	fs.StringVar(&o.Format, flagFormat, o.Format, "Log output Format,support plain or json")
	fs.BoolVar(&o.EnableColor, flagEnableColor, o.EnableColor, "Enable output ansi colors in plain format logs.")
	fs.StringSliceVar(&o.OutputPaths, flagOutputPath, o.OutputPaths, "Output paths of log.")
	fs.StringSliceVar(&o.ErrorOutputPaths, flagErrorOutputPaths, o.ErrorOutputPaths, "Error output paths of log.")
	fs.BoolVar(&o.Development, flagDevelopment, o.Development, "Development puts the logger in development mode, which changes "+
		"the behavior of DPanicLevel and takes stacktraces more liberally.")
	fs.StringVar(&o.Name, flagName, o.Name, "The name of the logger.")
}

func (o *Options) String() string {
	data, _ := json.Marshal(o)
	return string(data)
}

//Build constructs a zap logger from the option
func (o *Options) Build() error {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(o.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	encodeLevel := zapcore.CapitalLevelEncoder
	if o.Format == consoleFormat && o.EnableColor {
		encodeLevel = zapcore.CapitalColorLevelEncoder
	}
	zp := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       o.Development,
		DisableCaller:     o.DisableCaller,
		DisableStacktrace: o.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
			Hook:       nil,
		},
		Encoding: o.Format,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:          "message",
			LevelKey:            "level",
			TimeKey:             "timestamp",
			NameKey:             "logger",
			CallerKey:           "caller",
			StacktraceKey:       "stacktrace",
			LineEnding:          zapcore.DefaultLineEnding,
			EncodeLevel:         encodeLevel,
			EncodeTime:          timerEncoder,
			EncodeDuration:      milliSecondsDurationEncoder,
			EncodeCaller:        zapcore.ShortCallerEncoder,
			EncodeName:          zapcore.FullNameEncoder,
			NewReflectedEncoder: nil,
			ConsoleSeparator:    "",
		},
		OutputPaths:      o.OutputPaths,
		ErrorOutputPaths: o.ErrorOutputPaths,
		InitialFields:    nil,
	}
	//高于panic等级的记录堆栈信息
	logger, err := zp.Build(zap.AddStacktrace(zap.PanicLevel))
	if err != nil {
		return err
	}
	//替换全局的logger为自定义的logger
	zap.RedirectStdLog(logger.Named(o.Name))
	zap.ReplaceGlobals(logger)
	return nil
}
