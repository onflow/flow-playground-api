package telemetry

import (
	"github.com/sirupsen/logrus"
)

const loggerActive = false

func DebugLog(message string) {
	if !loggerActive {
		return
	}
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	logger.Info(message)
}
