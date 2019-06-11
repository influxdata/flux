package strings

// Transformation functions
builtin title
builtin toUpper
builtin toLower
builtin trim
builtin trimPrefix
builtin trimSpace
builtin trimSuffix
builtin trimRight
builtin trimLeft
builtin toTitle
builtin hasPrefix
builtin hasSuffix
builtin containsStr
builtin containsAny
builtin equalFold
builtin compare
builtin countStr
builtin index
builtin indexAny
builtin lastIndex
builtin lastIndexAny
builtin isDigit
builtin isLetter
builtin isLower
builtin isUpper
builtin repeat
builtin replace
builtin replaceAll
builtin split
builtin splitAfter
builtin splitN
builtin splitAfterN
builtin joinStr

// hack to simulate an imported strings package
strings = {
  title:title,
  toUpper:toUpper,
  toLower:toLower,
  trim:trim,
  trimPrefix:trimPrefix,
  trimSpace:trimSpace,
  trimSuffix:trimSuffix,
  trimRight:trimRight,
  trimLeft:trimLeft,
  toTitle:toTitle,
  hasPrefix:hasPrefix,
  hasSuffix:hasSuffix,
  containsStr:containsStr,
  containsAny:containsAny,
  equalFold:equalFold,
  compare:compare,
  countStr:countStr,
  index:index,
  indexAny:indexAny,
  lastIndex:lastIndex,
  lastIndexAny:lastIndexAny,
  isDigit:isDigit,
  isLetter:isLetter,
  isLower:isLower,
  isUpper:isUpper,
  repeat:repeat,
  replace:replace,
  replaceAll:replaceAll,
  split:split,
  splitAfter:splitAfter,
  splitN:splitN,
  splitAfterN:splitAfterN,
  joinStr:joinStr,
}
