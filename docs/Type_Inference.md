## Flux and Type Inference


Recently Flux saw an update that introduced a language feature called _type inference_.
What is type inference and how does it help me as a user of InfluxDB?

What is type inference?


Natural data types assigned to the data.
1 + 1



How does it help


catch errors earlier

errors that make more sense

LSP

Smart completion


Smart integration with to UI


Optimized code
knowing types statically allows for the Flux engine to be more efficient


For the curious how does it work?


We use algorithm W to determine the type of every expression in a Flux script.


Looks at how the variable is used and starts to restrict which types can work with which variables.

Remember algebra and solving a system of equations? Given that you have 2 variables and 2 equations you can solve for all the variables.

2x + y = 5

y - x = 2


y = 2 + x


2x + 2 + x = 5

3x = 3
x = 1


y = 3


2 + 3 = 5


This is similar, for each use of a variable we can write an equation. Once we have all the equations we solve them for each of the variables be replacing variables with their equation.
Eventually we learn the type of each variable.



Algorithm W is a very specific way to implement the creation of the equations and solving them such that it is efficient and will always have a known solution.

Interseted in learning more here are some pointers:

Algorithm W is a class of type inference algorithm known as Hindley-Milner . 
We also use an extension to Algorithm W for extensible records see the paper here.
Try out the language SML it has a very simple type system based on Hindley-Milner and is a quick language to learn.
Its similar to Haskel and OCaml but drastically easier to learn and simpler to understand.

I personally found no shortcuts in trying to learn type inference. I would start with lambda calculus as all literature about Hindley-Milner assumes familiarity with lambda calculus.
Then repetition and practice were key. 
