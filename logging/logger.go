package logging

/*
golang's `log` package sucks, so we wrap it.
*/

import (
	"fmt"
	"log"
	"runtime"
)

type Logger struct {
	delegate      *log.Logger
	prefix        string
	debugHeader   string
	infoHeader    string
	warningHeader string
	errorHeader   string
	fatalHeader   string
}

const (
	kDebug   = "DEBUG"
	kInfo    = "INFO"
	kWarning = "WARNING"
	kError   = "ERROR"
	kFatal   = "FATAL"
)

func NewWrappedLogger(delegate *log.Logger, name string) Logger {
	return Logger{
		delegate:      delegate,
		debugHeader:   fmt.Sprintf("[%v][%v]", name, kDebug),
		infoHeader:    fmt.Sprintf("[%v][%v]", name, kInfo),
		warningHeader: fmt.Sprintf("[%v][%v]", name, kWarning),
		errorHeader:   fmt.Sprintf("[%v][%v]", name, kError),
		fatalHeader:   fmt.Sprintf("[%v][%v]", name, kFatal),
	}
}

func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	return fmt.Sprintf("[%v:%v]", file, line)
}

func (l Logger) Debug(format string, v ...any) {
	l.delegate.Printf(l.debugHeader+getCallerInfo()+" "+format, v...)
}

func (l Logger) Info(format string, v ...any) {
	l.delegate.Printf(l.infoHeader+getCallerInfo()+" "+format, v...)
}

func (l Logger) Warning(format string, v ...any) {
	l.delegate.Printf(l.warningHeader+getCallerInfo()+" "+format, v...)
}

func (l Logger) Error(format string, v ...any) {
	l.delegate.Printf(l.errorHeader+getCallerInfo()+" "+format, v...)
}

func (l Logger) Fatal(format string, v ...any) {
	l.delegate.Printf(l.fatalHeader+getCallerInfo()+" "+format, v...)
}
