/*
The semantic package provides a graph structure that represents the meaning of a Flux script.
An AST is converted into a semantic graph for use with other systems.
Using a semantic graph representation of the Flux, enables highlevel meaning to be specified programatically.

The semantic structures are to be designed to facilitate the interpretation and compilation of Flux.

For example since Flux uses the javascript AST structures, arguments to a function are represented as a single positional argument that is always an object expression.
The semantic graph validates that the AST correctly follows these semantics, and use structures that are strongly typed for this expectation.
*/
package semantic
