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

Documentation: https://docs.python.org/3/howto/unicode.html

### Go

Documentation/Source:
- https://golang.org/pkg/unicode/
- https://coderwall.com/p/k7zvyg/dealing-with-unicode-in-go

### Rust 

??

### Ruby

Documentation/Source: 
- https://blog.daftcode.pl/fixing-unicode-for-ruby-developers-60d7f6377388
- https://www.honeybadger.io/blog/ruby-s-unicode-support/

### mySQL

Documentation/Source: 
- https://dev.mysql.com/doc/refman/8.0/en/charset-unicode.html

### Oracle

Does support unicode.

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


