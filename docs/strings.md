Discussion of possible functions to add to supplement and enhance the flux `strings` library.

### Ruby 
Ruby String Library Doc: https://ruby-doc.org/core-2.3.0/String.html

Ruby String Methods that Flux does not have and seems useful to have something similar:
- \#casecmp : Case-insensitive version of compare
- \#eql? : Looks convenient to have: easily implemented using equal = (v, t) => compare(v, t) == 0
- \#gsub : String substitution via regex
- \#insert : Insert string at desired index: can be implemented via replace?
- \#partition : Having split taking in regex as sep might be useful. Also consider splitting on regex.
- \#reverse : Reverse string order. Interesting. Unsure if useful?
- \#squeeze : Gets rid of repeated characters (shoot => shot). Looks somewhat dangerous if not used carefully but might be useful?

### Python and R 

Some top languages/software that people use in data analysis include Python, RapidMiner, and R, according to the [2019 Software Poll](https://www.kdnuggets.com/2019/05/poll-top-data-science-machine-learning-platforms.html) by KDNuggets. 

Specifically with regards to data manipulation and data cleaning, KDNuggets had a [poll](https://www.kdnuggets.com/polls/2008/tools-languages-used-data-cleaning.htm) from 2008, which ranked SQL as the top choice; however, this ranking may be outdated. 

The overall consensus of popular tools for data cleaning seems to be that Python and R are the best. Many of the reasons are reflected in this [post](https://www.quora.com/What-are-the-best-languages-and-libraries-for-cleaning-data) and this [article](https://www.newgenapps.com/blog/6-reasons-why-choose-r-programming-for-data-science-projects). As a summary, Python is appreciated for the ease of use and multitude of libraries (numpy, pandas, scipy, etc.). R similarly has popular packages (dyplr, data.table, etc.). 

**String Ideas from Python** 

Series.str.startswith(pat[, na]) | Test if the start of each string element matches a pattern.
- hasPrefix and hasSuffix take in regex

Is there a Flux function that fills in N/A values? 
Add a function specifically parsing date/time? 

**String Ideas from R**

Sorting Strings? But that'd mostly only be useful if we are working with the more than one datapoint/value.

**Documentations for Reference**
- https://pandas.pydata.org/pandas-docs/stable/reference/series.html
- https://github.com/rstudio/cheatsheets/blob/master/strings.pdf

### OpenRefine

OpenRefine (previously known as Google Refine) 

OpenRefine [String Documentation](https://github.com/OpenRefine/OpenRefine/wiki/GREL-String-Functions)

Many of the functions that seem useful have also been mentioned above in the "Ruby String Library" comment

The following functions are worth considering:

splitByLengths(string s, number n1, number n2, ...)
Returns the array of strings obtained by splitting s into substrings with the given lengths. For example, `splitByLengths("internationalization", 5, 6, 3)` returns an array of 3 strings: `inter`, `nation`, and `ali`.

### SQL

**String Ideas from SQL**

Given an array of strings, return the array with no repeats? (similar to Java's set)
Can have a `difference` function that returns the array of indices where the strings are different OR return the edit distance. Might be useful if someone is trying to see which values are actually the same, just entered differently

Source: https://www.sqlshack.com/sql-string-functions-for-data-munging-wrangling/

SideNote: Was there any toString method? 




## Unicode

This document details how unicode is handled in Flux as well as how unicode is handled 
in selected query languages.

### Unicode Summary

UTF-8 is defined to encode code points in 1 to 4 bytes. From Wikipedia, "The first 128 characters 
(US-ASCII) need one byte. The next 1,920 characters need two bytes to encode, which covers the remainder 
of almost all Latin-script alphabets, and also Greek, Cyrillic, Coptic, Armenian, Hebrew, Arabic, Syriac, 
Thaana and N'Ko alphabets, as well as Combining Diacritical Marks. Three bytes are needed for characters 
in the rest of the Basic Multilingual Plane, which contains virtually all characters in common use 
including most Chinese, Japanese and Korean characters. Four bytes are needed for characters in the other 
planes of Unicode, which include less common CJK characters, various historic scripts, mathematical 
symbols, and emoji (pictographic symbols)."

### Functions and Functionality in Flux affected by unicode

#### Length

Defined length as the number of characters where a character is defined to be an unicode code point.

#### Substring/Slicing/Indexing

Substring, slicing, and indexing are done based on the number of characters, not the number of bytes.
Substring in Flux is implemented using string slicing available in Go. 

### How unicode is dealt with in other languages

#### Python

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

#### Go

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

#### Ruby

Tracks `length` by number of bytes. Apparently does not support unicode very well. (Seems like a 
messed up version of Python's `normalize()` somewhere in the background?)

Documentation/Source: 
- https://blog.daftcode.pl/fixing-unicode-for-ruby-developers-60d7f6377388
- https://www.honeybadger.io/blog/ruby-s-unicode-support/

#### mySQL

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

#### Oracle

Does support unicode.

The `NCHAR` datatype stores data encoded as Unicode. The column length specified for the `NCHAR` and 
`NVARCHAR2` columns is always the number of characters instead of the number of bytes.

"The lengths of the SQL `NCHAR` datatypes are defined as number of characters. This is the same as 
the way they are treated when using wchar_t strings in Windows C/C++ programs. This reduces programming 
complexity."

Documentation: https://docs.oracle.com/cd/B19306_01/server.102/b14225/ch6unicode.htm

#### MongoDB

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


