#  ?????

Have you seen this when using the Flux code editor?

<Insert GIF of sweet autocompletetion>

Pretty neat to be able to get that much help from an editor while writing code.

Have you ever wondered how that worked? Today I'll take us on a bit of a deep dive on the behind the scenes that enable these autocompletion features in the editor.
You may also be surprised to learn that the same features that make autocompletion work in the editor also make Flux safer and faster when working with data.


Let's jump right in, there are two main properties of Flux that make the above possible:

1. Flux is strongly typed
2. Flux is statically typed

What do these terms mean?

First let's define what a _type_ is in the context of programming languages.
A _type_ in a programming language defines what kinds of values are allowed for a given variable.
For example if I have the variable `threshold` in my Flux script and its type is `integer` then I know that the value of `threshold` can only be a whole number, it cannot be a string, a duration, or even a table of data.

Flux is _strongly_ typed means that all variables in a Flux program have a known type and they do not change automatically.
Flux is _statically_ typed means that Flux knows the type of the variables without having to run the program. This is sometimes described as knowing the types as compile time instead of at runtime.
The opposite of a strongly statically typed language is a weakly dynamically typed language where variables do not have a known type until runtime and their type can change automatically.
The languages Go and C are strong and static while Javascript is weak and dynamic.

Here is a quick example to illustrate these points:

    a = "1"
    b = 1
    a + b

What happens if you execute that code in Flux?
You will get this error:

    Error: type error 3:1-3:6: string != int

This is because Flux knows the types of `a` and `b` and you cannot add a string to an integer.

What happens if you execute the same code in Javascript?
You will get the string `'11'`. This is because Javascript automatically converted the type of `b` from a number to a string and then performed string concatenation.

> As an aside, this is not a judgement on strong static languages versus weak dynamic languages. Each has their place and I hope this reading makes it clear why Flux chooses to be strongly and statically typed.

Now armed with some new concepts let's break down how this helps us build rich editor features.
Take for example this feature where we can make autocompletion suggestions based on the name of a function.

<Insert Function Autocompletion GIF>

This works by inspecting the code determining the type of the function by its name and then crafting some suggestions from the type.
A functions type contains the information about which parameters it needs, what type those parameters have as well as the return type of the function.
Using that information we construct detailed suggestions for what to insert next because we know exactly what parameters the function needs.

This is only possible because we can know the function's type based on its name without having to execute the Flux program, i.e. static typing.

As result of these two properties of Flux we get many benefits.
We have seen the benefits it provides to editor features like the above.
These editor features make developing Flux code a more productive experience because you get quick feedback as you write the code.
Things like autocompletion make it quicker to write code and reduces typos.
Helpful error messages that can be precise about the mistake make it quicker to fix typos.
All of these editor features make you more productive.

Additionally Flux can execute faster and safer because of its strong and static typing.
The code Flux executes is faster because we know the type of the data ahead of time so when we process the data we do not have to constantly check the type of the data.

Flux code also safer because data in Flux tables is known so it is consistent.
A type is not restricted to just simple types like `float` or `string`.
A type can also be composed of other types, for example the Flux table type is described as a set of columns where the type of each column is known.
This means that you do not have to deal with the class of data cleaning problems where some of your data is of a different type to the rest of the data.
Flux can ensure all records have the same type.

At this point I hope its clear why we have designed Flux to be a strong and static language.
If you would like to go a layer deeper we can explore how we implemented Flux such that it is both statically and strongly typed.


## Extra Credit

For extra credit what follows is an explanation of the process of how Flux determines the types of values within a program.
This process is known as _type inference_ and we use a specific implementation known as Algorithm W.

Type inference is the idea that Flux can determine the type of the values within the program without requiring that the user to be explicit about types in the code.
For example Flux knowns that the type of this line of code is a string:

    message = "The system is down"

To a reader of the code its obvious that message is a string because they can easily see its value is a string.
Other strongly typed programming languages require that the author of the code tell the program that `message` is a string.
For example here is the equivalent Go code:

    var message string = "The system is down"

The `string` keyword tells the Go compiler that the variable `message` has the string type.

> If you are familiar with Go, you may be thinking that you can use the `:=` syntax instead. That is correct and is in fact a limited case of type inference within Go as well.

In contrast Flux does not require those type annotations in the code because it can infer the type of any value without them.

Remember algebra and solving a system of equations?
It's those problems where they gave two equations that had two variables and you had to solve for both variables.
Those problems are solved in a very similar manner to how Flux solves for the types in a program.
For example given these two equations with variables `x` and `y` you can solve for both variables.

    2x + y = 5
    y = 2 + x


The process of solving these variable involves _substituting_ one variable equation into the other equations.
We have an equation what `y` is equal to, so we can substitute all values of `y` in the first equation with the right hand side of `y = ` equation.

    2x + (2 + x) = 5

Now that we have only a single variable `x` in the equation we _solve_ for `x` using the rules of algebra.
Here are the steps explicitly:

    2x + (2 + x) = 5
    3x + 2 = 5
    3x = 3
    x = 1

We have solved for `x` and learned its value is `1`. Now we can use the equation for `y` to determine its value.

    y = 2 + x
    y = 3


Using this process we have solved for the values of `x` and `y`.

Type inference is similar except that instead of solving a system of algebraic equations we solve equations about types.
We use _substitution_ and the _rules_ of the types to solve for the type of every expression in a Flux program.

Here is an overview of the process:

1. Assign a type variable to each Flux expression in the program.
2. Generate all the equations about the types.
3. Solve that set of equations.

First, for each expression in a Flux script we assign it a new _type variable_.
We call them _type variables_ because they represent types and we do not know which type until we solve for them.
Our goal is to solve for those type variables and in order to solve them we need some equations.
We get those equations by looking at how each type variable is used within the Flux program.
Once we have collected all the equations we can solve them.
We do this the same way we solve an algebraic equation, by rearranging the equations according to the rules and using substitution to replace variables in the equations until we have a simple `x = Type` equation for each type variable.

Here is a simple _hello world_ Flux program that will print "hello world" when run via the REPL.

    a = "hello "
    b = "world"
    a + b


We can follow the process that type inference takes to know the types of all of the expressions in this program.
I have chosen this example because all expressions in this program have the type string and this makes it easy to see the solution we are working towards as we take each step.

The first step is to assign the type variables to the expressions.
This program has three expressions one for each line.
Let's give these type variables the names `X`, `Y` and `Z` for each line in order.

    a = "hello " // X
    b = "world"  // Y
    a + b        // Z

Next we need to gather all the equations we know about those type variables.
First we have `X` with the expression:

    a = "hello "

Here since we see that only a string literal is used we know that the expression must have a string type.
We can write down an equation for `X` stating that `X` must be equal to the type `string`.

    X = string

We also see that we assigned a value to the variable `a`, so we write down the type variable that is associated with the normal variable like this:

    a -> X

Notice that we are talking about two different kinds of variables, there are variables in the Flux program itself like `a` and `b`, there are also the type variables like `X`, `Y` and `Z`.
To help keep them straight I will use upper case letters for type variables and lower case letters for program variables.
We track the association of the program variables to their corresponding type variables so we can look them up whenever they are used in other expressions.

The next expression for type variable `Y` is the same as `X` from a type perspective, therefore we get the following equation and type variable association.

    Y = string
    b -> Y

Finally we need to write down an equation for `Z`, remember its expression:

    a + b

This expression doesn't have any hints as to what type it is because only variables are used.
However we can still learn something useful from the expression.
We know that both sides of the `+` operator must be the same type. This is what we mean by using the _rules_ of types, certain operators must be used in specific ways, we write down equations to capture those constraints.
Since we don't know the type yet we create a new type variable `W` as a placeholder for now.
We can write down some more equations using `W` and the `lookup` function to indicate we need to lookup the association of a program variable to its type variable.

    W = lookup(a)
    W = lookup(b)
    Z = W

We can lookup the type variable for `a` and `b` and update the equations.

    W = X
    W = Y
    Z = W

Now we have examined all the expressions in the program and we can move onto the next step to solve those equations.
This part proceeds in the same way as solving the algebraic equations.
We start substituting variables for their equation until we learn the type of every variable.

Here are all the equations we know at this point:

    X = string
    Y = string
    W = X
    W = Y
    Z = W

We can substitute the type variables `X` and `Y` and get these simplified equations.

    W = string
    W = string
    Z = W

Now we know a value for `W` and we can substitute it in the equation for `Z` like so:

    Z = string

We now have a set of equations and we have solved for all type variables.

    X = string
    Y = string
    Z = string
    W = string

We can now say that all expression in the Flux script are strings as we expected when we first saw the program.

Taking a step back we can start to see how to systematically solve for the types of expressions within a Flux program.
We can also see how we can find errors in programs. For example take this slightly modified example program:

    a = "hello "
    b = 1
    a + b

When creating the type equations for this program we would get the following:

    X = string
    Y = integer
    W = X
    W = Y
    Z = W


Using substitution we would get:

    W = string
    W = integer

At this point we have discovered a type error because `W` cannot be both equal to a string and an integer.
When there is not a valid solution to the types, then there is a type error in the program.


And that's type inference in a nutshell. Generate a bunch of equations based on how expressions are used within the Flux program and then solve that set of equations for all the type variables.

So what was that Algorithm W that was mentioned earlier?

Algorithm W is a very specific way to implement the creation and solving of the type equations.
As you may remember from algebra class, picking which equations to substitute into which other equations can make solving the problem either easy or extremely difficult.
For algebra this becomes a bit of an intuitive art that students get good at with practice. Unfourtunately computers are not good at intuition.
This is where Algorithm W comes in, it is a specific methodoligy for how to generate and solve type equations efficiently and correctly.

Still want to learn more? Here are a few pointers.

Algorithm W is a class of type inference algorithm known as Hindley-Milner.
We also use an extension to Algorithm W for extensible records see the paper here.
Try out the language SML it has a very simple type system based on Hindley-Milner and is a quick language to learn.
Its similar to Haskel and OCaml but drastically easier to learn and simpler to understand.

I personally found no shortcuts in trying to learn type inference. I would start with lambda calculus as all literature about Hindley-Milner assumes familiarity with lambda calculus.
Then repetition and practice were key.
