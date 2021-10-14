package execute

// Disposable is an interface to be implemented for a resource
// that will be disposed of at a defined time.
type Disposable interface {
	// Dispose is invoked when the resource will no longer be used.
	Dispose()
}
