# Parser Grammar

The following is the EBNF grammar definition of Flux.

TODO(jsternberg): Move this to the spec. This information is in the spec, but it is spread out into different locations.

    Program = StatementList .
    StatementList = { Statement } .
    Statement = OptionStatement
              | VariableDeclaration
              | ReturnStatement
              | ExpressionStatement .
    OptionStatement = "option" VariableDeclaration .
    VariableDeclaration = ident assign Expression .
    ReturnStatement = return Expression .
    ExpressionStatement = Expression .
    Expression = LogicalExpression .
    LogicalExpression = ComparisonExpression
                      | LogicalExpression LogicalOperator ComparisonExpression .
    LogicalOperator = and | or .
    ComparisonExpression = MultiplicativeExpression
                         | ComparisonExpression ComparisonOperator MultiplicativeExpression .
    ComparisonOperator = eq | neq | regexeq | regexneq .
    MultiplicativeExpression = AdditiveExpression
                             | MultiplicativeExpression MultiplicativeOperator AdditiveExpression .
    MultiplicativeOperator = mul | div .
    AdditiveExpression = PipeExpression
                       | AdditiveExpression AdditiveOperator PipeExpression .
    AdditiveOperator = add | sub .
    PipeExpression = PostfixExpression
                   | PipeExpression PipeOperator PostfixExpression .
    PipeOperator = pipe_forward .
    PostfixExpression = UnaryExpression
                      | PostfixExpression PostfixOperator .
    PostfixOperator = DotExpression
                    | CallExpression
                    | IndexExpression .
    DotExpression = dot ident .
    CallExpression = lparen ParameterList rparen .
    IndexExpression = lbrack Expression rbrack .
    UnaryExpression = PrimaryExpression
                    | PrefixOperator PrimaryExpression .
    PrefixOperator = add | sub | not .
    PrimaryExpression = ident
                      | int
                      | float
                      | string
                      | regex
                      | duration
                      | pipe_receive
                      | ObjectLiteral
                      | ArrayLiteral
                      | ParenExpression
                      | ArrowExpression .
    ObjectLiteral = lbrace PropertyList rbrace .
    ArrayLiteral = lbrack ExpressionList rbrack .
    ParenExpression = lparen Expression rparen .
    ArrowExpression = lparen ParameterList rparen arrow ArrowBodyExpression .
    ArrowBodyExpression = BlockStatement | Expression .
    BlockStatement = lbrace StatementList rbrace .
    ExpressionList = [ Expression { comma Expression } ] .
    PropertyList = [ Property { comma Property } ] .
    Property = ident colon Expression .
    ParameterList = [ Parameter { comma Parameter } ] .
    Parameter = ident [ eq Expression ] .

Note: The "option" is an identifier with the literal "option". It will have the `ident` token type, but must contain that exact literal to follow that branch. It is possible for the "option" identifier to also be interpreted as a normal identifier as it is not a reserved keyword.

For the parser, the above grammar undergoes a process to have the left-recursion removed and is then left-factored to turn it into an LL(1) compliant grammar. For simplicity, an alternation operation will choose the first production that will accept the token type. This is because an existing production may be factored out for one token type, but may still exist in its current form for other token types when the first token in the production is an alternation over multiple terminals. To avoid creating more production rules that impact readability just to remove the now factored terminal, these are ignored.

    Program = StatementList .
    StatementList = { Statement } .
    Statement = OptionStatement
              | IdentStatement
              | ReturnStatement
              | ExpressionStatement .
    IdentStatement = ident ( AssignStatement | ExpressionSuffix ) .
    OptionStatement = "option" OptionDeclaration .
    OptionDeclaration = AssignStatement | VariableDeclaration | ExpressionSuffix
    VariableDeclaration = ident AssignStatement .
    AssignStatement = assign Expression .
    ReturnStatement = return Expression .
    ExpressionStatement = Expression .
    Expression = LogicalExpression .
    ExpressionSuffix = { PostfixOperator } { PipeExpressionSuffix } { AdditiveExpressionSuffix } { MultiplicativeExpressionSuffix } { ComparisonExpressionSuffix } { LogicalExpressionSuffix } .
    LogicalExpression = ComparisonExpression { LogicalExpressionSuffix } .
    LogicalExpressionSuffix = LogicalOperator ComparisonExpression .
    LogicalOperator = and | or .
    ComparisonExpression = MultiplicativeExpression { ComparisonExpressionSuffix } .
    ComparisonExpressionSuffix = ComparisonOperator MultiplicativeExpr .
    ComparisonOperator = eq | neq | lt | lte | gt | gte | regexeq | regexneq .
    MultiplicativeExpression = AdditiveExpression { MultiplicativeOperator AdditiveExpression } .
    MultiplicativeOperator = mul | div .
    AdditiveExpression = PipeExpression { AdditiveExpressionSuffix } .
    AdditiveExpressionSuffix = AdditiveOperator PipeExpression .
    AdditiveOperator = add | sub .
    PipeExpression = PostfixExpression { PipeExpressionSuffix } .
    PipeExpressionSuffix = PipeOperator PostfixExpression .
    PipeOperator = pipe_forward .
    PostfixExpression = UnaryExpression { PostfixOperator } .
    PostfixOperator = DotExpression
                    | CallExpression
                    | IndexExpression .
    DotExpression = dot ident .
    CallExpression = lparen ParameterList rparen .
    IndexExpression = lbrack Expression rbrack .
    UnaryExpression = PrimaryExpression
                    | PrefixOperator PrimaryExpression .
    PrefixOperator = add | sub | not .
    PrimaryExpression = ident
                      | int
                      | float
                      | string
                      | regex
                      | duration
                      | pipe_receive
                      | ObjectLiteral
                      | ArrayLiteral
                      | ParenExpression .
    ObjectLiteral = lbrace PropertyList rbrace .
    ArrayLiteral = lbrack ExpressionList rbrack .
    ParenExpression = lparen ParenExpressionBody .
    ParenExpressionBody = rparen ArrowExpression
                        | ident ParenIdentExpression
                        | Expression rparen .
    ParenIdentExpression = rparen [ ArrowExpression ]
                         | eq Expression [ comma ParameterList ] rparen ArrowExpression .
                         | comma ParameterList rparen ArrowExpression
                         | ExpressionSuffix rparen .
    ParenExpression = lparen Expression rparen .
    ArrowExpression = arrow ArrowBodyExpression .
    ArrowBodyExpression = BlockStatement | Expression .
    BlockStatement = lbrace StatementList rbrace .
    ExpressionList = [ Expression { comma Expression } ] .
    PropertyList = [ Property { comma Property } ] .
    Property = ident colon Expression .
    ParameterList = [ Parameter { comma Parameter } ] .
    Parameter = ident [ eq Expression ] .

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
