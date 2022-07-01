//go:build !debug
// +build !debug

package execute

import (
	"fmt"
	"runtime/debug"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (es *executionState) recover() {
	if e := recover(); e != nil {
		// We had a panic, abort the entire execution.
		err, ok := e.(error)
		if !ok {
			err = fmt.Errorf("%v", e)
		}

		if errors.Code(err) == codes.ResourceExhausted {
			es.abort(err)
			return
		}

		err = errors.Wrap(err, codes.Internal, "panic")
		es.abort(err)
		if entry := es.logger.Check(zapcore.InfoLevel, "Execute source panic"); entry != nil {
			entry.Stack = string(debug.Stack())
			entry.Write(zap.Error(err))
		}
	}
}

func (d *poolDispatcher) recover() {
	if e := recover(); e != nil {
		err, ok := e.(error)
		if !ok {
			err = fmt.Errorf("%v", e)
		}

		if errors.Code(err) == codes.ResourceExhausted {
			d.setErr(err)
			return
		}

		err = errors.Wrap(err, codes.Internal, "panic")
		d.setErr(err)
		if entry := d.logger.Check(zapcore.InfoLevel, "Dispatcher panic"); entry != nil {
			entry.Stack = string(debug.Stack())
			entry.Write(zap.Error(err))
		}
	}
}
