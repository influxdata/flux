package cmd

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/libflux/go/libflux"
)

const compileSuffix = ".fc"

func compilePackage(dir, pkgName string, astPkg *ast.Package) error {
	bs, err := json.Marshal(astPkg)
	if err != nil {
		return err
	}

	hdl, err := libflux.ParseJSON(bs)
	if err != nil {
		return err
	}

	pkg, err := libflux.Analyze(hdl)
	if err != nil {
		hdl.Free()
		return err
	}

	defer pkg.Free()
	bc, err := pkg.MarshalFB()
	if err != nil {
		return err
	}
	pkg.Free()

	if pkgDir := filepath.Dir(pkgName); pkgDir != "." {
		if err := os.MkdirAll(filepath.Join(dir, pkgDir), 0755); err != nil {
			return err
		}
	} else {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(bc); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(dir, pkgName+compileSuffix), buf.Bytes(), 0600)
}
