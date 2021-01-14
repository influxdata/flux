
use std::vec::Vec;

use crate::scanner::*;

%%{
    machine flux;

    alphtype u8;

    include WChar "unicode.rl";

    action advance_line {
        // We do this for every newline we find.
        // This allows us to return correct line/column for each token
        // back to the caller.
        *cur_line += 1;
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
        regex_lit => { tok = TokenType::REGEX; fbreak; };

        # We have to specify whitespace here so that leading whitespace doesn't cause a state transition.
        whitespace;

        newline => advance_line_between_tokens;

        # Any other character we transfer to the main state machine that defines the entire language.
        any => { fhold; fgoto main; };
    *|;

    # This machine does not contain the regex literal.
    main := |*
        single_line_comment => { tok = TokenType::COMMENT; fbreak; };

        "and" => { tok = TokenType::AND; fbreak; };
        "or" => { tok = TokenType::OR; fbreak; };
        "not" => { tok = TokenType::NOT; fbreak; };
        "empty" => { tok = TokenType::EMPTY; fbreak; };
        "in" => { tok = TokenType::IN; fbreak; };
        "import" => { tok = TokenType::IMPORT; fbreak; };
        "package" => { tok = TokenType::PACKAGE; fbreak; };
        "return" => { tok = TokenType::RETURN; fbreak; };
        "option" => { tok = TokenType::OPTION; fbreak; };
        "builtin" => { tok = TokenType::BUILTIN; fbreak; };
        "testcase" => { tok = TokenType::TESTCASE; fbreak; };
        "test" => { tok = TokenType::TEST; fbreak; };
        "if" => { tok = TokenType::IF; fbreak; };
        "then" => { tok = TokenType::THEN; fbreak; };
        "else" => { tok = TokenType::ELSE; fbreak; };
        "exists" => { tok = TokenType::EXISTS; fbreak; };

        identifier => { tok = TokenType::IDENT; fbreak; };
        int_lit => { tok = TokenType::INT; fbreak; };
        float_lit => { tok = TokenType::FLOAT; fbreak; };
        duration_lit => { tok = TokenType::DURATION; fbreak; };
        date_time_lit => { tok = TokenType::TIME; fbreak; };
        string_lit => { tok = TokenType::STRING; fbreak; };

        "+" => { tok = TokenType::ADD; fbreak; };
        "-" => { tok = TokenType::SUB; fbreak; };
        "*" => { tok = TokenType::MUL; fbreak; };
        "/" => { tok = TokenType::DIV; fbreak; };
        "%" => { tok = TokenType::MOD; fbreak; };
        "^" => { tok = TokenType::POW; fbreak; };
        "==" => { tok = TokenType::EQ; fbreak; };
        "<" => { tok = TokenType::LT; fbreak; };
        ">" => { tok = TokenType::GT; fbreak; };
        "<=" => { tok = TokenType::LTE; fbreak; };
        ">=" => { tok = TokenType::GTE; fbreak; };
        "!=" => { tok = TokenType::NEQ; fbreak; };
        "=~" => { tok = TokenType::REGEXEQ; fbreak; };
        "!~" => { tok = TokenType::REGEXNEQ; fbreak; };
        "=" => { tok = TokenType::ASSIGN; fbreak; };
        "=>" => { tok = TokenType::ARROW; fbreak; };
        "<-" => { tok = TokenType::PIPE_RECEIVE; fbreak; };
        "(" => { tok = TokenType::LPAREN; fbreak; };
        ")" => { tok = TokenType::RPAREN; fbreak; };
        "[" => { tok = TokenType::LBRACK; fbreak; };
        "]" => { tok = TokenType::RBRACK; fbreak; };
        "{" => { tok = TokenType::LBRACE; fbreak; };
        "}" => { tok = TokenType::RBRACE; fbreak; };
        ":" => { tok = TokenType::COLON; fbreak; };
        "|>" => { tok = TokenType::PIPE_FORWARD; fbreak; };
        "," => { tok = TokenType::COMMA; fbreak; };
        "." => { tok = TokenType::DOT; fbreak; };
        '"' => { tok = TokenType::QUOTE; fbreak; };
        '?' => { tok = TokenType::QUESTION_MARK; fbreak; };

        whitespace;

        newline => advance_line_between_tokens;
    *|;

    # This is the scanner used when parsing a string expression.
    string_expr := |*
        "${" => { tok = TokenType::STRINGEXPR; fbreak; };
        '"' => { tok = TokenType::QUOTE; fbreak; };
        (string_lit_char - "\"")+ => { tok = TokenType::TEXT; fbreak; };
    *|;
}%%

%% write data nofinal;

pub fn scan(
    data: &[u8],
    mode: i32,
    pp: &mut i32,
    _data: i32,
    pe: i32,
    eof: i32,
    last_newline: &mut i32,
    cur_line: &mut i32,
    token: &mut TokenType,
    token_start: &mut i32,
    token_start_line: &mut i32,
    token_start_col: &mut i32,
    token_end: &mut i32,
    token_end_line: &mut i32,
    token_end_col: &mut i32 ) -> u32
{
    let mut cs = flux_start;
    match mode {
        0 => { cs = flux_en_main },
        1 => { cs = flux_en_main_with_regex },
        2 => { cs = flux_en_string_expr },
        _ => {},
    }
    let mut p: i32 = *pp;

    let mut act: i32 = 0;
    let mut ts: i32 = 0;
    let mut te: i32 = 0;
    let mut tok: TokenType = TokenType::ILLEGAL;

    let mut last_newline_before_token: i32 = *last_newline;
    let mut cur_line_token_start: i32 = *cur_line;

    // alskdfj
    %% write init nocs;
    %% write exec;

    // Update output args.
    *token = tok;

    *token_start = ts - _data;
    *token_start_line = cur_line_token_start;
    *token_start_col = ts - last_newline_before_token + 1;

    *token_end = te - _data;

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
    if cs == flux_error {
        return 1
    } else {
        return 0;
    }
}
