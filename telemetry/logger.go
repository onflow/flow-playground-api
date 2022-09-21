package telemetry

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"runtime"
	"strconv"
	"time"
)

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func DebugLog(message string) {
	if logger == nil {
		logger = logrus.StandardLogger()
	}

	logger.
		WithField("timestamp", time.Now().UnixMilli()).
		WithField("subroutine ID", getGID()).
		Info(message)
}
