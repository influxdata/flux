#include <stdio.h>
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
        regex_lit => { s->token = REGEX; fbreak; };

        # We have to specify whitespace here so that leading whitespace doesn't cause a state transition.
        whitespace+;

        # Any other character we transfer to the main state machine that defines the entire language.
        any => { fhold; fgoto main; };
    *|;

    # This machine does not contain the regex literal.
    main := |*
        single_line_comment => { s->token = COMMENT; fbreak; };

        "and" => { s->token = AND; fbreak; };
        "or" => { s->token = OR; fbreak; };
        "not" => { s->token = NOT; fbreak; };
        "empty" => { s->token = EMPTY; fbreak; };
        "in" => { s->token = IN; fbreak; };
        "import" => { s->token = IMPORT; fbreak; };
        "package" => { s->token = PACKAGE; fbreak; };
        "return" => { s->token = RETURN; fbreak; };
        "option" => { s->token = OPTION; fbreak; };
        "builtin" => { s->token = BUILTIN; fbreak; };
        "test" => { s->token = TEST; fbreak; };
        "if" => { s->token = IF; fbreak; };
        "then" => { s->token = THEN; fbreak; };
        "else" => { s->token = ELSE; fbreak; };

        identifier => { s->token = IDENT; fbreak; };
        int_lit => { s->token = INT; fbreak; };
        float_lit => { s->token = FLOAT; fbreak; };
        duration_lit => { s->token = DURATION; fbreak; };
        date_time_lit => { s->token = TIME; fbreak; };
        string_lit => { s->token = STRING; fbreak; };

        "+" => { s->token = ADD; fbreak; };
        "-" => { s->token = SUB; fbreak; };
        "*" => { s->token = MUL; fbreak; };
        "/" => { s->token = DIV; fbreak; };
        "%" => { s->token = MOD; fbreak; };
        "==" => { s->token = EQ; fbreak; };
        "<" => { s->token = LT; fbreak; };
        ">" => { s->token = GT; fbreak; };
        "<=" => { s->token = LTE; fbreak; };
        ">=" => { s->token = GTE; fbreak; };
        "!=" => { s->token = NEQ; fbreak; };
        "=~" => { s->token = REGEXEQ; fbreak; };
        "!~" => { s->token = REGEXNEQ; fbreak; };
        "=" => { s->token = ASSIGN; fbreak; };
        "=>" => { s->token = ARROW; fbreak; };
        "<-" => { s->token = PIPE_RECEIVE; fbreak; };
        "(" => { s->token = LPAREN; fbreak; };
        ")" => { s->token = RPAREN; fbreak; };
        "[" => { s->token = LBRACK; fbreak; };
        "]" => { s->token = RBRACK; fbreak; };
        "{" => { s->token = LBRACE; fbreak; };
        "}" => { s->token = RBRACE; fbreak; };
        ":" => { s->token = COLON; fbreak; };
        "|>" => { s->token = PIPE_FORWARD; fbreak; };
        "," => { s->token = COMMA; fbreak; };
        "." => { s->token = DOT; fbreak; };

        whitespace+;
    *|;
}%%

%% write data;

// Scanner is used to tokenize Flux source.
struct scanner_t {
    char* p;
    char* pe;
    char* eof;
    char* ts;
    char* te;
    int token;
};

void init(struct scanner_t *s, char* data) {
    s->p = data;
    s->pe = data + strlen(data);
    s->eof = s->pe;
    s->ts = 0;
    s->te = 0;
    s->token = 0;
    printf("init '%s' %d %p %p\n", s->p, (int)strlen(data), (void *)s->p, (void *)s->pe);
}

void _scan(struct scanner_t *s, int cs) {
    printf("_scan start %p %p '%s'\n", (void *)(s->p), (void *)(s->pe), s->p);
    %% variable p s->p;
    %% variable pe s->pe;
    %% variable eof s->eof;
    %% variable ts s->ts;
    %% variable te s->te;

    int act;

    %% write init nocs;
    %% write exec;
    printf("_scan stop %d %.*s\n", s->token, (int)(s->te - s->ts), s->ts);
}

void scan(struct scanner_t *s) {
    _scan(s, flux_en_main);
}

void scan_with_regex(struct scanner_t *s) {
    _scan(s, flux_en_main_with_regex);
}
