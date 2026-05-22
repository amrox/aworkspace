package workspace

import (
	"fmt"
	"os"
)

type LogLevel int

const (
	LogLevelAlways = iota
	LogLevelNormal
	LogLevelVerbose
)

var CurLogLevel LogLevel = LogLevelNormal

func Log(level LogLevel, fmtString string, a ...any) int {

	if level <= CurLogLevel {
		n, err := fmt.Fprintf(os.Stderr, fmtString, a...)
		if err != nil {
			panic("Fprintf failed: " + err.Error())
		}
		return n
	}
	return 0
}

func LogWarning(fmtString string, a ...any) int {
	return Log(LogLevelAlways, "WARNING: " + fmtString, a...)
}

func LogInfo(fmtString string, a ...any) int {
	return Log(LogLevelNormal, fmtString, a...)
}

func LogVerbose(fmtString string, a ...any) int {
	return Log(LogLevelVerbose, fmtString, a...)
}

// TODO: LogError?

