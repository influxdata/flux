#[derive(Debug)]
pub struct ScopedVec<T> {
    elements: Vec<T>,
    scopes: Vec<usize>,
}

impl<T> Default for ScopedVec<T> {
    fn default() -> Self {
        Self::new()
    }
}

impl<T> ScopedVec<T> {
    pub fn new() -> Self {
        ScopedVec {
            elements: Vec::new(),
            scopes: Vec::new(),
        }
    }

    pub fn push(&mut self, element: T) {
        self.elements.push(element)
    }

    pub fn enter_scope(&mut self) {
        self.scopes.push(self.elements.len())
    }

    pub fn exit_scope(&mut self) -> std::iter::Rev<std::vec::Drain<'_, T>> {
        let start = self.scopes.pop().unwrap_or(0);
        self.elements.drain(start..).rev()
    }
}
