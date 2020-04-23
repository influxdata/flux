package system

import (
    "context"
    "testing"
    "time"

    "github.com/influxdata/flux/lang/execdeps"
)

func TestOverrideWithContext(t *testing.T) {
    dep := execdeps.DefaultExecutionDependencies()
    now := time.Unix(1, 0)

    dep.Now = &now

    v, err := systemTimeFunc.Function().Call(dep.Inject(context.Background()), nil)
    if err != nil {
        t.Fatal(err)
    }
    if !v.Time().Time().Equal(now) {
        t.Errorf("expected %v, but got %v", now.UTC(), v.Time().Time().UTC())
    }
}
