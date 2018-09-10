### Type Inference in Flux
--------------------------

Flux is a statically typed language.
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

The type inference algorithm that Flux uses is based on Wand's algorithm.
It is completed in the following 3 phases:

1. Type annotation phase

    Input: Semantic graph  
    Output: Type Environment  

    In this phase every expression is assigned a type variable which is a placeholder for a concrete type.
    A type environment is a mapping from expressions to type variables.

    Example:

        Program:
        
            x = 1 + 2

        Type Environment:

            x   -> t0
            1+2 -> t1
            1   -> t2
            2   -> t3

2. Constraint generation phase

    Input: Type Environment  
    Output: Set of constraints  

    In this phase a set of equations is generated according the Flux type inference rules.

    Example:

        Program:

            x = 1 + 2

        Type Environment:

            x   -> t0
            1+2 -> t1
            1   -> t2
            2   -> t3

        Constraints:

            t0 = t1
            t1 = t2
            t1 = t3
            t2 = int
            t3 = int

3. Sustitution phase

    Input: Set of constraints  
    Output: Set of solutions  

    In this phase variable substitution is performed for the type expressions on the right-hand side of each constraint.
    Any type errors will occur in this phase.
    If no type errors occur, then the result will be a typed program where each variable is given the most general type possible.

    Example:

        Program:

            x = 1 + 2

        Type Environment:

            x   -> t0
            1+2 -> t1
            1   -> t2
            2   -> t3

        Constraints:

            t0 = t1
            t1 = t2
            t1 = t3
            t2 = int
            t3 = int

        Solutions:

            t0 = int
            t1 = int
            t2 = int
            t3 = int

#### Type Expressions
---------------------

A type expression is either:

1. A concrete type such as int, float, bool, etc.
2. A type variable such as t0, t1, t2, etc.
3. A function type expression.

A function type expression is composed of two sets of type expressions - the type expressions for its input parameters and the type expression for its return type.