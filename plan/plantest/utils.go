package plantest

import (
	"github.com/influxdata/flux/plan"
	"github.com/satori/go.uuid"
)

func RandomProcedureID() plan.ProcedureID {
	return plan.ProcedureID(uuid.NewV4())
}
