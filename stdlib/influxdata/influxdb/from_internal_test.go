package influxdb

import (
	"context"

	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/mock"
)

func CreateSource(ctx context.Context, ps *FromRemoteProcedureSpec) (execute.Source, error) {
	id := executetest.RandomDatasetID()
	return createFromSource(ps, id, mock.AdministrationWithContext(ctx))
}
