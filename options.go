package telejoon

import (
	"github.com/sirupsen/logrus"
)

type Options struct {
	ErrorGroupID int64
	Logger       *logrus.Logger
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) SetLogger(logger *logrus.Logger) *Options {
	o.Logger = logger
	return o
}

func (o *Options) SetErrorGroupID(groupID int64) *Options {
	o.ErrorGroupID = groupID
	return o
}
