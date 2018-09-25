# Planner Design

This document lays out the design of the planner.
The planner is responsible for searching the space of equivalent execution plans for a low cost plan.
Plans have a cost which is a set of metrics which approximates the physical cost of executing the plan.

## Operations

The planner manipulates operations. An operation represents a transformation on the data in some form.

### Logical vs Physical

Operations can be either "logical" or "physical".
A logical operation is one that represents what operation to perform but does not specify how to execute that operation.
A physical operation is one that represents the specific execution of a specific operation.
For example join is a logical operation while merge-sort and hash-join are two different physical operations of the logical join operation.

## Requirements

The planner needs to be capable of the following:

* Search the space of logical and physcial plans
* Compute a cost for a given plan
* Pick a plan for execution
* Be agnostic to the set of logical and physical operations.
* Both logical and physical operations can be rewritten into new sets of physical and logical operations.
    For example a physical operation that represents a read from storage needs to be able to rewrite filter operations
    into a new physical filter+read operation so as to optimize the filter by pushing it down into the storage layer.
    This must be possible without the planner understanding the semantic meaning of the operations.
* Limit its execution time by some factor so as to avoid the planning step adding significant latency.


## Operation Behavior

The planner will not know the semantic meaning of each of the operations, but it can know about specific behavior of the operations.
As discussed above each operation is either a logical or a physical operation.
Operations can have many more properties.

### Commutative Property

An operation can commute with other operations.
The planner can leverage this behavior to explore other possible execution plans.

### Narrow vs Wide

Operations can be classified as either narrow or wide:

* Narrow operations map each parent table to exactly one child table.
    Specifically a narrow operation is a one-to-one mapping of parent to child tables.
* Wide operations map multiple parent tables to multiple child tables.
    Specifically a wide operation is a many-to-many mapping of parent to child tables.


### And more

There will be many more behaviors/properties that operations can expose to the planner.
