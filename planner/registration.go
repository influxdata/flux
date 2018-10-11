package planner

import (
	"fmt"

	"github.com/influxdata/flux"
	uuid "github.com/satori/go.uuid"
)

type ProcedureID uuid.UUID

var NilUUID uuid.UUID
var RootUUID = NilUUID

func (id ProcedureID) String() string {
	return uuid.UUID(id).String()
}

type Administration interface {
	ConvertID(flux.OperationID) ProcedureID
}

func ProcedureIDFromOperationID(id flux.OperationID) ProcedureID {
	return ProcedureID(uuid.NewV5(RootUUID, string(id)))
}

func ProcedureIDFromNodeID(id NodeID) ProcedureID {
	return ProcedureID(uuid.NewV5(RootUUID, string(id)))
}

type CreateProcedureSpec func(flux.OperationSpec, Administration) (ProcedureSpec, error)

var kindToProcedure = make(map[ProcedureKind]CreateProcedureSpec)
var queryOpToProcedure = make(map[flux.OperationKind][]CreateProcedureSpec)

// RegisterProcedureSpec registers a new procedure with the specified kind.
// The call panics if the kind is not unique.
func RegisterProcedureSpec(k ProcedureKind, c CreateProcedureSpec, qks ...flux.OperationKind) {
	if kindToProcedure[k] != nil {
		panic(fmt.Errorf("duplicate registration for procedure kind %v", k))
	}
	kindToProcedure[k] = c
	for _, qk := range qks {
		queryOpToProcedure[qk] = append(queryOpToProcedure[qk], c)
	}
}
