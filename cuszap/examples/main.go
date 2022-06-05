package main

import (
	cuszap "awesomeProject"
	"context"
	"flag"
)

type lvl int8

var (
	h bool

	level  int
	format string
)

func main() {
	flag.BoolVar(&h, "h", false, "Print this help.")
	flag.IntVar(&level, "l", 0, "Log level.")
	flag.StringVar(&format, "f", "console", "log output format.")

	flag.Parse()

	if h {
		flag.Usage()

		return
	}

	// logger配置
	opts := &cuszap.Options{
		Level:            "debug",
		Format:           "console",
		EnableColor:      true,
		DisableCaller:    true,
		OutputPaths:      []string{"test.log", "stdout"},
		ErrorOutputPaths: []string{"error.log"},
	}
	// 初始化全局logger
	cuszap.Init(opts)
	defer cuszap.Flush()
	// Debug、Info(with field)、Warnf、Errorw使用
	cuszap.Debug("This is a debug message")
	cuszap.Info("This is a info message", cuszap.Int32("int_key", 10))
	cuszap.Warnf("This is a formatted %s message", "warn")
	cuszap.Errorw("Message printed with Errorw", "X-Request-ID", "fbf54504-64da-4088-9b86-67824a7fb508")

	// WithValues使用
	lv := cuszap.WithValues("X-Request-ID", "7a7b9f24-4cae-4b2a-9464-69088b45b904")
	lv.Infow("Info message printed with [WithValues] logger")
	lv.Infow("Debug message printed with [WithValues] logger")

	// Context使用
	ctx := lv.WithContext(context.Background())
	lc := cuszap.FromContext(ctx)
	lc.Info("Message printed with [WithContext] logger")

	ln := lv.WithName("test")
	ln.Info("Message printed with [WithName] logger")

	// V level使用
	cuszap.V(-1).Info("This is a V level message")
	cuszap.V(-5).Infow("This is a V level message with fields", "X-Request-ID", "7a7b9f24-4cae-4b2a-9464-69088b45b904")
}
