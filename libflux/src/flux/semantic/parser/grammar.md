# Polytype Parser Grammar 

[Modified from Joshua Lapacik's work here](https://github.com/jlapacik/ftel/blob/master/antlr/flux.g4)

```
? = optional; once or none 
* = zero or more times
| = or 
(values) = a list of possibly repeating values
single quotes ('') indicate a literal value


polytype    = 'forall' '[' vars? ']' ('where' constraints)? monotype

vars        = type_var (',' type_var)* 
constraints = constraint ( (',' | 'and') constraint)* 
constraint  = type_var (':') kinds
kinds       = kind ( '+' kind)*
kind        = IDENTIFIER 
monotype    = type_var | primitive | array | row | function

type_var    = 't' ([0-9])*
primitive   = INT | FLOAT | STRING | BOOL | DURATION | TIME | REGEXP | BYTES
array       = '[' monotype ']'
row         = '{' properties? '}'
function    = '(' arguments? ')' '->' monotype
properties  = property ( '|' property )* ( '|' type_var)?
property    = IDENTIFIER ':' monotype
arguments   = argument ( ',' argument )*
argument    = required | optional | pipe
required    = IDENTIFIER ':' monotype
optional    = '?' IDENTIFIER ':' monotype
pipe        = '<-' IDENTIFIER? ':' monotype


INT         = 'int'
UINT        = 'uint'
FLOAT       = 'float'
STRING      = 'string'
BOOL        = 'bool'
DURATION    = 'duration'
TIME        = 'time'
REGEXP      = 'regexp'
BYTES       = 'bytes'
IDENTIFIER  = [a-zA-Z_] ([0-9a-zA-Z_])* | '"' [a-zA-Z_] ([0-9a-zA-Z_])* '"'
WHITESPACE  = [ \t\r\n]+ -> skip
```