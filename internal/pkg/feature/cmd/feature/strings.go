// Sprig
// Copyright (C) 2013 Masterminds
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func quote(str ...interface{}) string {
	out := make([]string, 0, len(str))
	for _, s := range str {
		if s != nil {
			out = append(out, fmt.Sprintf("%q", strval(s)))
		}
	}
	return strings.Join(out, " ")
}

func strval(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func replace(old, new, src string) string {
	return strings.Replace(src, old, new, -1)
}

// The MIT License (MIT)
//
// Copyright (c) 2015 Huan Du
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// https://github.com/huandu/xstrings

// camelcase is to convert words separated by space, underscore and hyphen to camel case.
//
// Some samples.
//     "some_words"      => "SomeWords"
//     "http_server"     => "HttpServer"
//     "no_https"        => "NoHttps"
//     "_complex__case_" => "_Complex_Case_"
//     "some words"      => "SomeWords"
func camelcase(str string) string {
	if len(str) == 0 {
		return ""
	}

	var buf strings.Builder
	var r0, r1 rune
	var size int

	// leading connector will appear in output.
	for len(str) > 0 {
		r0, size = utf8.DecodeRuneInString(str)
		str = str[size:]

		if !isConnector(r0) {
			r0 = unicode.ToUpper(r0)
			break
		}

		buf.WriteRune(r0)
	}

	if len(str) == 0 {
		// A special case for a string contains only 1 rune.
		if size != 0 {
			buf.WriteRune(r0)
		}

		return buf.String()
	}

	for len(str) > 0 {
		r1 = r0
		r0, size = utf8.DecodeRuneInString(str)
		str = str[size:]

		if isConnector(r0) && isConnector(r1) {
			buf.WriteRune(r1)
			continue
		}

		if isConnector(r1) {
			r0 = unicode.ToUpper(r0)
		} else {
			r0 = unicode.ToLower(r0)
			buf.WriteRune(r1)
		}
	}

	buf.WriteRune(r0)
	return buf.String()
}

func isConnector(r rune) bool {
	return r == '-' || r == '_' || unicode.IsSpace(r)
}
