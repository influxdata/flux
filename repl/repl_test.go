package repl

import (
	"context"
	"testing"

	"github.com/influxdata/flux/fluxinit"
	"github.com/stretchr/testify/require"
)

func TestReplNewLine(t *testing.T) {
	ctx := context.TODO()
	fluxinit.FluxInit()

	r := New(ctx)
	errI, err := r.executeLine(`import "sampledata"`)
	require.Nil(t, errI)
	require.Nil(t, err)

	errI, err = r.executeLine(`sampledata.int() \`)
	require.Nil(t, errI)
	require.Nil(t, err)

	errI, err = r.executeLine(`|> sum()`)
	require.Nil(t, errI)
	require.Nil(t, err)
}
