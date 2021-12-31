package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	defaultJSONFormatterStackKey = "stack"
)

type JSONFormatter struct {
	logrus.JSONFormatter
	StackKey                  string
	EnableFuncCallDepthRecord bool
}

func (f *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if f.StackKey == "" {
		f.StackKey = defaultTextFormatterStackKey
	}

	if f.EnableFuncCallDepthRecord {
		file, line := getCaller(1, "/logrus", "asm_amd64.s")
		entry.Data[f.StackKey] = fmt.Sprintf("%s:%d", file, line)
	}
	return f.JSONFormatter.Format(entry)
}
