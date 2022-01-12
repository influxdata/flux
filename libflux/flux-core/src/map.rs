// Aliasing over map types makes it easy to switch them out across the crate to test alternatives

/// Hashmap type used with libflux
pub(crate) type HashMap<K, V> = std::collections::HashMap<K, V>;

/// Hashset type used with libflux
pub(crate) type HashSet<T> = std::collections::HashSet<T>;
