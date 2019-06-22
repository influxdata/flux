# Unicode

This document details suggestions on how to handle unicode in Flux as well as how unicode is handled 
in selected query languages.

## Unicode Summary

UTF-8 is defined to encode code points in 1 to 4 bytes. From Wikipedia, "The first 128 characters 
(US-ASCII) need one byte. The next 1,920 characters need two bytes to encode, which covers the remainder 
of almost all Latin-script alphabets, and also Greek, Cyrillic, Coptic, Armenian, Hebrew, Arabic, Syriac, 
Thaana and N'Ko alphabets, as well as Combining Diacritical Marks. Three bytes are needed for characters 
in the rest of the Basic Multilingual Plane, which contains virtually all characters in common use 
including most Chinese, Japanese and Korean characters. Four bytes are needed for characters in the other 
planes of Unicode, which include less common CJK characters, various historic scripts, mathematical 
symbols, and emoji (pictographic symbols)."

## Functions and Functionality in Flux affected by unicode

### Length

Should we define length as the number of characters or the number of bytes?
(Will be the same for the 128 US-ASCII characters that we are used to, but will be different for other
languages that Flux may want to support.)

### Substring/Slicing/Indexing

Will probably be more convenient to define length as number of values in the resulting array 
when `strings.split(v: "...", t: "")` is applied. Alternatively, can define length of array 
and length of strings differently?

## How unicode is dealt with in other languages

### Python

Has `bytes.decode()` and `str.encode()` methods that translate unicode to string and vice versa. 
One-character unicode strings can also be created with the chr() built-in function.

`unicodedata.normalize()` converts strings to one of several normal forms, where letters followed
 by a combining character are replaced with single characters. normalize() can be used to perform
 string comparisons that won’t falsely report inequality if two strings use combining characters
 differently.

```
import unicodedata

def compare_strs(s1, s2):
    def NFD(s):
        return unicodedata.normalize('NFD', s)

    return NFD(s1) == NFD(s2)

single_char = 'ê'
multiple_chars = '\N{LATIN SMALL LETTER E}\N{COMBINING CIRCUMFLEX ACCENT}'
print('length of first string=', len(single_char))
print('length of second string=', len(multiple_chars))
print(compare_strs(single_char, multiple_chars))
```
Returns:
```
length of first string= 1
length of second string= 2
True
```

Documentation: https://docs.python.org/3/howto/unicode.html

### Go

Go encodes string as byte array and thus the `length` is determined by the number of bytes 
instead of the number of characters.

Functions such as `utf8.DecodeRuneInString` allow users to print out how many bytes are in each character.

```
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	str := "Hello, 世界"

	for len(str) > 0 {
		r, size := utf8.DecodeRuneInString(str)
		fmt.Printf("%c %v\n", r, size)

		str = str[size:]
	}
}

```
Returns
```
H 1
e 1
l 1
l 1
o 1
, 1
  1
世 3
界 3
```

Go does support getting the length of a string via the number of runes/characters instead of the number 
of bytes using `utf8.RuneCountInString`.

Documentation/Source:
- https://golang.org/pkg/unicode/
- https://coderwall.com/p/k7zvyg/dealing-with-unicode-in-go
- https://golang.org/pkg/unicode/utf8/

### Ruby

Tracks `length` by number of bytes. Apparently does not support unicode very well. (Seems like a 
messed up version of Python's `normalize()` somewhere in the background?)

Documentation/Source: 
- https://blog.daftcode.pl/fixing-unicode-for-ruby-developers-60d7f6377388
- https://www.honeybadger.io/blog/ruby-s-unicode-support/

### mySQL

MySQL supports these Unicode character sets:
- utf8mb4: A UTF-8 encoding of the Unicode character set using one to four bytes per character.
- utf8mb3: A UTF-8 encoding of the Unicode character set using one to three bytes per character.
- utf8: An alias for utf8mb3.
- ucs2: The UCS-2 encoding of the Unicode character set using two bytes per character.
- utf16: The UTF-16 encoding for the Unicode character set using two or four bytes per character. Like ucs2 but with an extension for supplementary characters.
- utf16le: The UTF-16LE encoding for the Unicode character set. Like utf16 but little-endian rather than big-endian.
- utf32: The UTF-32 encoding for the Unicode character set using four bytes per character.

Documentation/Source: 
- https://dev.mysql.com/doc/refman/8.0/en/charset-unicode.html
- https://dev.mysql.com/doc/refman/8.0/en/charset-unicode-utf8mb4.html

### Oracle

Does support unicode.

The `NCHAR` datatype stores data encoded as Unicode. The column length specified for the `NCHAR` and 
`NVARCHAR2` columns is always the number of characters instead of the number of bytes.

"The lengths of the SQL `NCHAR` datatypes are defined as number of characters. This is the same as 
the way they are treated when using wchar_t strings in Windows C/C++ programs. This reduces programming 
complexity."

Documentation: https://docs.oracle.com/cd/B19306_01/server.102/b14225/ch6unicode.htm

### MongoDB

Has functions to deal by both bytes and by code points. For those functions that work through bytes, returns
an error if user requests something that does not split a character evenly according to its number of bytes.

Example: `{ $substrBytes: [ "cafétéria", 7, 3 ] }` \
Result: `"Error: Invalid range, starting index is a UTF-8 continuation byte."`

Documentation/Source:
- https://docs.mongodb.com/manual/reference/operator/aggregation/strLenBytes/index.html
- https://docs.mongodb.com/manual/reference/operator/aggregation/strLenCP/index.html
- https://docs.mongodb.com/manual/reference/operator/aggregation/substrBytes/index.html
- https://docs.mongodb.com/manual/reference/operator/aggregation/substrCP/index.html
- https://docs.mongodb.com/manual/reference/operator/aggregation/indexOfBytes/index.html
- https://docs.mongodb.com/manual/reference/operator/aggregation/indexOfCP/index.html


