package sql

import (
	"context"
	"crypto/x509"
	"database/sql"
	"fmt"
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/execute"
	"github.com/InfluxCommunity/flux/values"
	influxdbiox "github.com/metrico/influxdb-iox-client-go/v2"
	"github.com/metrico/influxdb-iox-client-go/v2/ioxsql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	neturl "net/url"
	"strconv"
	"time"
)

type IOXRowReader struct {
	rows           *sql.Rows
	rawColumnNames []string
	rawColumnTypes []*sql.ColumnType
}

func (i *IOXRowReader) Next() bool {
	return i.rows.Next()
}

func nullOrVal[a any](f func(a a) values.Value, v *a) values.Value {
	if v == nil {
		return values.Null
	}
	return f(*v)
}

func doublePtr[T any](a T) **T {
	b := &a
	return &b
}

func (i *IOXRowReader) GetNextRow() ([]values.Value, error) {
	cTypes, err := i.rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	scans := make([]interface{}, len(cTypes))
	for i, ct := range cTypes {
		switch ct.ScanType().String() {
		case "float32":
			scans[i] = doublePtr[float32](0)
		case "float64":
			scans[i] = doublePtr[float64](0)
		case "time.Time":
			scans[i] = doublePtr[time.Time](time.Time{})
		case "uint64":
			scans[i] = doublePtr[uint64](0)
		case "int64":
			scans[i] = doublePtr[int64](0)
		case "string":
			scans[i] = doublePtr[string]("")
		case "[]uint8":
			scans[i] = doublePtr[[]uint8]([]uint8{})
		case "bool":
			scans[i] = doublePtr[bool](false)
		}
	}

	err = i.rows.Scan(scans...)
	if err != nil {
		return nil, err
	}

	res := make([]values.Value, len(cTypes))
	for i, v := range scans {
		switch v := v.(type) {
		case **string:
			if v == nil {
				res[i] = values.NewString("")
			}
			res[i] = nullOrVal(values.NewString, *v)
		case **[]byte:
			res[i] = nullOrVal(values.NewBytes, *v)
		case **int64:
			res[i] = nullOrVal(values.NewInt, *v)
		case **uint64:
			res[i] = nullOrVal(values.NewUInt, *v)
		case **float64:
			res[i] = nullOrVal(values.NewFloat, *v)
		case **bool:
			res[i] = nullOrVal(values.NewBool, *v)
		case **time.Time:
			res[i] = nullOrVal(func(v time.Time) values.Value {
				return values.NewTime(values.ConvertTime(v))
			}, *v)
		case **time.Duration:
			res[i] = nullOrVal(func(v time.Duration) values.Value {
				return values.NewDuration(values.ConvertDurationNsecs(v))
			}, *v)
		default:
			res[i] = values.InvalidValue
		}
	}
	return res, nil
}

func (i *IOXRowReader) ColumnNames() []string {
	return i.rawColumnNames
}

func (i *IOXRowReader) ColumnTypes() []flux.ColType {
	res := make([]flux.ColType, len(i.ColumnNames()))
	for i, tp := range i.rawColumnTypes {
		switch tp.ScanType().String() {
		case "float32", "float64":
			res[i] = flux.TFloat
		case "time.Time":
			res[i] = flux.TTime
		case "uint64":
			res[i] = flux.TUInt
		case "int64":
			res[i] = flux.TInt
		case "string":
			res[i] = flux.TString
		case "[]uint8":
			res[i] = flux.TString
		case "bool":
			res[i] = flux.TBool
		}
	}
	return res
}

func (i *IOXRowReader) SetColumns(j []interface{}) {
	panic("not implemented")
}

func (i *IOXRowReader) Close() error {
	return i.rows.Close()
}

func NewIOXRowReader(r *sql.Rows) (execute.RowReader, error) {
	types, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}
	names, err := r.Columns()
	if err != nil {
		return nil, err
	}
	return &IOXRowReader{r, names, types}, nil
}

type perRpcCredentials struct {
	db    string
	token string
}

func (m *perRpcCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	res := map[string]string{
		"database": m.db,
	}
	if m.token != "" {
		res["authorization"] = "Bearer " + m.token
	}
	return res, nil
}

func (m *perRpcCredentials) RequireTransportSecurity() bool {
	return false
}

func ioxOpenFunction(driverName, datasourceName string) func() (*sql.DB, error) {
	return func() (*sql.DB, error) {
		url, err := neturl.Parse(datasourceName)
		if err != nil {
			return nil, err
		}

		creds := &perRpcCredentials{db: url.Path[1:]}

		cfg := influxdbiox.ClientConfig{
			Address:     fmt.Sprintf("%s:%s", url.Hostname(), url.Port()),
			Namespace:   url.Path[1:],
			DialOptions: []grpc.DialOption{grpc.WithPerRPCCredentials(creds)},
		}
		for k, v := range url.Query() {
			switch k {
			case "secure":
				_v, err := strconv.ParseBool(v[0])
				if err != nil {
					return nil, err
				}
				if !_v {
					continue
				}
				pool, err := x509.SystemCertPool()
				if err != nil {
					return nil, err
				}
				cfg.DialOptions = append(cfg.DialOptions,
					grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")))
			case "token":
				creds.token = v[0]
			}
		}

		return sql.OpenDB(ioxsql.NewConnector(&cfg)), nil
	}
}
