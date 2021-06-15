package klog

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	LevelInfo uint = iota
	LevelError
	LevelWarning
	LevelDebug
)

const (
	PrefixInfo    = "INF: "
	PrefixError   = "ERR: "
	PrefixWarning = "WRN: "
	PrefixDebug   = "DBG: "
	PrefixFatal   = "ERR: "
	PrefixPanic   = "PNC: "
)

type Logger struct {
	Level  uint
	Output io.Writer
}

var DefaultLogger = &Logger{
	Level:  LevelError,
	Output: os.Stdout,
}

func (l *Logger) out(s string) error {
	s = fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), s)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		s += "\n"
	}
	_, err := l.Output.Write([]byte(s))
	return err
}

func Print(v ...interface{}) {
	DefaultLogger.out(PrefixInfo + fmt.Sprint(v...))
}

func Printf(format string, v ...interface{}) {
	DefaultLogger.out(PrefixInfo + fmt.Sprintf(format, v...))
}

func Println(v ...interface{}) {
	DefaultLogger.out(PrefixInfo + fmt.Sprintln(v...))
}

func Fatal(v ...interface{}) {
	DefaultLogger.out(PrefixFatal + fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	DefaultLogger.out(PrefixFatal + fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	DefaultLogger.out(PrefixFatal + fmt.Sprintln(v...))
	os.Exit(1)
}

func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	DefaultLogger.out(PrefixPanic + s)
	panic(s)
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	DefaultLogger.out(PrefixPanic + s)
	panic(s)
}

func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	DefaultLogger.out(PrefixPanic + s)
	panic(s)
}

func Debug(v ...interface{}) {
	if DefaultLogger.Level < LevelDebug {
		return
	}
	DefaultLogger.out(PrefixDebug + fmt.Sprint(v...))
}

func Debugf(format string, v ...interface{}) {
	if DefaultLogger.Level < LevelDebug {
		return
	}
	DefaultLogger.out(PrefixDebug + fmt.Sprintf(format, v...))
}

func Debugln(v ...interface{}) {
	if DefaultLogger.Level < LevelDebug {
		return
	}
	DefaultLogger.out(PrefixDebug + fmt.Sprintln(v...))
}
func Warn(v ...interface{}) {
	if DefaultLogger.Level < LevelWarning {
		return
	}
	DefaultLogger.out(PrefixWarning + fmt.Sprint(v...))
}

func Warnf(format string, v ...interface{}) {
	if DefaultLogger.Level < LevelWarning {
		return
	}
	DefaultLogger.out(PrefixWarning + fmt.Sprintf(format, v...))
}

func Warnln(v ...interface{}) {
	if DefaultLogger.Level < LevelWarning {
		return
	}
	DefaultLogger.out(PrefixWarning + fmt.Sprintln(v...))
}
