package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/dependencies/feature"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/fluxinit"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/libflux/go/libflux"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/repl"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type ScopeHolder struct {
	ctx context.Context

	scope    values.Scope
	itrp     *interpreter.Interpreter
	analyzer *libflux.Analyzer
	importer interpreter.Importer

	cancelMu   sync.Mutex
	cancelFunc context.CancelFunc

	resChan chan string
}

func New(ctx context.Context) *ScopeHolder {
	scope := values.NewScope()
	importer := runtime.StdLib()
	for _, p := range runtime.PreludeList {
		pkg, err := importer.ImportPackageObject(p)
		if err != nil {
			panic(err)
		}
		pkg.Range(scope.Set)
	}

	analyzer, err := libflux.NewAnalyzerWithOptions(libflux.NewOptions(ctx))
	if err != nil {
		panic(err)
	}

	repl := &ScopeHolder{
		ctx:      ctx,
		scope:    scope,
		itrp:     interpreter.NewInterpreter(nil, &lang.ExecOptsConfig{}),
		analyzer: analyzer,
		importer: importer,
	}

	return repl
}

type rwCloser struct {
	io.ReadCloser
	io.WriteCloser
}

func (rw rwCloser) Close() error {
	err := rw.ReadCloser.Close()
	if err := rw.WriteCloser.Close(); err != nil {
		return err
	}
	return err
}

type Response struct {
	Result string
}

type Request struct {
	A string `json:"input"`
}

type Service struct {
	c   chan string
	res chan string
}

func (s *Service) DidOutput(req Request, resp *Response) error {
	s.c <- req.A
	result := <-s.res
	*resp = Response{result}
	return nil
}

type API int

func (r *ScopeHolder) Run() {
	s := rpc.NewServer()
	c := make(chan string)
	//for the input result
	calc_chan := make(chan string)
	r.resChan = calc_chan

	serv := Service{c, calc_chan}
	s.Register(&serv)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	go func() {
		for range sigs {
			r.cancel()
		}
	}()

	go s.ServeCodec(jsonrpc.NewServerCodec(rwCloser{os.Stdin, os.Stdout})) //somehow need to get the input that is being
	for {
		res := <-c
		r.InputAndPrintError(res)
	}

}

func (r *ScopeHolder) cancel() {
	r.cancelMu.Lock()
	defer r.cancelMu.Unlock()
	if r.cancelFunc != nil {
		r.cancelFunc()
		r.cancelFunc = nil
	}
}

func (r *ScopeHolder) setCancel(cf context.CancelFunc) {
	r.cancelMu.Lock()
	defer r.cancelMu.Unlock()
	r.cancelFunc = cf
}
func (r *ScopeHolder) clearCancel() {
	r.setCancel(nil)
}

func (r *ScopeHolder) Input(t string) (*libflux.FluxError, error) {
	a, err := r.executeLine(t)
	return a, err
}

// input processes a line of input and prints the result.
func (r *ScopeHolder) InputAndPrintError(t string) {
	if fluxError, err := r.executeLine(t); err != nil {
		if fluxError != nil {

			fluxError.Print()
		} else {
			fmt.Println("Error:", err)
		}
	}
}

func (r *ScopeHolder) Eval(t string) ([]interpreter.SideEffect, error) {
	s, _, err := r.evalWithFluxError(t)
	return s, err
}

func (r *ScopeHolder) evalWithFluxError(t string) ([]interpreter.SideEffect, *libflux.FluxError, error) {
	if t == "" {
		return nil, nil, nil
	}

	if t[0] == '@' {
		q, err := LoadQuery(t)
		if err != nil {
			return nil, nil, err
		}
		t = q
	}

	pkg, fluxError, err := r.analyzeLine(t)
	if err != nil {
		return nil, fluxError, err
	}

	ctx, span := dependency.Inject(r.ctx, execute.DefaultExecutionDependencies())
	defer span.Finish()

	x, err := r.itrp.Eval(ctx, pkg, r.scope, r.importer)
	return x, nil, err
}

// executeLine processes a line of input.
// If the input evaluates to a valid value, that value is returned.
func (r *ScopeHolder) executeLine(t string) (*libflux.FluxError, error) {
	ses, fluxError, err := r.evalWithFluxError(t)
	if err != nil {
		return fluxError, err
	}

	for _, se := range ses {
		if _, ok := se.Node.(*semantic.ExpressionStatement); ok {
			if t, ok := se.Value.(*flux.TableObject); ok {
				now, ok := r.scope.Lookup("now")
				if !ok {
					return nil, fmt.Errorf("now option not set")
				}
				nowTime, err := now.Function().Call(r.ctx, nil)
				if err != nil {
					return nil, err
				}
				s, err := spec.FromTableObject(r.ctx, t, nowTime.Time().Time())
				if err != nil {
					return nil, err
				}
				if err := r.doQuery(r.ctx, s); err != nil {
					return nil, err
				}
			} else {

				var a []byte
				buf := bytes.NewBuffer(a)
				values.Display(buf, se.Value)
				r.resChan <- buf.String()
			}
		}
	}
	return nil, nil
}

func (r *ScopeHolder) analyzeLine(t string) (*semantic.Package, *libflux.FluxError, error) {
	pkg, fluxError := r.analyzer.AnalyzeString(t)
	if fluxError != nil {
		return nil, fluxError, fluxError.GoError()
	}

	bs, err := pkg.MarshalFB()
	if err != nil {
		return nil, nil, err
	}
	x, err := semantic.DeserializeFromFlatBuffer(bs)
	return x, nil, err
}

func (r *ScopeHolder) doQuery(ctx context.Context, spec *flux.Spec) error {
	// Setup cancel context
	ctx, cancelFunc := context.WithCancel(ctx)
	r.setCancel(cancelFunc)
	defer cancelFunc()
	defer r.clearCancel()

	c := repl.Compiler{
		Spec: spec,
	}

	program, err := c.Compile(ctx, runtime.Default)
	if err != nil {
		return err
	}
	alloc := &memory.ResourceAllocator{}

	qry, err := program.Start(ctx, alloc)
	if err != nil {
		return err
	}
	defer qry.Done()

	for result := range qry.Results() {
		tables := result.Tables()
		fmt.Println("Result:", result.Name())
		if err := tables.Do(func(tbl flux.Table) error {
			_, err := execute.NewFormatter(tbl, nil).WriteTo(os.Stdout)
			return err
		}); err != nil {
			return err
		}
	}
	qry.Done()
	return qry.Err()
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

const DefaultInfluxDBHost = "http://localhost:9999"

func injectDependencies(ctx context.Context) (context.Context, *dependency.Span) {
	deps := dependencies.NewDefaultDependencies(DefaultInfluxDBHost)
	return dependency.Inject(ctx, deps)
}

func runE(args []string) error {

	ctx, close := context.Background(), func() {}

	defer close()

	// Defer initialization until other common errors
	// have already passed to avoid a long load time
	// for a simple unrelated error.
	fluxinit.FluxInit()
	ctx, span := injectDependencies(ctx)
	defer span.Finish()

	flagger := executetest.TestFlagger{}

	ctx = feature.Dependency{Flagger: flagger}.Inject(ctx)

	var opts []repl.Option
	opts = append(opts, repl.EnableSuggestions())
	return replE(ctx, opts...)

}

func replE(ctx context.Context, opts ...repl.Option) error {
	r := New(ctx)
	r.Run()
	return nil
}

func main() {
	runE(nil)
}
