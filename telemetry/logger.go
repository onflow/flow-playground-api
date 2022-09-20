package telemetry

import "github.com/sirupsen/logrus"

const loggerActive = true

var logger *logrus.Logger

func DebugLog(message string) {
	if loggerActive {
		if logger == nil {
			logger = logrus.StandardLogger()
		}
		logger.Info(message)
	}
}
