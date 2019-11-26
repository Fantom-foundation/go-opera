package tracing

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
)

var (
	enabled   bool
	txSpans   = make(map[common.Hash]opentracing.Span)
	txSpansMu sync.RWMutex

	noopSpan = opentracing.NoopTracer{}.StartSpan("")
)

func SetEnabled(val bool) {
	enabled = val
}

func StartTx(ctx context.Context, operation string, tx common.Hash) {
	if !enabled {
		return
	}

	span, _ := opentracing.StartSpanFromContext(ctx, operation)

	txSpansMu.Lock()
	defer txSpansMu.Unlock()

	if _, ok := txSpans[tx]; ok {
		//panic("tracing: tx double")
		return
	}
	txSpans[tx] = span
}

func FinishTx(tx common.Hash) {
	if !enabled {
		return
	}

	txSpansMu.Lock()
	defer txSpansMu.Unlock()

	span := txSpans[tx]
	if span == nil {
		//panic("tracing: FinishTx before StartTx")
		return
	}
	delete(txSpans, tx)
	span.Finish()
}

func CheckTx(tx common.Hash, operation string) opentracing.Span {
	if !enabled {
		return noopSpan
	}

	txSpansMu.RLock()
	span := txSpans[tx]
	txSpansMu.RUnlock()

	if span == nil {
		//panic("tracing: CheckTx before StartTx")
		return noopSpan
	}

	return opentracing.GlobalTracer().StartSpan(
		operation,
		opentracing.ChildOf(span.Context()),
	)
}
