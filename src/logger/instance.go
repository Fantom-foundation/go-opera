package logger

import (
	"github.com/sirupsen/logrus"
)

type Instance struct {
	*logrus.Entry
}

func MakeInstance() Instance {
	return Instance{logrus.NewEntry(Get())}

}

func (i *Instance) SetName(host string) {
	i.Entry = Get().WithField("node", host)
}
