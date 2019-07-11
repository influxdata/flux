
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
    POW                           = 29,
    EQ                            = 30,
    LT                            = 31,
    GT                            = 32,
    LTE                           = 33,
    GTE                           = 34,
    NEQ                           = 35,
    REGEXEQ                       = 36,
    REGEXNEQ                      = 37,
    ASSIGN                        = 38,
    ARROW                         = 39,
    LPAREN                        = 40,
    RPAREN                        = 41,
    LBRACK                        = 42,
    RBRACK                        = 43,
    LBRACE                        = 44,
    RBRACE                        = 45,
    COMMA                         = 46,
    DOT                           = 47,
    COLON                         = 48,
    PIPE_FORWARD                  = 49,
    PIPE_RECEIVE                  = 50,
    EXISTS                        = 51,
};


#define WASM_EXPORT __attribute__ ((visibility("default")))

// Scan reads the input and reports the next lexical token. Returns the execution state.
WASM_EXPORT int scan(int with_regex, const char **p, const char *data, const char *pe, const char *eof, unsigned int *token, unsigned int *token_start, unsigned int *token_end);
