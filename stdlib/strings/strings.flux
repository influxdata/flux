package strings

// Transformation functions
builtin title
builtin toUpper
builtin toLower
builtin trim
builtin trimPrefix
builtin trimSpace
builtin trimSuffix

// hack to simulate an imported strings package
strings = {
  title:title,
  toUpper:toUpper,
  toLower:toLower,
  trim:trim,
  trimPrefix:trimPrefix,
  trimSpace:trimSpace,
  trimSuffix:trimSuffix,
}
