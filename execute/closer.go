package execute

// Closer is an interface to be implemented for a resource
// that will be closed at a defined time.
type Closer interface {
	// Close is invoked when the resource will no longer be used.
	Close() error
}

// Close is a convenience method that will take an error and a
// Closer. This will call the Close method on the Closer. If the
// error is nil, it will return any error from the Close method.
// If the error was not nil, it will return the error.
func Close(err error, c Closer) error {
	if e := c.Close(); e != nil && err == nil {
		err = e
	}
	return err
}
