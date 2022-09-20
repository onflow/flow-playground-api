package telemetry

import "github.com/sirupsen/logrus"

var logger *logrus.Logger

// todo temp telemtry
func Logger() *logrus.Logger {
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	return logger
}
