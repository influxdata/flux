package values

import (
	"io"
	"regexp"

	"github.com/influxdata/flux/semantic"
)

const (
	// Stream types
	ReadOnly      = "ro"
	WriteOnly     = "wo"
	ReadWrite     = "rw"
	ReadSeek      = "rs"
	WriteSeek     = "ws"
	ReadWriteSeek = "rws"
)

type Stream interface {
	Value
}

type stream struct {
	t semantic.Type
}

func (s stream) Type() semantic.Type {
	return s.t
}

func (s stream) PolyType() semantic.PolyType {
	return s.t.PolyType()
}

func (s stream) IsNull() bool {
	return false
}

func (s stream) Str() string {
	panic(UnexpectedKind(semantic.Object, semantic.String))
}

func (s stream) Int() int64 {
	panic(UnexpectedKind(semantic.Object, semantic.Int))
}

func (s stream) UInt() uint64 {
	panic(UnexpectedKind(semantic.Object, semantic.UInt))
}

func (s stream) Float() float64 {
	panic(UnexpectedKind(semantic.Object, semantic.Float))
}

func (s stream) Bool() bool {
	panic(UnexpectedKind(semantic.Object, semantic.Bool))
}

func (s stream) Time() Time {
	panic(UnexpectedKind(semantic.Object, semantic.Time))
}

func (s stream) Duration() Duration {
	panic(UnexpectedKind(semantic.Object, semantic.Duration))
}

func (s stream) Regexp() *regexp.Regexp {
	panic(UnexpectedKind(semantic.Object, semantic.Regexp))
}

func (s stream) Array() Array {
	panic(UnexpectedKind(semantic.Object, semantic.Array))
}

func (s stream) Object() Object {
	panic(UnexpectedKind(semantic.Object, semantic.Object))
}

func (s stream) Function() Function {
	panic(UnexpectedKind(semantic.Object, semantic.Function))
}

func (s stream) Stream() Stream {
	return s
}

func (s stream) Equal(rhs Value) bool {
	if s.Type() != rhs.Type() {
		return false
	}
	v, ok := rhs.(stream)
	return ok && (s == v)
}

type readStream struct {
	stream
	r io.Reader
}

func NewReadStream(r io.Reader) Stream {
	return &readStream{
		stream: stream{semantic.Stream},
		r:      r,
	}
}

func (s readStream) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}

type readSeekStream struct {
	stream
	r io.ReadSeeker
}

func NewReadSeekStream(r io.ReadSeeker) Stream {
	return &readSeekStream{
		stream: stream{semantic.Stream},
		r:      r,
	}
}

func (s readSeekStream) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}

func (s readSeekStream) Seek(offset int64, whence int) (int64, error) {
	return s.r.Seek(offset, whence)
}

type writeStream struct {
	stream
	r io.Writer
}

func NewWriteStream(r io.Writer) Stream {
	return &writeStream{
		stream: stream{semantic.Stream},
		r:      r,
	}
}

func (s writeStream) Write(p []byte) (n int, err error) {
	return s.r.Write(p)
}
