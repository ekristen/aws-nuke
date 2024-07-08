package testsuite

import (
	"testing"

	"github.com/sirupsen/logrus"
)

type GlobalHook struct {
	T  *testing.T
	TF func(t *testing.T, e *logrus.Entry)
}

func (h *GlobalHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *GlobalHook) Fire(e *logrus.Entry) error {
	if h.TF != nil {
		h.TF(h.T, e)
	}

	return nil
}

func (h *GlobalHook) Cleanup() {
	logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))
}

// NewGlobalHook creates a new GlobalHook for logrus testing
func NewGlobalHook(t *testing.T, tf func(t *testing.T, e *logrus.Entry)) *GlobalHook {
	gh := &GlobalHook{
		T:  t,
		TF: tf,
	}
	logrus.SetReportCaller(true)
	logrus.AddHook(gh)
	return gh
}
