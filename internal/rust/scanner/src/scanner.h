
// T represents all possible Tokens
enum T {
    ILLEGAL                       = 0,
    EOF                           = 1,
    COMMENT                       = 2,

    // Reserved keywords->
    AND                           = 3,
    OR                            = 4,
    NOT                           = 5,
    EMPTY                         = 6,
    IN                            = 7,
    IMPORT                        = 8,
    PACKAGE                       = 9,
    RETURN                        = 10,
    OPTION                        = 11,
    BUILTIN                       = 12,
    TEST                          = 13,
    IF                            = 14,
    THEN                          = 15,
    ELSE                          = 16,

    // Identifiers and literals->
    IDENT                         = 17,
    INT                           = 18,
    FLOAT                         = 19,
    STRING                        = 20,
    REGEX                         = 21,
    TIME                          = 22,
    DURATION                      = 23,

    // Operators->
    ADD                           = 24,
    SUB                           = 25,
    MUL                           = 26,
    DIV                           = 27,
    MOD                           = 28,
    EQ                            = 29,
    LT                            = 30,
    GT                            = 31,
    LTE                           = 32,
    GTE                           = 33,
    NEQ                           = 34,
    REGEXEQ                       = 35,
    REGEXNEQ                      = 36,
    ASSIGN                        = 37,
    ARROW                         = 38,
    LPAREN                        = 39,
    RPAREN                        = 40,
    LBRACK                        = 41,
    RBRACK                        = 42,
    LBRACE                        = 43,
    RBRACE                        = 44,
    COMMA                         = 45,
    DOT                           = 46,
    COLON                         = 47,
    PIPE_FORWARD                  = 48,
    PIPE_RECEIVE                  = 49,
    EXISTS                        = 50,
};


#define WASM_EXPORT __attribute__ ((visibility("default")))

// Scan reads the input and reports the next lexical token. Returns the execution state.
WASM_EXPORT int scan(int with_regex, const char **p, const char *data, const char *pe, const char *eof, unsigned int *token, unsigned int *token_start, unsigned int *token_end);
