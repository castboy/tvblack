package logger

import (
	"runtime"
	"strings"
)

func getCaller(callDepth int, containsToIgnore ...string) (file string, line int) {
	callDepth++
	var ok bool
	var isSkip bool
	_, file, line, ok = runtime.Caller(callDepth+1)
	if !ok {
		file = "???"
		line = 0
		return file, line
	}

	for _, s := range containsToIgnore {
		if strings.Contains(file, s) {
			callDepth++
			isSkip = true
			break
		}
		isSkip = false
	}

	if isSkip {
		_, file, line, ok = runtime.Caller(callDepth+3)
		if !ok {
			file = "???"
			line = 0
		}
	}

	return file, line
}
