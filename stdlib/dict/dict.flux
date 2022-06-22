// Package dict provides functions for interacting with dictionary types.
//
// ## Metadata
// introduced: 0.97.0
package dict


// fromList creates a dictionary from a list of records with `key` and `value`
// properties.
//
// ## Parameters
// - pairs: List of records with `key` and `value` properties.
//
// ## Examples
//
// ### Create a dictionary from a list of records
//
// ```no_run
// import "dict"
//
// d = dict.fromList(
//   pairs: [
//     {key: 1, value: "foo"},
//     {key: 2, value: "bar"}
//   ]
// )
//
// // Returns [1: "foo", 2: "bar"]
// ```
//
builtin fromList : (pairs: [{key: K, value: V}]) => [K:V] where K: Comparable

// get returns the value of a specified key in a dictionary or a default value
// if the key does not exist.
//
// ## Parameters
// - dict: Dictionary to return a value from.
// - key: Key to return from the dictionary.
// - default: Default value to return if the key does not exist in the
//   dictionary. Must be the same type as values in the dictionary.
//
// ## Examples
//
// ### Return a property of a dictionary
//
// ```no_run
// import "dict"
//
// d = [1: "foo", 2: "bar"]
//
// dict.get(
//   dict: d,
//   key: 1,
//   default: ""
// )
// // Returns "foo"
// ```
builtin get : (dict: [K:V], key: K, default: V) => V where K: Comparable

// insert inserts a key-value pair into a dictionary and returns a new,
// updated dictionary.
//
// If the key already exists in the dictionary, the function overwrites
// the existing value.
//
// ## Parameters
// - dict: Dictionary to update.
// - key: Key to insert into the dictionary.
//   Must be the same type as the existing keys in the dictionary.
// - value: Value to insert into the dictionary.
//   Must be the same type as the existing values in the dictionary.
//
// ## Examples
//
// ### Insert a new key-value pair into the a dictionary
//
// ```no_run
// import "dict"
//
// d = [1: "foo", 2: "bar"]
//
// dict.insert(
//   dict: d,
//   key: 3,
//   value: "baz"
// )
//
// // Returns [1: "foo", 2: "bar", 3: "baz"]
// ```
//
// ### Overwrite an existing key-value pair in a dictionary
//
// ```no_run
// import "dict"
//
// d = [1: "foo", 2: "bar"]
//
// dict.insert(
//   dict: d,
//   key: 2,
//   value: "baz"
// )
//
// // Returns [1: "foo", 2: "baz"]
// ```
builtin insert : (dict: [K:V], key: K, value: V) => [K:V] where K: Comparable

// remove removes a key value pair from a dictionary and returns an updated
// dictionary.
//
// ## Parameters
// - dict: Dictionary to remove the key-value pair from.
// - key: Key to remove from the dictionary.
//   Must be the same type as existing keys in the dictionary.
//
// ## Examples
//
// ### Remove a property from a dictionary
//
// ```no_run
// import "dict"
//
// d = [1: "foo", 2: "bar"]
//
// dict.remove(
//   dict: d,
//   key: 1
// )
//
// // Returns [2: "bar"]
// ```
builtin remove : (dict: [K:V], key: K) => [K:V] where K: Comparable
