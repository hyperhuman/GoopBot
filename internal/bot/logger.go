package logger

import (
	"log"
	"os"
)

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

type loggerImpl struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
	debugLogger *log.Logger
}

func NewLogger(level string) Logger {
	return &loggerImpl{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime),
		warnLogger:  log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime),
	}
}
