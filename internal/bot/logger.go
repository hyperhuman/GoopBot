package bot

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
func (l *loggerImpl) Infof(format string, args ...interface{}) {
	l.infoLogger.Printf(format, args...)
}
func (l *loggerImpl) Errorf(format string, args ...interface{}) {
	l.errorLogger.Printf(format, args...)
}
func (l *loggerImpl) Warnf(format string, args ...interface{}) {
	l.warnLogger.Printf(format, args...)
}
func (l *loggerImpl) Debugf(format string, args ...interface{}) {
	l.debugLogger.Printf(format, args...)
}
func (l *loggerImpl) SetLevel(level string) {
	switch level {
	case "info":
		l.infoLogger.SetOutput(os.Stdout)
		l.errorLogger.SetOutput(os.Stderr)
		l.warnLogger.SetOutput(os.Stdout)
		l.debugLogger.SetOutput(os.Stdout)
	case "error":
		l.infoLogger.SetOutput(os.Stderr)
		l.errorLogger.SetOutput(os.Stderr)
		l.warnLogger.SetOutput(os.Stderr)
		l.debugLogger.SetOutput(os.Stderr)
	case "warn":
		l.infoLogger.SetOutput(os.Stdout)
		l.errorLogger.SetOutput(os.Stderr)
		l.warnLogger.SetOutput(os.Stdout)
		l.debugLogger.SetOutput(os.Stdout)
	case "debug":
		l.infoLogger.SetOutput(os.Stdout)
		l.errorLogger.SetOutput(os.Stderr)
		l.warnLogger.SetOutput(os.Stdout)
		l.debugLogger.SetOutput(os.Stdout)
	default:
		log.Println("Unknown log level, defaulting to info")
	}
}
func (l *loggerImpl) GetLevel() string {
	if l.infoLogger.Writer() == os.Stdout {
		return "info"
	} else if l.errorLogger.Writer() == os.Stderr {
		return "error"
	} else if l.warnLogger.Writer() == os.Stdout {
		return "warn"
	} else if l.debugLogger.Writer() == os.Stdout {
		return "debug"
	}
	return "unknown"
}
func (l *loggerImpl) SetOutput(output *os.File) {
	l.infoLogger.SetOutput(output)
	l.errorLogger.SetOutput(output)
	l.warnLogger.SetOutput(output)
	l.debugLogger.SetOutput(output)
}
func (l *loggerImpl) GetOutput() *os.File {
	if l.infoLogger.Writer() == os.Stdout {
		return os.Stdout
	} else if l.errorLogger.Writer() == os.Stderr {
		return os.Stderr
	} else if l.warnLogger.Writer() == os.Stdout {
		return os.Stdout
	} else if l.debugLogger.Writer() == os.Stdout {
		return os.Stdout
	}
	return nil
}
func (l *loggerImpl) Close() error {
	if l.infoLogger.Writer() != nil {
		if closer, ok := l.infoLogger.Writer().(interface{ Close() error }); ok {
			return closer.Close()
		}
	}
	if l.errorLogger.Writer() != nil {
		if closer, ok := l.errorLogger.Writer().(interface{ Close() error }); ok {
			return closer.Close()
		}
	}
	if l.warnLogger.Writer() != nil {
		if closer, ok := l.warnLogger.Writer().(interface{ Close() error }); ok {
			return closer.Close()
		}
	}
	if l.debugLogger.Writer() != nil {
		if closer, ok := l.debugLogger.Writer().(interface{ Close() error }); ok {
			return closer.Close()
		}
	}
	return nil
}
