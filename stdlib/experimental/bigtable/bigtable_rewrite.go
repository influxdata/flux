package bigtable

import (
	"cloud.google.com/go/bigtable"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe"
	"time"
)

func AddFilterToNode(queryNode plan.Node, filterNode plan.Node) (plan.Node, bool) {
	querySpec := queryNode.ProcedureSpec().(*FromBigtableProcedureSpec)
	filterSpec := filterNode.ProcedureSpec().(*universe.FilterProcedureSpec)

	switch body := filterSpec.Fn.Block.Body.(type) {
	case *semantic.BinaryExpression:
		switch body.Operator {
		case ast.EqualOperator:
			// Look for a Single Row filter
			if isRRowKey(body.Left) {
				if name, ok := body.Right.(*semantic.StringLiteral); ok {
					querySpec.RowSet = bigtable.SingleRow(name.Value)
					return queryNode, true
				}
			}
			// Look for a Family filter
			if isRFamily(body.Left) {
				if family, ok := body.Right.(*semantic.StringLiteral); ok {
					querySpec.Filter = bigtable.ChainFilters(querySpec.Filter, bigtable.FamilyFilter(family.Value))
					return queryNode, true
				}
			}
		case ast.GreaterThanEqualOperator:
			// Look for an Infinite Range filter (>=)
			if isRRowKey(body.Left) {
				if name, ok := body.Right.(*semantic.StringLiteral); ok {
					querySpec.RowSet = bigtable.InfiniteRange(name.Value)
					return queryNode, true
				}
			}
			// Filter from startTime with no upper bound
			if isRTime(body.Left) {
				if startTime, ok := body.Right.(*semantic.DateTimeLiteral); ok {
					querySpec.Filter = bigtable.ChainFilters(querySpec.Filter, bigtable.TimestampRangeFilter(startTime.Value, time.Time{}))
					return queryNode, true
				}
			}
		case ast.LessThanOperator:
			// Filter to endTime with no lower bound
			if isRTime(body.Left) {
				if endTime, ok := body.Right.(*semantic.DateTimeLiteral); ok {
					querySpec.Filter = bigtable.ChainFilters(querySpec.Filter, bigtable.TimestampRangeFilter(time.Time{}, endTime.Value))
					return queryNode, true
				}
			}
		}
	case *semantic.LogicalExpression:
		// Look for a Range filter
		if begin, end, ok := getRange(body); ok {
			querySpec.RowSet = bigtable.NewRange(begin, end)
			return queryNode, true
		}
		// Look for Timestamp Range filter
		if startTime, endTime, ok := getTimeRange(body); ok {
			querySpec.Filter = bigtable.ChainFilters(querySpec.Filter, bigtable.TimestampRangeFilter(startTime, endTime))
			return queryNode, true
		}
	// Look for a Prefix filter
	case *semantic.CallExpression:
		if prefix, ok := getPrefix(body); ok {
			querySpec.RowSet = bigtable.PrefixRange(prefix)
			return queryNode, true
		}
	}

	return filterNode, false
}

func AddLimitToNode(queryNode plan.Node, limitNode plan.Node) (plan.Node, bool) {
	querySpec := queryNode.ProcedureSpec().(*FromBigtableProcedureSpec)
	limitSpec := limitNode.ProcedureSpec().(*universe.LimitProcedureSpec)

	if limitSpec.Offset != 0 {
		return limitNode, false
	}

	querySpec.ReadOptions = append(querySpec.ReadOptions, bigtable.LimitRows(limitSpec.N))
	return queryNode, true
}

func getRange(logic *semantic.LogicalExpression) (string, string, bool) {
	if logic.Operator == ast.AndOperator {
		if left, ok := logic.Left.(*semantic.BinaryExpression); ok {
			if right, ok := logic.Right.(*semantic.BinaryExpression); ok {
				if isRRowKey(left.Left) && isRRowKey(right.Left) {
					if left.Operator == ast.GreaterThanEqualOperator && right.Operator == ast.LessThanOperator {
						if leftVal, ok := left.Right.(*semantic.StringLiteral); ok {
							if rightVal, ok := right.Right.(*semantic.StringLiteral); ok {
								return leftVal.Value, rightVal.Value, ok
							}
						}
					} else if left.Operator == ast.LessThanOperator && right.Operator == ast.GreaterThanEqualOperator {
						if leftVal, ok := left.Right.(*semantic.StringLiteral); ok {
							if rightVal, ok := right.Right.(*semantic.StringLiteral); ok {
								return rightVal.Value, leftVal.Value, ok
							}
						}
					}
				}

			}
		}
	}
	return "", "", false
}

func getTimeRange(logic *semantic.LogicalExpression) (time.Time, time.Time, bool) {
	if logic.Operator == ast.AndOperator {
		if left, ok := logic.Left.(*semantic.BinaryExpression); ok {
			if right, ok := logic.Right.(*semantic.BinaryExpression); ok {
				if isRTime(left.Left) && isRTime(right.Left) {
					if left.Operator == ast.GreaterThanEqualOperator && right.Operator == ast.LessThanOperator {
						if leftVal, ok := left.Right.(*semantic.DateTimeLiteral); ok {
							if rightVal, ok := right.Right.(*semantic.DateTimeLiteral); ok {
								return leftVal.Value, rightVal.Value, ok
							}
						}
					} else if left.Operator == ast.LessThanOperator && right.Operator == ast.GreaterThanEqualOperator {
						if leftVal, ok := left.Right.(*semantic.DateTimeLiteral); ok {
							if rightVal, ok := right.Right.(*semantic.DateTimeLiteral); ok {
								return rightVal.Value, leftVal.Value, ok
							}
						}
					}
				}
			}
		}
	}
	return time.Time{}, time.Time{}, false
}

func getPrefix(callExpression *semantic.CallExpression) (string, bool) {
	if callee, ok := callExpression.Callee.(*semantic.MemberExpression); ok && callee.Property == "hasPrefix" {
		var rowKey bool
		prefix := ""
		for _, prop := range callExpression.Arguments.Properties {
			if key, ok := prop.Key.(*semantic.Identifier); ok {
				if isRRowKey(prop.Value) {
					rowKey = true
				} else if key.Name == "prefix" {
					if val, ok := prop.Value.(*semantic.StringLiteral); ok {
						prefix = val.Value
					}
				}
			}
		}
		return prefix, prefix != "" && rowKey
	}
	return "", false
}

// helper function to identify `r.rowKey`
func isRRowKey(i interface{}) bool {
	if exp, ok := i.(*semantic.MemberExpression); ok {
		if obj, ok := exp.Object.(*semantic.IdentifierExpression); ok && obj.Name == "r" {
			return exp.Property == "rowKey"
		}
	}
	return false
}

// helper function to identify `r.family`
func isRFamily(i interface{}) bool {
	if exp, ok := i.(*semantic.MemberExpression); ok {
		if obj, ok := exp.Object.(*semantic.IdentifierExpression); ok && obj.Name == "r" {
			return exp.Property == "family"
		}
	}
	return false
}

// helper function to identify `r._time`
func isRTime(i interface{}) bool {
	if exp, ok := i.(*semantic.MemberExpression); ok {
		if obj, ok := exp.Object.(*semantic.IdentifierExpression); ok && obj.Name == "r" {
			return exp.Property == "_time"
		}
	}
	return false
}
