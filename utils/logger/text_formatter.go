package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	defaultTextFormatterStackKey = "stack"
)

type TextFormatter struct {
	logrus.TextFormatter
	StackKey                  string
	EnableFuncCallDepthRecord bool
}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if f.StackKey == "" {
		f.StackKey = defaultTextFormatterStackKey
	}

	if f.EnableFuncCallDepthRecord {
		file, line := getCaller(1, "/logrus", "asm_amd64.s")
		entry.Data[f.StackKey] = fmt.Sprintf("%s:%d", file, line)
	}
	return f.TextFormatter.Format(entry)
}
