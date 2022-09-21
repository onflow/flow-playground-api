package telemetry

import (
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
	"time"
)

// runtimeDebugActive determines if runtime debugging is active
const runtimeDebugActive = false

var logger *logrus.Logger

// debugFrame stack frame for debug info
type debugFrame struct {
	FuncName  string
	StartTime time.Time
}

var runtimeStack []debugFrame

// StartRuntimeCalculation adds a debug frame containing the start time of execution for the calling function
// EndRuntimeCalculation should be deferred on the next line
func StartRuntimeCalculation() {
	defer func() {
		recover()
	}()
	if !runtimeDebugActive {
		return
	}
	// Obtain and format the callers file name
	pc, file, _, _ := runtime.Caller(1)
	nameSplit := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	fileSplit := strings.Split(file, "/")
	fileName := strings.Split(fileSplit[len(fileSplit)-1], ".")[0]
	funcName := "[" + fileName + "] " + nameSplit[len(nameSplit)-1]
	// Add start time to runtime stack
	runtimeStack = append(runtimeStack, debugFrame{funcName, time.Now()})
}

// EndRuntimeCalculation calculates the elapsed time from the debug stack frame
// Should be called as a deferred function immediately after StartRuntimeCalculation
func EndRuntimeCalculation() {
	defer func() {
		recover()
	}()
	if !runtimeDebugActive {
		return
	}
	// Calculate elapsed time from previous debug stack frame
	stackSize := len(runtimeStack) - 1
	frame := runtimeStack[stackSize]
	elapsedTime := time.Since(frame.StartTime)
	// pop stack
	runtimeStack = runtimeStack[:stackSize]
	runtimeDebugLog(frame.FuncName + " was executed in " + elapsedTime.String())
}

// runtimeDebugLog logs message if runtimeDebugActive is true
func runtimeDebugLog(message string) {
	defer func() {
		recover()
	}()
	if !runtimeDebugActive {
		return
	}
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	logger.
		WithField("timestamp", time.Now().UnixMilli()).
		WithField("subroutine ID", getGID()).
		Info(message)
}
