package execute

import (
	"context"

	"github.com/influxdata/flux/metadata"
)

func RecordEvent(ctx context.Context, key string) {
	if HaveExecutionDependencies(ctx) {
		deps := GetExecutionDependencies(ctx)
		deps.Metadata.ReadWriteView(func(meta *metadata.Metadata) {
			if _, ok := meta.Get(key); !ok {
				meta.Add(key, true)
			}
		})
	}
}
