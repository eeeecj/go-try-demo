package cuszap

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"sync"
)

//InfoLogger 详细记录非错误信息
type InfoLogger interface {
	// Info 以给定的键/值对作为上下文记录一条非错误消息。
	//
	// msg 参数应该用于添加一些常量描述
	// 日志行。 然后可以使用键/值对添加额外的变量信息。
	// 键/值对应交替字符串 键和任意值。
	Info(msg string, fields ...Field)
	Infof(format string, v ...interface{})
	Infow(msg string, keyAndValues ...interface{})

	//	Enabled 测试是否能够正常使用
	Enabled() bool
}
type Logger interface {
	InfoLogger

	Debug(msg string, fields ...Field)
	Debugf(format string, v ...interface{})
	Debugw(msg string, keyAndValues ...interface{})

	Warn(msg string, fields ...Field)
	Warnf(format string, v ...interface{})
	Warnw(msg string, keyAndValues ...interface{})

	Error(msg string, fields ...Field)
	Errorf(format string, v ...interface{})
	Errorw(msg string, keyAndValues ...interface{})

	Panic(msg string, fields ...Field)
	Panicf(format string, v ...interface{})
	Panicw(msg string, keyAndValues ...interface{})

	Fatal(msg string, fields ...Field)
	Fatalf(format string, v ...interface{})
	Fatalw(msg string, keyAndValues ...interface{})

	//	V 返回一个特定级别的InfoLogger，值越大，越不重要
	V(level int) InfoLogger

	Write(p []byte) (n int, err error)

	//	WithValues 向一个logger添加键值对
	WithValues(keyAndValues ...interface{}) Logger

	//	WithName 向logger添加名称，他回忆前缀方式出现在logger name
	WithName(name string) Logger

	//	WithContext 返回一个上下文，将log作为value进行设置
	WithContext(ctx context.Context) context.Context

	//Flush 调用底层 Core 的 Sync 方法，刷新任何缓冲的
	//日志条目。 应用程序应注意在退出之前调用 Sync。
	Flush()
}

var _ Logger = &zapLogger{}

//一个无法使用的InfoLogger
type noopInfoLogger struct{}

func (n *noopInfoLogger) Enabled() bool {
	return false
}

func (n *noopInfoLogger) Info(_ string, _ ...Field)        {}
func (n *noopInfoLogger) Infof(_ string, _ ...interface{}) {}
func (n *noopInfoLogger) Infow(_ string, _ ...interface{}) {}

var disabledInfoLogger = &noopInfoLogger{}

// infoLogger is a logr.InfoLogger that uses Zap to log at a particular
// level.  The level has already been converted to a Zap level, which
// is to say that `logrLevel = -1*zapLevel`.
type infoLogger struct {
	level zapcore.Level
	log   *zap.Logger
}

func (l *infoLogger) Enabled() bool {
	return true
}
func (l *infoLogger) Info(msg string, fields ...Field) {
	if checkedEntry := l.log.Check(l.level, msg); checkedEntry != nil {
		checkedEntry.Write(fields...)
	}
}

func (l *infoLogger) Infof(format string, args ...interface{}) {
	if checkEntry := l.log.Check(l.level, fmt.Sprintf(format, args...)); checkEntry != nil {
		checkEntry.Write()
	}
}

func (l *infoLogger) Infow(msg string, keyAndValues ...interface{}) {
	fmt.Println(l.level)
	if checkEntry := l.log.Check(l.level, msg); checkEntry != nil {
		checkEntry.Write(handleFeilds(l.log, keyAndValues)...)
	}
}

// handleFields converts a bunch of arbitrary key-value pairs into Zap fields.  It takes
// additional pre-converted Zap fields, for use with automatically attached fields, like
// `error`.
func handleFeilds(l *zap.Logger, args []interface{}, additioanal ...zap.Field) []zap.Field {
	if len(args) == 0 {
		//fast-return if we have no suggared fields
		return additioanal
	}
	fields := make([]zap.Field, 0, len(args)/2+len(additioanal))
	for i := 0; i < len(args); {
		// 检查zap字段强类型
		if _, ok := args[i].(zap.Field); ok {
			//Any 接受一个键和一个任意值，并选择将它们表示为字段的最佳方式，仅在必要时才回退到基于反射的方法。
			l.DPanic("strongly-typed zap filed passed to log", zap.Any("zap field", args[i]))
			break
		}
		// make sure this isn't a mismatched key
		if i == len(args)-1 {
			l.DPanic("odd number of arguments passed as key-value pairs for logging", zap.Any("ignored key", args[i]))
			break
		}
		key, val := args[i], args[i+1]
		keystr, isString := key.(string)
		if !isString {
			//	if the key isn't a string,Panic and stop logging
			l.DPanic("non-string key argument passed to logging,, ignoring all later arguments", zap.Any("invalid key", key))
			break
		}
		fields = append(fields, zap.Any(keystr, val))
		i += 2
	}
	return append(fields, additioanal...)
}

// zapLogger is a Logger that uses Zap to log.
type zapLogger struct {
	zapLogger *zap.Logger
	infoLogger
}

var (
	std = New(NewOptions())
	mu  sync.Mutex
)

//Init logger with Options
func Init(opt *Options) {
	mu.Lock()
	defer mu.Unlock()
	std = New(opt)
}

//New create logger by opts
func New(opts *Options) *zapLogger {
	//判断选项是否为空
	if opts == nil {
		opts = NewOptions()
	}

	//定义level，如果能够从选项中获取opt，则使用optlevel，否则使用Infolevel
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	//序列化选项
	encodeLevel := zapcore.CapitalLevelEncoder

	//	设置本地输出颜色
	if opts.Format == consoleFormat && opts.EnableColor {
		encodeLevel = zapcore.CapitalColorLevelEncoder
	}
	encoderConfig := zapcore.EncoderConfig{
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
	}
	loggerCofig := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       opts.Development,
		DisableCaller:     opts.DisableCaller,
		DisableStacktrace: opts.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
			Hook:       nil,
		},
		Encoding:         opts.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      opts.OutputPaths,
		ErrorOutputPaths: opts.ErrorOutputPaths,
		InitialFields:    nil,
	}

	var err error
	l, err := loggerCofig.Build(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		panic(err)
	}
	logger := &zapLogger{
		zapLogger: l.Named(opts.Name),
		infoLogger: infoLogger{
			level: zap.InfoLevel,
			log:   l,
		},
	}
	zap.RedirectStdLog(l)
	return logger
}

func SugaredLogger() *zap.SugaredLogger {
	return std.zapLogger.Sugar()
}

// StdErrLogger returns logger of standard library which writes to supplied zap
// logger at error level.

func StdErrLogger() *log.Logger {
	if std == nil {
		return nil
	}
	if l, err := zap.NewStdLogAt(std.zapLogger, zapcore.ErrorLevel); err != nil {
		return l
	}
	return nil
}

// StdInfoLogger returns logger of standard library which writes to supplied zap
// logger at info level.
func StdInfoLogger() *log.Logger {
	if std == nil {
		return nil
	}
	if l, err := zap.NewStdLogAt(std.zapLogger, zapcore.InfoLevel); err != nil {
		return l
	}
	return nil
}

// V return a leveled InfoLogger.
func V(level int) InfoLogger {
	return std.V(level)
}

func (l *zapLogger) V(level int) InfoLogger {
	lvl := zapcore.Level(-1 * level)
	if l.zapLogger.Core().Enabled(lvl) {
		return &infoLogger{
			level: lvl,
			log:   l.zapLogger,
		}
	}
	return disabledInfoLogger
}

func (l *zapLogger) Write(p []byte) (int, error) {
	l.zapLogger.Info(string(p))
	return len(p), nil
}

// WithValues creates a child logger and adds adds Zap fields to it.
func WithValues(keyAndValues ...interface{}) Logger {
	return std.WithValues(keyAndValues...)
}

func (l *zapLogger) WithValues(keyAndValues ...interface{}) Logger {
	newLogger := l.zapLogger.With(handleFeilds(l.zapLogger, keyAndValues)...)
	return NewLogger(newLogger)
}

// WithName adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func WithName(s string) Logger {
	return std.WithName(s)
}
func (l *zapLogger) WithName(s string) Logger {
	newLogger := l.zapLogger.Named(s)
	return NewLogger(newLogger)
}

// Flush calls the underlying Core's Sync method, flushing any buffered
// log entries. Applications should take care to call Sync before exiting.
func Flush() {
	std.Flush()
}

func (l *zapLogger) Flush() {
	_ = l.zapLogger.Sync()
}

// NewLogger creates a new cuslog.Logger using the given Zap Logger to log.
func NewLogger(z *zap.Logger) Logger {
	return &zapLogger{
		zapLogger: z,
		infoLogger: infoLogger{
			log:   z,
			level: zap.InfoLevel,
		},
	}
}

// ZapLogger used for other log wrapper such as klog.
func ZapLogger() *zap.Logger {
	return std.zapLogger
}

// CheckIntLevel used for other log wrapper such as klog which return if logging a
// message at the specified level is enabled.
func CheckIntLevel(level int32) bool {
	var lvl zapcore.Level
	if level < 5 {
		lvl = zap.InfoLevel
	} else {
		lvl = zap.DebugLevel
	}
	checkEntry := std.zapLogger.Check(lvl, "")
	return checkEntry != nil
}

// Debug method output debug level log.
func Debug(msg string, field ...Field) {
	std.zapLogger.Debug(msg, field...)
}

func (l *zapLogger) Debug(msg string, field ...Field) {
	l.zapLogger.Debug(msg, field...)
}

// Debugf method output debug level log.
func Debugf(format string, v ...interface{}) {
	std.zapLogger.Sugar().Debugf(format, v...)
}

func (l *zapLogger) Debugf(format string, v ...interface{}) {
	l.zapLogger.Sugar().Debugf(format, v...)
}

func Debugw(msg string, keyAndValue ...interface{}) {
	std.zapLogger.Sugar().Debugw(msg, keyAndValue...)
}

func (l *zapLogger) Debugw(msg string, keyAndValue ...interface{}) {
	l.zapLogger.Sugar().Debugw(msg, keyAndValue...)
}

func Info(msg string, fields ...Field) {
	std.zapLogger.Info(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.zapLogger.Info(msg, fields...)
}

func Infof(format string, v ...interface{}) {
	std.zapLogger.Sugar().Infof(format, v...)
}
func (l *zapLogger) Infof(format string, v ...interface{}) {
	l.zapLogger.Sugar().Infof(format, v...)
}

func Infow(msg string, keyAndValue ...interface{}) {
	std.zapLogger.Sugar().Infow(msg, keyAndValue...)
}
func (l *zapLogger) Infow(msg string, keyAndValue ...interface{}) {
	l.zapLogger.Sugar().Infof(msg, keyAndValue...)
}

// Warn method output warning level log.
func Warn(msg string, fields ...Field) {
	std.zapLogger.Warn(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.zapLogger.Warn(msg, fields...)
}

// Warnf method output warning level log.
func Warnf(format string, v ...interface{}) {
	std.zapLogger.Sugar().Warnf(format, v...)
}

func (l *zapLogger) Warnf(format string, v ...interface{}) {
	l.zapLogger.Sugar().Warnf(format, v...)
}

// Warnw method output warning level log.
func Warnw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Warnw(msg, keysAndValues...)
}

func (l *zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Warnw(msg, keysAndValues...)
}

// Error method output error level log.
func Error(msg string, fields ...Field) {
	std.zapLogger.Error(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.zapLogger.Error(msg, fields...)
}

// Errorf method output error level log.
func Errorf(format string, v ...interface{}) {
	std.zapLogger.Sugar().Errorf(format, v...)
}

func (l *zapLogger) Errorf(format string, v ...interface{}) {
	l.zapLogger.Sugar().Errorf(format, v...)
}

// Errorw method output error level log.
func Errorw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Errorw(msg, keysAndValues...)
}

// Panic method output panic level log and shutdown application.
func Panic(msg string, fields ...Field) {
	std.zapLogger.Panic(msg, fields...)
}

func (l *zapLogger) Panic(msg string, fields ...Field) {
	l.zapLogger.Panic(msg, fields...)
}

// Panicf method output panic level log and shutdown application.
func Panicf(format string, v ...interface{}) {
	std.zapLogger.Sugar().Panicf(format, v...)
}

func (l *zapLogger) Panicf(format string, v ...interface{}) {
	l.zapLogger.Sugar().Panicf(format, v...)
}

// Panicw method output panic level log.
func Panicw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Panicw(msg, keysAndValues...)
}

func (l *zapLogger) Panicw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Panicw(msg, keysAndValues...)
}

// Fatal method output fatal level log.
func Fatal(msg string, fields ...Field) {
	std.zapLogger.Fatal(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.zapLogger.Fatal(msg, fields...)
}

// Fatalf method output fatal level log.
func Fatalf(format string, v ...interface{}) {
	std.zapLogger.Sugar().Fatalf(format, v...)
}

func (l *zapLogger) Fatalf(format string, v ...interface{}) {
	l.zapLogger.Sugar().Fatalf(format, v...)
}

// Fatalw method output Fatalw level log.
func Fatalw(msg string, keysAndValues ...interface{}) {
	std.zapLogger.Sugar().Fatalw(msg, keysAndValues...)
}

func (l *zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Sugar().Fatalw(msg, keysAndValues...)
}

// L method output with specified context value.
func L(ctx context.Context) *zapLogger {
	return std.L(ctx)
}

func (l *zapLogger) L(ctx context.Context) *zapLogger {
	lg := l.clone()
	requestID, _ := ctx.Value(KeyRequestID).(string)
	username, _ := ctx.Value(KeyUserName).(string)
	lg.zapLogger = lg.zapLogger.With(zap.String(KeyRequestID, requestID), zap.String(KeyUserName, username))
	return lg
}

func (l *zapLogger) clone() *zapLogger {
	copy := *l
	return &copy
}
