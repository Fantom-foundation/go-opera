package logger

import (
	"github.com/ethereum/go-ethereum/log"
)

type Instance struct {
	Log log.Logger
}

func MakeInstance() Instance {
	return Instance{
		Log: log.New(),
	}

}

func (i *Instance) SetName(name string) {
	i.Log = log.New("name", name)
}
