package logger

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/sirupsen/logrus"
)

// LogrusHandler converts logrus hook to log handler.
func LogrusHandler(hook logrus.Hook) log.Handler {
	return log.FuncHandler(func(r *log.Record) error {
		data := &logrus.Entry{
			Message: r.Msg,
			Data:    fields(r.Ctx),
		}
		return hook.Fire(data)
	})
}

func fields(ctx []interface{}) logrus.Fields {
	ff := make(logrus.Fields, len(ctx)/2)
	for i := 1; i < len(ctx); i += 2 {
		k, ok := ctx[i-1].(string)
		if !ok {
			continue
		}
		ff[k] = ctx[i]
	}

	return ff
}
