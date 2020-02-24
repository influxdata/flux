// Package repl implements the read-eval-print-loop for the command line flux query console.
package repl

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/c-bata/go-prompt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type REPL struct {
	scope   values.Scope
	querier Querier
	ctx     context.Context
	deps    flux.Dependencies

	cancelMu   sync.Mutex
	cancelFunc context.CancelFunc
}

type Querier interface {
	Query(ctx context.Context, deps flux.Dependencies, compiler flux.Compiler) (flux.ResultIterator, error)
}

func New(ctx context.Context, deps flux.Dependencies, q Querier) *REPL {
	return &REPL{
		scope:   values.NewScope(),
		querier: q,
		ctx:     ctx,
		deps:    deps,
	}
}

func (r *REPL) Run() {
	p := prompt.New(
		r.input,
		r.completer,
		prompt.OptionPrefix("> "),
		prompt.OptionTitle("flux"),
	)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	go func() {
		for range sigs {
			r.cancel()
		}
	}()
	p.Run()
}

func (r *REPL) cancel() {
	r.cancelMu.Lock()
	defer r.cancelMu.Unlock()
	if r.cancelFunc != nil {
		r.cancelFunc()
		r.cancelFunc = nil
	}
}

func (r *REPL) setCancel(cf context.CancelFunc) {
	r.cancelMu.Lock()
	defer r.cancelMu.Unlock()
	r.cancelFunc = cf
}
func (r *REPL) clearCancel() {
	r.setCancel(nil)
}

func (r *REPL) completer(d prompt.Document) []prompt.Suggest {
	names := make([]string, 0, r.scope.Size())
	r.scope.Range(func(k string, v values.Value) {
		names = append(names, k)
	})
	sort.Strings(names)

	s := make([]prompt.Suggest, 0, len(names))
	for _, n := range names {
		if n == "_" || !strings.HasPrefix(n, "_") {
			s = append(s, prompt.Suggest{Text: n})
		}
	}
	if d.Text == "" || strings.HasPrefix(d.Text, "@") {
		root := "./" + strings.TrimPrefix(d.Text, "@")
		fluxFiles, err := getFluxFiles(root)
		if err == nil {
			for _, fName := range fluxFiles {
				s = append(s, prompt.Suggest{Text: "@" + fName})
			}
		}
		dirs, err := getDirs(root)
		if err == nil {
			for _, fName := range dirs {
				s = append(s, prompt.Suggest{Text: "@" + fName + string(os.PathSeparator)})
			}
		}
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func (r *REPL) Input(t string) error {
	return r.executeLine(t)
}

// input processes a line of input and prints the result.
func (r *REPL) input(t string) {
	if err := r.executeLine(t); err != nil {
		fmt.Println("Error:", err)
	}
}

// executeLine processes a line of input.
// If the input evaluates to a valid value, that value is returned.
func (r *REPL) executeLine(t string) error {
	if t == "" {
		return nil
	}

	if t[0] == '@' {
		q, err := LoadQuery(t)
		if err != nil {
			return err
		}
		t = q
	}

	ses, scope, err := runtime.Eval(r.ctx, t, func(ns values.Scope) {
		// copy values saved in the cached scope to the new interpreter's scope
		r.scope.Range(func(k string, v values.Value) {
			ns.Set(k, v)
		})
	})
	if err != nil {
		return err
	}
	r.scope = scope

	for _, se := range ses {
		if _, ok := se.Node.(*semantic.ExpressionStatement); ok {
			if t, ok := se.Value.(*flux.TableObject); ok {
				now, ok := r.scope.Lookup("now")
				if !ok {
					return fmt.Errorf("now option not set")
				}
				ctx := r.deps.Inject(context.TODO())
				nowTime, err := now.Function().Call(ctx, nil)
				if err != nil {
					return err
				}
				s, err := spec.FromTableObject(r.ctx, t, nowTime.Time().Time())
				if err != nil {
					return err
				}
				if err := r.doQuery(r.ctx, s, r.deps); err != nil {
					return err
				}
			} else {
				fmt.Println(se.Value)
			}
		}
	}
	return nil
}

func (r *REPL) doQuery(cx context.Context, spec *flux.Spec, deps flux.Dependencies) error {
	// Setup cancel context
	ctx, cancelFunc := context.WithCancel(cx)
	r.setCancel(cancelFunc)
	defer cancelFunc()
	defer r.clearCancel()

	replCompiler := Compiler{
		Spec: spec,
	}

	results, err := r.querier.Query(ctx, deps, replCompiler)
	if err != nil {
		return err
	}
	defer results.Release()

	for results.More() {
		result := results.Next()
		tables := result.Tables()
		fmt.Println("Result:", result.Name())
		err := tables.Do(func(tbl flux.Table) error {
			_, err := execute.NewFormatter(tbl, nil).WriteTo(os.Stdout)
			return err
		})
		if err != nil {
			return err
		}
	}
	return results.Err()
}

func getFluxFiles(path string) ([]string, error) {
	return filepath.Glob(path + "*.flux")
}

func getDirs(path string) ([]string, error) {
	dir := filepath.Dir(path)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, filepath.Join(dir, f.Name()))
		}
	}
	return dirs, nil
}

// LoadQuery returns the Flux query q, except for two special cases:
// if q is exactly "-", the query will be read from stdin;
// and if the first character of q is "@",
// the @ prefix is removed and the contents of the file specified by the rest of q are returned.
func LoadQuery(q string) (string, error) {
	if q == "-" {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	if len(q) > 0 && q[0] == '@' {
		data, err := ioutil.ReadFile(q[1:])
		if err != nil {
			return "", err
		}

		return string(data), nil
	}

	return q, nil
}
