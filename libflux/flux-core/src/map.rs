// Aliasing over map types makes it easy to switch them out across the crate to test alternatives
use std::{borrow::Borrow, cmp::Ordering, hash::Hash};

/// Hashmap type used with libflux. Forces iteration to be deterministic (unless unordered
/// iteration is explicitly opted in to)
#[derive(Debug, Clone)]
pub struct HashMap<K, V>(std::collections::HashMap<K, V>);

impl<K, V> From<std::collections::HashMap<K, V>> for HashMap<K, V> {
    fn from(v: std::collections::HashMap<K, V>) -> Self {
        Self(v)
    }
}

impl<K, V> PartialEq for HashMap<K, V>
where
    K: Eq + Hash,
    V: PartialEq,
{
    fn eq(&self, other: &Self) -> bool {
        self.0 == other.0
    }
}

impl<K, V> Eq for HashMap<K, V>
where
    K: Eq + Hash,
    V: Eq,
{
}

impl<K, V> Default for HashMap<K, V> {
    fn default() -> Self {
        Self(std::collections::HashMap::default())
    }
}

impl<K, V> FromIterator<(K, V)> for HashMap<K, V>
where
    K: Eq + Hash,
{
    fn from_iter<I: IntoIterator<Item = (K, V)>>(iter: I) -> Self {
        Self(std::collections::HashMap::from_iter(iter))
    }
}

impl<K, V> HashMap<K, V> {
    pub fn new() -> Self {
        Self(std::collections::HashMap::new())
    }
}

impl<K, V> HashMap<K, V>
where
    K: Eq + Hash,
{
    pub fn get<Q>(&self, key: &Q) -> Option<&V>
    where
        K: Borrow<Q>,
        Q: ?Sized + Eq + Hash,
    {
        self.0.get(key)
    }

    pub fn contains_key<Q>(&self, key: &Q) -> bool
    where
        K: Borrow<Q>,
        Q: ?Sized + Eq + Hash,
    {
        self.0.contains_key(key)
    }

    pub fn insert(&mut self, key: K, value: V) -> Option<V> {
        self.0.insert(key, value)
    }

    pub fn remove<Q>(&mut self, key: &Q) -> Option<V>
    where
        K: Borrow<Q>,
        Q: ?Sized + Eq + Hash,
    {
        self.0.remove(key)
    }

    pub fn entry(&mut self, key: K) -> std::collections::hash_map::Entry<'_, K, V> {
        self.0.entry(key)
    }

    pub fn clear(&mut self) {
        self.0.clear()
    }

    pub fn iter_by(
        &self,
        mut compare: impl FnMut(&K, &K) -> Ordering,
    ) -> impl Iterator<Item = (&K, &V)> {
        let mut vec: Vec<_> = self.0.iter().collect();
        vec.sort_by(|(l, _), (r, _)| compare(l, r));
        vec.into_iter()
    }

    pub fn into_iter_by(
        self,
        mut compare: impl FnMut(&K, &K) -> Ordering,
    ) -> impl Iterator<Item = (K, V)> {
        let mut vec: Vec<_> = self.0.into_iter().collect();
        vec.sort_by(|(l, _), (r, _)| compare(l, r));
        vec.into_iter()
    }

    pub fn values_by(&self, compare: impl FnMut(&K, &K) -> Ordering) -> impl Iterator<Item = &V> {
        self.iter_by(compare).map(|(_, v)| v)
    }

    pub fn keys_by(&self, compare: impl FnMut(&K, &K) -> Ordering) -> impl Iterator<Item = &K> {
        self.iter_by(compare).map(|(k, _)| k)
    }

    pub fn unordered_iter(&self) -> std::collections::hash_map::Iter<'_, K, V> {
        self.0.iter()
    }

    pub fn unordered_into_iter(self) -> std::collections::hash_map::IntoIter<K, V> {
        self.0.into_iter()
    }

    pub fn unordered_values(&self) -> std::collections::hash_map::Values<'_, K, V> {
        self.0.values()
    }

    pub fn unordered_keys(&self) -> std::collections::hash_map::Keys<'_, K, V> {
        self.0.keys()
    }
}

impl<K, V> HashMap<K, V>
where
    K: Ord + Eq + Hash,
{
    pub fn iter(&self) -> impl Iterator<Item = (&K, &V)> {
        self.iter_by(std::cmp::Ord::cmp)
    }

    pub fn into_iter(self) -> impl Iterator<Item = (K, V)> {
        self.into_iter_by(std::cmp::Ord::cmp)
    }

    pub fn values(&self) -> impl Iterator<Item = &V> {
        self.values_by(std::cmp::Ord::cmp)
    }

    pub fn keys(&self) -> impl Iterator<Item = &K> {
        self.keys_by(std::cmp::Ord::cmp)
    }
}

/// Hashset type used with libflux. Forces iteration to be deterministic (unless unordered
/// iteration is explicitly opted in to)
#[derive(Debug, Clone)]
pub(crate) struct HashSet<T>(std::collections::HashSet<T>);

impl<T> Default for HashSet<T> {
    fn default() -> Self {
        Self(std::collections::HashSet::default())
    }
}

impl<T> HashSet<T> {
    pub fn new() -> Self {
        Self(std::collections::HashSet::new())
    }
}

impl<T> HashSet<T>
where
    T: Eq + Hash,
{
    pub fn contains<Q>(&self, key: &Q) -> bool
    where
        T: Borrow<Q>,
        Q: ?Sized + Eq + Hash,
    {
        self.0.contains(key)
    }

    pub fn insert(&mut self, key: T) -> bool {
        self.0.insert(key)
    }
}
