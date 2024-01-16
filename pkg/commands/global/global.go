package global

import (
	"fmt"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func Flags() []cli.Flag {
	globalFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "log-level",
			Usage:   "Log Level",
			Aliases: []string{"l"},
			EnvVars: []string{"LOGLEVEL"},
			Value:   "info",
		},
		&cli.BoolFlag{
			Name:  "log-caller",
			Usage: "log the caller (aka line number and file)",
		},
		&cli.BoolFlag{
			Name:  "log-disable-color",
			Usage: "disable log coloring",
		},
		&cli.BoolFlag{
			Name:  "log-full-timestamp",
			Usage: "force log output to always show full timestamp",
		},
	}

	return globalFlags
}

func Before(c *cli.Context) error {
	formatter := &logrus.TextFormatter{
		DisableColors: c.Bool("log-disable-color"),
		FullTimestamp: c.Bool("log-full-timestamp"),
	}
	if c.Bool("log-caller") {
		logrus.SetReportCaller(true)

		formatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		}
	}

	logrus.SetFormatter(formatter)

	switch c.String("log-level") {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	}

	return nil
}
