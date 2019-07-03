# Parser Grammar

The SPEC contains an EBNF grammar definition of Flux.

For the parser, the SPEC grammar undergoes a process to have the left-recursion removed and is then left-factored to turn it into an LL(1) compliant grammar. For simplicity, an alternation operation will choose the first production that will accept the token type. This is because an existing production may be factored out for one token type, but may still exist in its current form for other token types when the first token in the production is an alternation over multiple terminals. To avoid creating more production rules that impact readability just to remove the now factored terminal, these are ignored.

The parser directly implements the following grammar.

    File                           = [ PackageClause ] [ ImportList ] StatementList .
    PackageClause                  = "package" identifier .
    ImportList                     = { ImportDeclaration } .
    ImportDeclaration              = "import" [ identifier ] string_lit
    StatementList                  = { Statement } .
    Statement                      = OptionAssignment
                                   | BuiltinStatement
                                   | TestStatement
                                   | IdentStatement
                                   | ReturnStatement
                                   | ExpressionStatement .
    IdentStatement                 = identifer ( AssignStatement | ExpressionSuffix ) .
    OptionAssignment               = "option" identifier OptionAssignmentSuffix .
    OptionAssignmentSuffix         = AssignStatement
                                   | "." identifier AssignStatement .
    BuiltinStatement               = "builtin" identifier .
    TestStatement                  = "test" identifier AssignStatement .
    AssignStatement                = "=" Expression .
    ReturnStatement                = "return" Expression .
    ExpressionStatement            = Expression .
    Expression                     = ConditionalExpression .
    ConditionalExpression          = LogicalOrExpression
                                   | "if" Expression "then" Expression "else" Expression .
    ExpressionSuffix               = { PostfixOperator } { PipeExpressionSuffix } { MultiplicativeExpressionSuffix } { AdditiveExpressionSuffix } { ComparisonExpressionSuffix } { LogicalAndExpressionSuffix } { LogicalOrExpressionSuffix } .
    LogicalOrExpression            = LogicalAndExpression { LogicalOrExpressionSuffix } .
    LogicalOrExpressionSuffix      = LogicalOrOperator LogicalAndExpression .
    LogicalOrOperator              = "or" .
    LogicalAndExpression           = UnaryLogicalExpression { LogicalAndExpressionSuffix } .
    LogicalAndExpressionSuffix     = LogicalAndOperator UnaryLogicalExpression .
    LogicalAndOperator             = "and" .
    UnaryLogicalExpression         = ComparisonExpression
                                   | UnaryLogicalOperator UnaryLogicalExpression .
    UnaryLogicalOperator           = "not" | "exists" .
    ComparisonExpression           = AdditiveExpression { ComparisonExpressionSuffix } .
    ComparisonExpressionSuffix     = ComparisonOperator AdditiveExpression .
    ComparisonOperator             = "==" | "!=" | "<" | "<=" | ">" | ">=" | "=~" | "!~" .
    AdditiveExpression             = MultiplicativeExpression { AdditiveExpressionSuffix } .
    AdditiveExpressionSuffix       = AdditiveOperator MultiplicativeExpression .
    AdditiveOperator               = "+" | "-" .
    MultiplicativeExpression       = PipeExpression { MultiplicativeExpressionSuffix } .
    MultiplicativeExpressionSuffix = MultiplicativeOperator PipeExpression .
    MultiplicativeOperator         = "*"| "/".
    PipeExpression                 = UnaryExpression { PipeExpressionSuffix } .
    PipeExpressionSuffix           = PipeOperator UnaryExpression .
    PipeOperator                   = pipe_forward .
    UnaryExpression                = PostfixExpression
                                   | PrefixOperator UnaryExpression .
    PrefixOperator                 = "+" | "-" .
    PostfixExpression              = PrimaryExpression { PostfixOperator } .
    PostfixOperator                = MemberExpression
                                   | CallExpression
                                   | IndexExpression .
    MemberExpression               = DotExpression  | MemberBracketExpression
    DotExpression                  = "." identifer
    MemberBracketExpression        = "[" string "]" .
    CallExpression                 = "(" ParameterList ")" .
    IndexExpression                = "[" Expression "]" .
    PrimaryExpression              = identifer
                                   | int_lit
                                   | float_lit
                                   | string_lit
                                   | regex_lit
                                   | duration_lit
                                   | pipe_receive_lit
                                   | ObjectLiteral
                                   | ArrayLiteral
                                   | ParenExpression .
    ObjectLiteral                  = "{" ObjectLiteralBody "}"
    ObjectLiteralBody              = [ ObjectBody ]
    ArrayLiteral                   = "[" ExpressionList "]" .
    ParenExpression                = "(" ParenExpressionBody .
    ParenExpressionBody            = ")" FunctionExpressionSuffix
                                   | identifer ParenIdentExpression
                                   | Expression ")" .
    ParenIdentExpression           = ")" [ FunctionExpressionSuffix ]
                                   | "=" Expression [ "," ParameterList ] ")" FunctionExpressionSuffix .
                                   | "," ParameterList ")" FunctionExpressionSuffix
                                   | ExpressionSuffix ")" .
    ParenExpression                = "(" Expression ")" .
    FunctionExpressionSuffix       = "=>" FunctionBodyExpression .
    FunctionBodyExpression         = Block | Expression .
    Block                          = "{" StatementList "}" .
    ExpressionList                 = [ Expression { "," Expression } ] .
    ObjectBody                     = identifier ObjectBodySuffix 
                                   | string_lit PropertyListSuffix .
    ObjectBodySuffix               = "with" PropertyList
                                   | PropertyListSuffix .
    PropertyListSuffix             = PropertySuffix { "," PropertyList } .
    PropertyList                   = [ Property { "," Property } ] .
    Property                       = identifier PropertySuffix
                                   | string_lit PropertySuffix .
    PropertySuffix                 = [ ":" Expression ].
    ParameterList                  = [ Parameter { "," Parameter } ] .
    Parameter                      = identifer [ "=" Expression ] .

When processing the grammar, the parser follows a few simple rules.

1. It will attempt to expand each production that it encounters.
2. If the production accepts the empty set, it will be considered complete when it encounters a token that is not accepted by the grammar.
3. If the production sees a token that it does not accept and it does not accept the empty set, it will generate an error within the AST and skip to the next token.
4. When a production contains an alternation, the parser will choose the first production that accepts the token.
5. At most one production in an alternation can accept the empty set and the empty set will only be used if none of the productions can accept the current token.

To determine which tokens a production accepts, compute `FIRST(X)` for each production with `X` being the name of the production. This is computed by reading each production with the following rules:

1. For a terminal, `FIRST(X) = {X}`.
2. For a alternation, calculate `FIRST(X)` for each production and form the union.
3. For concatentation, calculate `FIRST(X)` for the first production. If this set contains the empty set, calculate `FIRST(X)` for the next production. Continue until you hit a production that does not contain the empty set or, if all productions have been evaluated, then the empty set is accepted for this production.

## Operator Precedence

Operator precedence is defined within the grammar itself so a precedence table is not needed. While the table itself is not needed and not used in the parser itself, a table is provided below for human readability with verifying the grammar.

|Precedence| Operator       |        Description        |
|----------|----------------|---------------------------|
|     1    |  `a()`         |       Function call       |
|          |  `a[]`         |  Member or index access   |
|          |   `.`          |       Member access       |
|     2    | `*` `/`        |Multiplication and division|
|     3    | `+` `-`        | Addition and subtraction  |
|     4    | `==` `!=`      |   Comparison operators    |
|          | `<` `<=`       |                           |
|          | `>` `>=`       |                           |
|          | `=~` `!~`      |                           |
|     5    |  `not`         | Unary logical expression  |
|     6    |  `and`         |       Logical AND         |
|     7    |   `or`         |       Logical OR          |
|     8    | `if/then/else` |  conditional expression   |

Within the grammar itself, precedence is reversed so lower precedence operators appear above higher precedence operators. This ensures that the higher precedence values are nested within lower precedence operators.
