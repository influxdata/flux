use derive_more::Display;

/// Enum representing the possible types of tokens.
#[allow(missing_docs)] // Not documenting token type variants.
#[derive(Clone, Copy, Debug, Display, Hash, Eq, PartialEq)]
pub enum TokenType {
    #[display("ILLEGAL")]
    Illegal,
    #[display("EOF")]
    Eof,
    #[display("COMMENT")]
    Comment,

    // Reserved keywords
    #[display("AND")]
    And,
    #[display("OR")]
    Or,
    #[display("NOT")]
    Not,
    #[display("IMPORT")]
    Import,
    #[display("PACKAGE")]
    Package,
    #[display("RETURN")]
    Return,
    #[display("OPTION")]
    Option,
    #[display("BUILTIN")]
    Builtin,
    #[display("TESTCASE")]
    TestCase,
    #[display("IF")]
    If,
    #[display("THEN")]
    Then,
    #[display("ELSE")]
    Else,

    // Identifiers and literals
    #[display("IDENT")]
    Ident,
    #[display("INT")]
    Int,
    #[display("FLOAT")]
    Float,
    #[display("STRING")]
    String,
    #[display("REGEX")]
    Regex,
    #[display("TIME")]
    Time,
    #[display("DURATION")]
    Duration,

    // Operators
    #[display("ADD")]
    Add,
    #[display("SUB")]
    Sub,
    #[display("MUL")]
    Mul,
    #[display("DIV")]
    Div,
    #[display("MOD")]
    Mod,
    #[display("POW")]
    Pow,
    #[display("EQ")]
    Eq,
    #[display("LT")]
    Lt,
    #[display("GT")]
    Gt,
    #[display("LTE")]
    Lte,
    #[display("GTE")]
    Gte,
    #[display("NEQ")]
    Neq,
    #[display("REGEXEQ")]
    RegexEq,
    #[display("REGEXNEQ")]
    RegexNeq,
    #[display("ASSIGN")]
    Assign,
    #[display("ARROW")]
    Arrow,
    #[display("LPAREN")]
    LParen,
    #[display("RPAREN")]
    RParen,
    #[display("LBRACK")]
    LBrack,
    #[display("RBRACK")]
    RBrack,
    #[display("LBRACE")]
    LBrace,
    #[display("RBRACE")]
    RBrace,
    #[display("COMMA")]
    Comma,
    #[display("DOT")]
    Dot,
    #[display("COLON")]
    Colon,
    #[display("PIPE_FORWARD")]
    PipeForward,
    #[display("PIPE_RECEIVE")]
    PipeReceive,
    #[display("EXISTS")]
    Exists,

    // String expression tokens
    #[display("QUOTE")]
    Quote,
    #[display("STRINGEXPR")]
    StringExpr,
    #[display("TEXT")]
    Text,

    #[display("QUESTION_MARK")]
    QuestionMark,

    #[display("ATTRIBUTE")]
    Attribute,
}
