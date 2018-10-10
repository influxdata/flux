package plantest

import (
	"github.com/influxdata/flux/planner"
	"github.com/satori/go.uuid"
)

func RandomProcedureID() planner.ProcedureID {
	return planner.ProcedureID(uuid.NewV4())
}
