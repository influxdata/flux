//#include <stdio.h>
#include <string.h>

enum Token {
    ILLEGAL                       = 0,
    EOFTOK                        = 1,
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
};

%%{
    machine flux;

    alphtype unsigned char;

    include WChar "unicode.rl";

    newline = '\n';
    any_count_line = any | newline;

    identifier = ( ualpha | "_" ) ( ualnum | "_" )*;

    decimal_lit = (digit - "0") digit*;
    int_lit = "0" | decimal_lit;

    float_lit = (digit+ "." digit*) | ("." digit+);

    duration_unit = "y" | "mo" | "w" | "d" | "h" | "m" | "s" | "ms" | "us" | "µs" | "ns";
    duration_lit = ( int_lit duration_unit )+;

    date = digit{4} "-" digit{2} "-" digit{2};
    time_offset = "Z" | (("+" | "-") digit{2} ":" digit{2});
    time = digit{2} ":" digit{2} ":" digit{2} ( "." digit* )? time_offset?;
    date_time_lit = date ( "T" time )?;

    # todo(jsternberg): string expressions have to be included in the string literal.
    escaped_char = "\\" ( "n" | "r" | "t" | "\\" | '"' );
    unicode_value = (any_count_line - "\\") | escaped_char;
    byte_value = "\\x" xdigit{2};
    string_lit = '"' ( unicode_value | byte_value )* :> '"';

    regex_escaped_char = "\\" ( "/" | "\\");
    regex_unicode_value = (any_count_line - "/") | regex_escaped_char;
    regex_lit = "/" ( regex_unicode_value | byte_value )+ "/";

    # The newline is optional so that a comment at the end of a file is considered valid.
    single_line_comment = "//" [^\n]* newline?;

    # Whitespace is standard ws, newlines and control codes->
    whitespace = ( newline | space )+ ;

    # The regex literal is not compatible with division so we need two machines->
    # One machine contains the full grammar and is the main one, the other is used to scan when we are
    # in the middle of an expression and we are potentially expecting a division operator.
    main_with_regex := |*
        # If we see a regex literal, we accept that and do not go to the other scanner.
        regex_lit => { tok = REGEX; fbreak; };

        # We have to specify whitespace here so that leading whitespace doesn't cause a state transition.
        whitespace+;

        # Any other character we transfer to the main state machine that defines the entire language.
        any => { fhold; fgoto main; };
    *|;

    # This machine does not contain the regex literal.
    main := |*
        single_line_comment => { tok = COMMENT; fbreak; };

        "and" => { tok = AND; fbreak; };
        "or" => { tok = OR; fbreak; };
        "not" => { tok = NOT; fbreak; };
        "empty" => { tok = EMPTY; fbreak; };
        "in" => { tok = IN; fbreak; };
        "import" => { tok = IMPORT; fbreak; };
        "package" => { tok = PACKAGE; fbreak; };
        "return" => { tok = RETURN; fbreak; };
        "option" => { tok = OPTION; fbreak; };
        "builtin" => { tok = BUILTIN; fbreak; };
        "test" => { tok = TEST; fbreak; };
        "if" => { tok = IF; fbreak; };
        "then" => { tok = THEN; fbreak; };
        "else" => { tok = ELSE; fbreak; };

        identifier => { tok = IDENT; fbreak; };
        int_lit => { tok = INT; fbreak; };
        float_lit => { tok = FLOAT; fbreak; };
        duration_lit => { tok = DURATION; fbreak; };
        date_time_lit => { tok = TIME; fbreak; };
        string_lit => { tok = STRING; fbreak; };

        "+" => { tok = ADD; fbreak; };
        "-" => { tok = SUB; fbreak; };
        "*" => { tok = MUL; fbreak; };
        "/" => { tok = DIV; fbreak; };
        "%" => { tok = MOD; fbreak; };
        "==" => { tok = EQ; fbreak; };
        "<" => { tok = LT; fbreak; };
        ">" => { tok = GT; fbreak; };
        "<=" => { tok = LTE; fbreak; };
        ">=" => { tok = GTE; fbreak; };
        "!=" => { tok = NEQ; fbreak; };
        "=~" => { tok = REGEXEQ; fbreak; };
        "!~" => { tok = REGEXNEQ; fbreak; };
        "=" => { tok = ASSIGN; fbreak; };
        "=>" => { tok = ARROW; fbreak; };
        "<-" => { tok = PIPE_RECEIVE; fbreak; };
        "(" => { tok = LPAREN; fbreak; };
        ")" => { tok = RPAREN; fbreak; };
        "[" => { tok = LBRACK; fbreak; };
        "]" => { tok = RBRACK; fbreak; };
        "{" => { tok = LBRACE; fbreak; };
        "}" => { tok = RBRACE; fbreak; };
        ":" => { tok = COLON; fbreak; };
        "|>" => { tok = PIPE_FORWARD; fbreak; };
        "," => { tok = COMMA; fbreak; };
        "." => { tok = DOT; fbreak; };

        whitespace+;
    *|;
}%%

%% write data;

void _scan(int cs, char **pp, char *data, char *pe, char *eof, int *token, int *token_start, int *token_end) {
    char *p = *pp;
    int act;
    char *ts;
    char *te;
    int tok;

    %% write init nocs;
    %% write exec;

    // Update output args
    *token = tok;
    *token_start = ts - data;
    *token_end = te - data;

    *pp = p;

    //printf("_scan done %d '%.*s'\n",
    //    *token,
    //    te - ts,
    //    ts);
}

void scan(char **p, char *data, char *pe, char *eof, int *token, int *token_start, int *token_end) {
    _scan(flux_en_main, p, data, pe, eof, token, token_start, token_end);
}

