package test

import (
	"fmt"
	"testing"

	"go.uber.org/zap"
	"knative.dev/pkg/test/logging"
)

type Logger interface {
	// Log formats its arguments using default formatting, analogous to Println,
	// and records the text in the error log.
	Log(args ...interface{})

	// Logf formats its arguments according to the format, analogous to Printf, and
	// records the text in the error log. A final newline is added if not provided.
	Logf(format string, args ...interface{})

	// Fatal is equivalent to Log followed by panic.
	Fatal(args ...interface{})

	// Fatalf is equivalent to Logf followed by panic.
	Fatalf(format string, args ...interface{})

	finalize()
}

func GetLogger(t *testing.T) Logger {
	if zapLoggerInstance == nil {
		logging.InitializeLogger()
		zapLoggerInstance = &zapLogger{
			log: zap.S(),
		}
		zapLoggerInstance.Log("TestLogger initialized")
	}
	if t != nil && zapLoggerInstance.tlog == nil {
		tlog, finalizeFunc := logging.NewTLogger(t)
		zapLoggerInstance.tlog = tlog
		zapLoggerInstance.finalizeFunc = finalizeFunc
		zapLoggerInstance.Log("TestLogger enhanced with testing.T")
	}
	return zapLoggerInstance
}

var zapLoggerInstance *zapLogger

type zapLogger struct {
	log      *zap.SugaredLogger
	tlog     *logging.TLogger
	finalizeFunc func()
}

func (z *zapLogger) finalize() {
	if z.finalizeFunc != nil {
		z.log.Debug("Finalizing TestLogger...")
		z.finalizeFunc()
	}
}

func (z *zapLogger) Log(args ...interface{}) {
	z.log.Info(args)
}

func (z *zapLogger) Logf(format string, args ...interface{}) {
	z.log.Infof(format, args)
}

func (z *zapLogger) Fatal(args ...interface{}) {
	if z.tlog != nil {
		z.tlog.Fatal(args)
	} else {
		z.log.Fatal(args)
	}
}

func (z *zapLogger) Fatalf(format string, args ...interface{}) {
	if z.tlog != nil {
		z.tlog.Fatal(fmt.Errorf(format, args))
	} else {
		z.log.Fatalf(format, args)
	}
}
