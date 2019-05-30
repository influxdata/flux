#include "scanner.h"

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

void _scan(int cs, const char **pp, const char *data, const char *pe, const char *eof, unsigned int *token, unsigned int *token_start, unsigned int *token_end) {
    const char *p = *pp;
    int act;
    const char *ts;
    const char *te;
    unsigned int tok;

    %% write init nocs;
    %% write exec;

    // Update output args
    *token = tok;
    *token_start = ts - data;
    *token_end = te - data;

    *pp = p;
}

void scan(const char **p, const char *data, const char *pe, const char *eof, unsigned int *token, unsigned int *token_start, unsigned int *token_end) {
    _scan(flux_en_main, p, data, pe, eof, token, token_start, token_end);
}

void scan_with_regex(const char **p, const char *data, const char *pe, const char *eof, unsigned int *token, unsigned int *token_start, unsigned int *token_end) {
    _scan(flux_en_main_with_regex, p, data, pe, eof, token, token_start, token_end);
}
