package klog

import (
	"fmt"
	"io"
	"os"
)

const (
	LevelInfo uint = iota
	LevelError
	LevelWarning
	LevelDebug
)

const ()

type Logger struct {
	Level  uint
	Output io.Writer
}

var DefaultLogger = &Logger{
	Level:  LevelError,
	Output: os.Stderr,
}

func (l *Logger) out(s string) error {
	if len(s) == 0 || s[len(s)-1] != '\n' {
		s += "\n"
	}
	_, err := l.Output.Write([]byte(s))
	return err
}

func Print(v ...interface{}) {
	DefaultLogger.out(fmt.Sprint(v...))
}

func Printf(format string, v ...interface{}) {
	DefaultLogger.out(fmt.Sprintf(format, v...))
}

func Println(v ...interface{}) {
	DefaultLogger.out(fmt.Sprintln(v...))
}

func Fatal(v ...interface{}) {
	DefaultLogger.out(fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	DefaultLogger.out(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	DefaultLogger.out(fmt.Sprintln(v...))
	os.Exit(1)
}

func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	DefaultLogger.out(s)
	panic(s)
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	DefaultLogger.out(s)
	panic(s)
}

func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	DefaultLogger.out(s)
	panic(s)
}

func Debug(v ...interface{}) {
	if DefaultLogger.Level < LevelDebug {
		return
	}
	DefaultLogger.out(fmt.Sprint(v...))
}

func Debugf(format string, v ...interface{}) {
	if DefaultLogger.Level < LevelDebug {
		return
	}
	DefaultLogger.out(fmt.Sprintf(format, v...))
}

func Debugln(v ...interface{}) {
	if DefaultLogger.Level < LevelDebug {
		return
	}
	DefaultLogger.out(fmt.Sprintln(v...))
}
func Warn(v ...interface{}) {
	if DefaultLogger.Level < LevelWarning {
		return
	}
	DefaultLogger.out(fmt.Sprint(v...))
}

func Warnf(format string, v ...interface{}) {
	if DefaultLogger.Level < LevelWarning {
		return
	}
	DefaultLogger.out(fmt.Sprintf(format, v...))
}

func Warnln(v ...interface{}) {
	if DefaultLogger.Level < LevelWarning {
		return
	}
	DefaultLogger.out(fmt.Sprintln(v...))
}
