package logger

import (
	"github.com/ethereum/go-ethereum/log"
)

type Instance struct {
	Log log.Logger
}

func New(name ...string) Instance {
	if len(name) == 0 {
		return Instance{
			Log: log.New(),
		}
	}
	return Instance{
		Log: log.New("module", name[0]),
	}
}
