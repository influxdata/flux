package plan

import (
	"fmt"
	"time"

	"github.com/influxdata/flux"
)

type Administration interface {
	Now() time.Time
}

// CreateProcedureSpec creates a ProcedureSpec from an OperationSpec and Administration
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

var ruleNameToLogicalRule = make(map[string]Rule)
var ruleNameToPhysicalRule = make(map[string]Rule)

// RegisterLogicalRule registers the rule created by createFn with the logical plan.
func RegisterLogicalRule(rule Rule) {
	registerRule(ruleNameToLogicalRule, rule)
}

// RegisterPhysicalRule registers the rule created by createFn with the physical plan.
func RegisterPhysicalRule(rule Rule) {
	registerRule(ruleNameToPhysicalRule, rule)
}

func registerRule(ruleMap map[string]Rule, rule Rule) {
	name := rule.Name()
	if _, ok := ruleMap[name]; ok {
		panic(fmt.Errorf(`rule with name "%v" has already been registered`, name))
	}
	ruleMap[name] = rule
}
