package dict


// fromList will convert an array of key/value pairs
// into a dictionary.
builtin fromList : (pairs: [{key: K, value: V}]) => [K:V] where K: Comparable

// get will retrieve a value from a dictionary. If there is no
// key in the dictionary, the default value will be returned.
builtin get : (dict: [K:V], key: K, default: V) => V where K: Comparable

// insert will insert a key/value pair into the dictionary
// and return a new dictionary with that value inserted.
// If the key already exists in the dictionary, it will
// be overwritten.
builtin insert : (dict: [K:V], key: K, value: V) => [K:V] where K: Comparable

// remove will remove a key/value pair from the dictionary
// and return a new dictionary with that value removed.
builtin remove : (dict: [K:V], key: K) => [K:V] where K: Comparable
