package nuke

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ekristen/libnuke/pkg/resource"
)

const (
	Account resource.Scope = "account"
)

type ListerOpts struct {
	Region  *Region
	Session *session.Session
}

func (o ListerOpts) ID() string {
	return ""
}

type Lister struct {
	opts ListerOpts
}

func (l *Lister) SetOptions(opts interface{}) {
	l.opts = opts.(ListerOpts)
}
