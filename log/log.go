package log

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"os"
)

// The global log used throughout MIDA (not to be confused with individual logs for each crawl)
var Log = logrus.New()

func InitGlobalLogger(logfile string) {
	fileFormatter := new(logrus.TextFormatter)
	fileFormatter.FullTimestamp = true
	fileFormatter.DisableColors = true

	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   logfile,
		MaxSize:    50, //megabytes
		MaxBackups: 3,
		MaxAge:     180, //days
		Level:      logrus.InfoLevel,
		Formatter:  fileFormatter,
	})
	if err != nil {
		Log.Fatal("Logging initialization error: ", err)
	}

	consoleFormatter := new(logrus.TextFormatter)
	consoleFormatter.FullTimestamp = false
	consoleFormatter.ForceColors = true

	Log.SetOutput(os.Stdout)
	Log.SetFormatter(consoleFormatter)
	Log.AddHook(rotateFileHook)
}

// Helper function to setup logging using parameters from Cobra command
func ConfigureLogging(level int) error {
	switch level {
	case 0:
		Log.SetLevel(logrus.ErrorLevel)
	case 1:
		Log.SetLevel(logrus.WarnLevel)
	case 2:
		Log.SetLevel(logrus.InfoLevel)
	case 3:
		Log.SetLevel(logrus.DebugLevel)
		Log.SetReportCaller(true)
	default:
		return errors.New("invalid log level (Valid values: 0, 1, 2, 3)")
	}
	return nil
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Warn(args ...interface{}) {
	Log.Warn(args...)
}

func Error(args ...interface{}) {
	Log.Error(args...)
}

func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

func Debugf(format string, args ...interface{}) {
	Log.Debug(format, args)
}

func Infof(format string, args ...interface{}) {
	Log.Info(format, args)
}

func Warnf(format string, args ...interface{}) {
	Log.Warn(format, args)
}

func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

func WithURL(url string) *logrus.Entry {
	return Log.WithField("URL", url)
}
