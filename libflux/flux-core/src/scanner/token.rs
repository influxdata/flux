use derive_more::Display;

/// Enum representing the possible types of tokens.
#[allow(missing_docs)] // Not documenting token type variants.
#[derive(Clone, Copy, Debug, Display, Hash, Eq, PartialEq)]
pub enum TokenType {
    #[display(fmt = "ILLEGAL")]
    Illegal,
    #[display(fmt = "EOF")]
    Eof,
    #[display(fmt = "COMMENT")]
    Comment,

    // Reserved keywords
    #[display(fmt = "AND")]
    And,
    #[display(fmt = "OR")]
    Or,
    #[display(fmt = "NOT")]
    Not,
    #[display(fmt = "EMPTY")]
    Empty,
    #[display(fmt = "IN")]
    In,
    #[display(fmt = "IMPORT")]
    Import,
    #[display(fmt = "PACKAGE")]
    Package,
    #[display(fmt = "RETURN")]
    Return,
    #[display(fmt = "OPTION")]
    Option,
    #[display(fmt = "BUILTIN")]
    Builtin,
    #[display(fmt = "TEST")]
    Test,
    #[display(fmt = "TESTCASE")]
    TestCase,
    #[display(fmt = "IF")]
    If,
    #[display(fmt = "THEN")]
    Then,
    #[display(fmt = "ELSE")]
    Else,

    // Identifiers and literals
    #[display(fmt = "IDENT")]
    Ident,
    #[display(fmt = "INT")]
    Int,
    #[display(fmt = "FLOAT")]
    Float,
    #[display(fmt = "STRING")]
    String,
    #[display(fmt = "REGEX")]
    Regex,
    #[display(fmt = "TIME")]
    Time,
    #[display(fmt = "DURATION")]
    Duration,

    // Operators
    #[display(fmt = "ADD")]
    Add,
    #[display(fmt = "SUB")]
    Sub,
    #[display(fmt = "MUL")]
    Mul,
    #[display(fmt = "DIV")]
    Div,
    #[display(fmt = "MOD")]
    Mod,
    #[display(fmt = "POW")]
    Pow,
    #[display(fmt = "EQ")]
    Eq,
    #[display(fmt = "LT")]
    Lt,
    #[display(fmt = "GT")]
    Gt,
    #[display(fmt = "LTE")]
    Lte,
    #[display(fmt = "GTE")]
    Gte,
    #[display(fmt = "NEQ")]
    Neq,
    #[display(fmt = "REGEXEQ")]
    RegexEq,
    #[display(fmt = "REGEXNEQ")]
    RegexNeq,
    #[display(fmt = "ASSIGN")]
    Assign,
    #[display(fmt = "ARROW")]
    Arrow,
    #[display(fmt = "LPAREN")]
    LParen,
    #[display(fmt = "RPAREN")]
    RParen,
    #[display(fmt = "LBRACK")]
    LBrack,
    #[display(fmt = "RBRACK")]
    RBrack,
    #[display(fmt = "LBRACE")]
    LBrace,
    #[display(fmt = "RBRACE")]
    RBrace,
    #[display(fmt = "COMMA")]
    Comma,
    #[display(fmt = "DOT")]
    Dot,
    #[display(fmt = "COLON")]
    Colon,
    #[display(fmt = "PIPE_FORWARD")]
    PipeForward,
    #[display(fmt = "PIPE_RECEIVE")]
    PipeReceive,
    #[display(fmt = "EXISTS")]
    Exists,

    // String expression tokens
    #[display(fmt = "QUOTE")]
    Quote,
    #[display(fmt = "STRINGEXPR")]
    StringExpr,
    #[display(fmt = "TEXT")]
    Text,

    #[display(fmt = "QUESTION_MARK")]
    QuestionMark,
}
