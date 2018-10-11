### Type Inference in Flux
--------------------------

Flux is a strongly and statically typed language.
Flux does not require explicit type annotions but rather infers the most general type of a variable.
Flux supports parametric polymorphism.

#### Rules
----------

In order to perform type inference Flux has a collection of rules that govern the relationship between types and expressions.
Take the following program for example.

```
f = (a, b) => return a + b
x = f(a:1, b:2)
```

The following rules will be used to solve for the types in the above program:

1. Binary Addition Rule

    typeof(a + b) = typeof(a)
    typeof(a + b) = typeof(b)

2. Function Rule

    typeof(f) = [ typeof(a), typeof(b) ] -> typeof(return statement)

3. Call Expression Rule

    typeof(f) = [ typeof(a), typeof(b) ] -> typeof( f(a:1, b:2) )

4. Variable Declaration Rule

    typeof(x) = typeof( f(a:1, b:2) )

#### Algorithm
--------------

The type inference algorithm that Flux uses is based on algorithm W.
The algorithm operates on a semantic graph.
The algorithm builds off some basic concepts.

1. Monotype and polytypes

    A type can be a monotype or a polytype.
    Monotypes can be only one type, for example `int`, `string` and `boolean` are monotypes.
    Polytypes can be all types, for example function type `{x:t0} -> t0` takes as input an object with a single key `x` that can have any type.
    The return value of the function is the type of the `x` parameter.
    This function's type is a polytype because it is defined for all possible types.
    It is said the `t0` is a free type variable because it is free to be any type.

2. Type expression

    A type expression is description of a type.
    Type expressions may describe monotypes and polytypes.

3. Type annotations

    Every node in a semantic graph may have a type.
    A type annotation records the node's type.
    A node's type may be unknown, in which case its annotation is an instance of a type variable.
    A type variable is a placeholder for an unknown type.

4. Type environment

    A type environment maps nodes to a type scheme. A type scheme is a type expression with a list of free type variables.
    A scheme can be "instantiated" into an equivalent but unique type expression.
    This instantiation process is what enables parametric polymorphism.

5. Constraints

    A constraint is a requirement that a given type expression be equal to another type expression.
    As the semantic graph is traversed constraints are applied by unifying types based on the rules of a given node.

6. Unification

    To unify two types is to ensure that the two types are equal.
    If the types are not equal then a type error has occurred.
    When a type is unified with a type variable, the variable is "set" to the value of that type.
    By updating the type variables in this manner the solution to the inference problem is determined.


The algorithm walks the semantic graph and for each node it does the following:

1. Set an empty type annotation on the node.
2. Update the current type environment based on the scoping rules.
3. Recurse into each child node, i.e. depth first traversal.
4. Apply all constraints as unify operations on child node types.
5. Annotate the type of the current node.

Once the process is completed the type is known for all nodes or an error has occurred.
Nodes that are polymorphic will have polytypes and monomorphic nodes will have monotypes.


