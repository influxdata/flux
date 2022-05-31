# Label polymorphism (take 2)

Investigating https://github.com/influxdata/flux/issues/4740 I finally found one (minimized) issue that I can't figure out a simple fix for.

```flux
builtin contains: (value: B, x: [B]) => bool where B: Nullable
builtin tagValues: => (bucket: string, tag: A) => stream[{F with _value:G}]
    where A: Label

x = (value) => {
    y = tagValues(bucket: "a", tag: value)
    // string is not Label (argument value)
    x = contains(set: ["_stop", "_start"], value: value)
    return 0
}


```

Here the constraint `A: Label` from `tagValues` propagates to the type of `value`. In the contains call we then infer the variable `B` to be a `string` due to the `set` argument, as expected. It is the second argument where we pass `value` ends up causing a new error. Since the type of `value` is inferred to have the `Label` kind from the previous call we end up with this new error.

This is an error not with label polymorphism directly, but with the (attempted) sub typing I added between `label(A)` and `string` to allow string literals to continue to be passed to the affected functions.

As such I can think of two ways forward. We either fully implement subtyping to handle this case or we remove all subtyping, making `label(A)` its own distinct type (and syntax element, where I would propose `.label`).

## Implement sub typing

Type inference and sub typing generally do not mix very well so we will need to do some work, possibly a sizeable amount to implement it (my hope was that the limited hack I did would be enough to avoid full sub typing...). Looking at some papers [MLsub][] looks pretty promising, though it is only a few years old. The gist of it is that type inference works mostly the same, however it records the direction that a value moves instead of just doing equality constraints which allows sub-typing to work.

[MLsub]:https://www.repository.cam.ac.uk/bitstream/handle/1810/261583/Dolan_and_Mycrof-2017-POPL-AM.pdf

## Remove sub typing

Removing the `string <: label(A)` sub typing simply removes this issue directly, however it creates new ones. With string literals no longer being labels we can no longer update the type signatures without causing breakages for existing users.

* We could create distinct functions when updating a function to use labels.
  \+ Simple
  \- Users are unlikely to change to the new signatures, so they won't see any benefit
  \- Complicates documentation as there are now two similar functions

* We could pre-process the semantic graph and replace string literals with their label equivalent when they are used in a label position (`tagValues(bucket: "a", tag: "mylabel") => tagValues(bucket: "a", tag: .mylabel)`. This would not take type inference into account so it avoids the main issue brought up here, however that also means that the cases it can handle is much more limited, we couldn't handle a string literal being assigned to a variable before being used as a label.

```flux
tag = "mylabel" // Can't rewrite this to tag = .mylabel
tagValues(bucket: "a", tag: tag)
x = tag + "2"
```

* More alternatives?
