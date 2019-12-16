#include "scanner.h"

%%{
    machine flux;

    alphtype unsigned char;

    include WChar "unicode.rl";

    action advance_line {
        // We do this for every newline we find.
        // This allows us to return correct line/column for each token
        // back to the caller.
        (*cur_line)++;
        *last_newline = fpc + 1;
    }

    action advance_line_between_tokens {
        // We do this for each newline we find in the whitespace between tokens,
        // so we can record the location of the first byte of a token.
        last_newline_before_token = *last_newline;
        cur_line_token_start = *cur_line;
    }

    newline = '\n' @advance_line;
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

    escaped_char = "\\" ( "n" | "r" | "t" | "\\" | '"' | "${" );
    unicode_value = (any_count_line - [\\$]) | escaped_char;
    byte_value = "\\x" xdigit{2};
    dollar_value = "$" ( any_count_line - "{" );
    string_lit_char = ( unicode_value | byte_value | dollar_value );
    string_lit = '"' string_lit_char* "$"? :> '"';

    regex_escaped_char = "\\" ( "/" | "\\");
    regex_unicode_value = (any_count_line - "/") | regex_escaped_char;
    regex_lit = "/" ( regex_unicode_value | byte_value )+ "/";

    # The newline is optional so that a comment at the end of a file is considered valid.
    single_line_comment = "//" [^\n]* newline?;

    # Whitespace is standard ws and control codes->
    # (Note that newlines are handled separately; see notes above)
    whitespace = (space - '\n')+;

    # The regex literal is not compatible with division so we need two machines->
    # One machine contains the full grammar and is the main one, the other is used to scan when we are
    # in the middle of an expression and we are potentially expecting a division operator.
    main_with_regex := |*
        # If we see a regex literal, we accept that and do not go to the other scanner.
        regex_lit => { tok = REGEX; fbreak; };

        # We have to specify whitespace here so that leading whitespace doesn't cause a state transition.
        whitespace;

        newline => advance_line_between_tokens;

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
        "exists" => { tok = EXISTS; fbreak; };

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
        "^" => { tok = POW; fbreak; };
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
        '"' => { tok = QUOTE; fbreak; };

        whitespace;

        newline => advance_line_between_tokens;
    *|;

    # This is the scanner used when parsing a string expression.
    string_expr := |*
        "${" => { tok = STRINGEXPR; fbreak; };
        '"' => { tok = QUOTE; fbreak; };
        (string_lit_char - "\"")+ => { tok = TEXT; fbreak; };
    *|;
}%%

%% write data nofinal;

int scan(
    int mode,
    const unsigned char **pp,
    const unsigned char *data,
    const unsigned char *pe,
    const unsigned char *eof,
    const unsigned char **last_newline,
    unsigned int *cur_line,
    unsigned int *token,
    unsigned int *token_start,
    unsigned int *token_start_line,
    unsigned int *token_start_col,
    unsigned int *token_end,
    unsigned int *token_end_line,
    unsigned int *token_end_col
) {
    int cs = flux_start;
    switch (mode) {
    case 0:
        cs = flux_en_main;
        break;
    case 1:
        cs = flux_en_main_with_regex;
        break;
    case 2:
        cs = flux_en_string_expr;
        break;
    }
    const unsigned char *p = *pp;
    int act;
    const unsigned char *ts;
    const unsigned char *te;
    unsigned int tok = ILLEGAL;
    const unsigned char *last_newline_before_token = *last_newline;
    unsigned int cur_line_token_start = *cur_line;

    %% write init nocs;
    %% write exec;

    // Update output args.
    *token = tok;

    *token_start = ts - data;
    *token_start_line = cur_line_token_start;
    *token_start_col = ts - last_newline_before_token + 1;

    *token_end = te - data;

    if (*last_newline > te) {
        // te (the token end pointer) will only be less than last_newline
        // (pointer to the last newline the scanner saw) if we are trying
        // to find a multi-line token (either string or regex literal)
        // but don't find the closing `/` or `"`.
        // In that case we need to reset last_newline and cur_line.
        *cur_line = cur_line_token_start;
        *last_newline = last_newline_before_token;
    }

    *token_end_line = *cur_line;
    *token_end_col = te - *last_newline + 1;

    *pp = p;
    return cs == flux_error;
}
