### Type Inference in Flux
--------------------------

Flux is a strongly and statically typed language supporting parametric polymorphism.
Flux does not require explicit type annotions but rather infers the most general type of an expression.

#### Key Concepts

1. Monotypes

    Monotypes are non-parameterized types.
    Examples include `int`, `string`, `boolean`, `(x: int) => int`, etc.

2. Polytypes

    Polytypes are parameterized types sometimes referred to as type schemes in other literature.
    Parameters are type variables that can be substituted with any monotype.
    The following is a polytype with a single parameter `T`

        (x: T) => T

3. Constraints

    Type inference generates constraints that are later solved in order to determine the type of every expression in a flux program.
    Type inference generates two types of constraints - equality constraints and kind constraints.
    An equality constraint asserts that two types are equal.
    A kind constraint is used for implementing ad hoc polymorphism.
    It asserts that a type is one of a finite set of types.

4. Substitution

    A substitution is a map.
    It maps type variables to monotypes.

5. Unification

    Unification is the process of solving type constraints.
    Concretely unification is the process of solving for the type variables in a set of type constraints.
    The output of unification is a substitution.

6. Type Environment

    A type environment maps program identifiers to their corresponding polytypes.

7. Generalization

    Generalization is the process of converting a monotype into a polytype.
    See https://en.wikipedia.org/wiki/Hindley%E2%80%93Milner_type_system#Let-polymorphism_2.

8. Specialization

    Specialization is the process of converting a polytype into a monotype.
    The monotype returned has new fresh type variables with respect to the current type environment.
    Specialization and generalization are used for implementing parametric polymorphism

#### Algorithm

The type inference algorithm that Flux uses is based on Wand's algorithm.
It operates in two phases.
First it generates a series of type constraints for a given expression.
Then it solves those constraints using a process called unification.

Example:
```
f = (a, b) => a + b
x = f(a: 0, b: 1)
```

The algorithm will generate the following constraints for the function expression:

    typeof(a) = typeof(a + b)
    typeof(b) = typeof(a + b)
    typeof(a) in [int, float, string]
    typeof(b) in [int, float, string]

Note the first two constraints are equality constraints whereas the latter two constraints are kind constraints.
After unification we've inferred a monotype for the function expression.
We then generalize this monotype and associate `f` with the resulting polytype in the type environment.

The algorithm then generates the following constraints for the call expression:

    typeof(f) = (a: int, b: int) => t0
    typeof(f) = instantiate(environment(f))

The algorithm continues in the same way, generalizing the inferred type for the call expression and adding a new mapping for `x` in the type environment.

#### Polymorphism

Flux supports the following types of polymorphism.

##### Parametric Polymorphism

Parametric polymorphism is the notion that a function can be applied uniformly to arguments of any type.
The identity function `(x) => x` is one such example of a function that can be applied to any type.

##### Record Polymorphism

Record polymorphism is the notion that a function can be applied to records of different types so long as they contain the necessary properties.
The necessary properties of a record are determined by the use of a record.
For example, the following function asserts that `r` must be a record having a label `a`.

    f = (r) => r.a

Record polymorphism allows for one to pass `f` any record so long as it has a label `a`.
The following records are all valid inputs to `f`.

    {a: 0}
    {a: "string", b: 1}
    {c: "string", a: 1.1}

`{b: 0, c: 1}` however is not a valid input to `f` and the flux type checker will catch any cases where such a type is passed to `f`.

##### Ad hoc Polymorphism

Ad hoc polymorphism is the notion that a function can be applied to a finite set of types, with different behavior depending on the type.
For example the `add` function does not support all types.

    add = (a, b) => a + b

It supports integers, floating point numbers, and even strings as `+` represents concatenation for string types.
However boolean types are not supported and the flux type checker will catch any cases where unsupported types such as booleans are passed to `add`.
