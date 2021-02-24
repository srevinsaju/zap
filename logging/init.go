package logging

import (
	"github.com/withmandala/go-log"
	"os"
)

var Logger *log.Logger

/* GetLogger returns a pre-initialized logger instance if its available, else creates
a new one and return them after checking debug specifications */
func GetLogger() *log.Logger {
	if Logger != nil {
		return Logger
	}

	Logger = log.New(os.Stdout).WithColor()
	if os.Getenv("ZAP_DEBUG") == "1" {
		Logger = Logger.WithDebug()
	}

	return Logger

}