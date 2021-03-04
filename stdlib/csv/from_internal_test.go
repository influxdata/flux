package csv

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"
)

func TestSkipBOMReader(t *testing.T) {
	testCases := []struct {
		name string
		in   io.ReadCloser
		want []byte
		err  error
	}{
		{
			name: "no BOM",
			in:   ioutil.NopCloser(bytes.NewReader([]byte("hello world"))),
			want: []byte{104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100},
		},
		{
			name: "has BOM",
			in:   ioutil.NopCloser(bytes.NewReader([]byte{0xEF, 0xBB, 0xBF, 104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100})),
			want: []byte{104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100},
		},
		{
			name: "empty",
			in:   ioutil.NopCloser(bytes.NewReader([]byte{})),
			want: []byte{},
		},
		{
			name: "short",
			in:   ioutil.NopCloser(bytes.NewReader([]byte{1, 2})),
			want: []byte{1, 2},
		},
		{
			name: "error",
			in:   ioutil.NopCloser(errReader{err: errors.New("test error")}),
			err:  errors.New("test error"),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := ioutil.ReadAll(newSkipBOMReader(tc.in))
			if !bytes.Equal(tc.want, got) {
				t.Errorf("unequal bytes \nwant:\n%v\ngot: %v\n", tc.want, got)
			}
			if tc.err != nil {
				wantErr := tc.err.Error()
				if err != nil && wantErr != err.Error() {
					t.Errorf("unexpected error want:%q got %q", wantErr, err.Error())
				} else if err == nil {
					t.Errorf("expected error: %v", wantErr)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})

	}
}

type errReader struct {
	err error
}

func (e errReader) Read(_ []byte) (int, error) {
	return 0, e.err
}
