// Package dictionary provides functions for interacting with dictionary types.
package dict


// fromList is a function that creates a dictionary from a list of records
//  with key and value properties.
//
// ## Parameters
// - `pairs` is the list of records, each containing key and value properties.
//
// ## Create a dictionary from a list of records
//
// ```
// import "dict"
//
// // Define a new dictionary using an array of records
// d = dict.fromList(
//   pairs: [
//     {key: 1, value: "foo"},
//     {key: 2, value: "bar"}
//   ]
// )
//
// // Return a property of the dictionary
// dict.get(dict: d, key: 1, default: "") // returns foo
// ```
builtin fromList : (pairs: [{key: K, value: V}]) => [K:V] where K: Comparable

// get is a function that returns the value of a specified key in the
//  dictionary or a default value if the key does not exist.
//
// ## Parameters
// - `dict` is the dictionary to return a value from.
// - `key` is the key to return from the dictionary.
// - `default` is the default value to return if the key does not
//   exist in the dictionary.
//
//   Must be the same type as values in the dictionary.
//
// ## Return a property of a dictionary
//
// ```
// import "dict"
//
// d = [1: "foo", 2: "bar"]
//
// dict.get(
//   dict: d,
//   key: 1,
//   default: ""
// )
// // returns foo
// ```
builtin get : (dict: [K:V], key: K, default: V) => V where K: Comparable

// insert is a function that inserts a key-value pair into a dictionary and
//  returns a new, updated dictionary.
//
//  If the key already exists in the dictionary, the function overwrites
//  the existing value.
//
// ## Parameters
// - `dict` is the dictionary to update.
// - `key` is the key to insert into the dictionary.
//
//   Must be the same type as the existing keys in the dictionary.
//
// - `default` is the value to insert into the dictionary.
//
//   Must be the same type as the existing values in the dictionary. 
//
// ## Insert a new key-value pair into the a dictionary
//
// ```
// import "dict"
//
// d = [1: "foo", 2: "bar"]
//
// dNew = dict.insert(
//   dict: d,
//   key: 3,
//   value: "baz"
// )
//
// // Verify the new key-value pair was inserted
// dict.get(dict: dNew, key: 3, default: "")
// ```
//
// ## Overwrite an existing key-value pair in a dictionary
//
// ```
// import "dict"
//
// d = [1: "foo", 2: "bar"]
//
// dNew = dict.insert(
//   dict: d,
//   key: 2,
//   value: "baz"
// )
//
// // Verify the new key-value pair was overwritten
// dict.get(dict: dNew, key: 2, default: "")
// ```
builtin insert : (dict: [K:V], key: K, value: V) => [K:V] where K: Comparable

// remove is a function that removes a key value pair from a dictionary and
//  returns an updated dictionary. 
//
// ## Parameters
// - `dict` is the dictionary to remove the key-value pair from.
// - `key` is the key to remove from the dictionary.
//
//   Must be the same type as existing keys in the dictionary.
//
// ## Remove a property from a dictionary
//
// ```
// import "dict"
//
// d = [1: "foo", 2: "bar"]
//
// dNew = dict.remove(
//   dict: d,
//   key: 1
// )
//
// // Verify the key-value pairs was removed
//
// dict.get(dict: dNew, key: 1, default: "")
// // Returns an empty string
//
// dict.get(dict: dNew, key: 2, default: "")
// // Returns bar
// ```
builtin remove : (dict: [K:V], key: K) => [K:V] where K: Comparable
