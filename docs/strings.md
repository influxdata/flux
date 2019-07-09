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