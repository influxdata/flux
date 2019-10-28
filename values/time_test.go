package values

import (
	"testing"
	"time"
)

func TestTime_Round(t *testing.T) {
	for _, tt := range []struct {
		ts   Time
		d    Duration
		want Time
	}{
		{
			ts:   Time(time.Second + 500*time.Millisecond),
			d:    ConvertDuration(time.Second),
			want: Time(2 * time.Second),
		},
		{
			ts:   Time(time.Second + 501*time.Millisecond),
			d:    ConvertDuration(time.Second),
			want: Time(2 * time.Second),
		},
		{
			ts:   Time(time.Second + 499*time.Millisecond),
			d:    ConvertDuration(time.Second),
			want: Time(time.Second),
		},
		{
			ts:   Time(time.Second + 0*time.Millisecond),
			d:    ConvertDuration(time.Second),
			want: Time(time.Second),
		},
	} {
		t.Run(tt.ts.String(), func(t *testing.T) {
			if want, got := tt.want, tt.ts.Round(tt.d); want != got {
				t.Fatalf("unexpected time -want/+got\n\t- %s\n\t%s", want, got)
			}
		})
	}
}

func TestTime_Truncate(t *testing.T) {
	for _, tt := range []struct {
		ts   Time
		d    Duration
		want Time
	}{
		{
			ts:   Time(time.Second + 500*time.Millisecond),
			d:    ConvertDuration(time.Second),
			want: Time(time.Second),
		},
		{
			ts:   Time(time.Second + 501*time.Millisecond),
			d:    ConvertDuration(time.Second),
			want: Time(time.Second),
		},
		{
			ts:   Time(time.Second + 499*time.Millisecond),
			d:    ConvertDuration(time.Second),
			want: Time(time.Second),
		},
		{
			ts:   Time(time.Second + 0*time.Millisecond),
			d:    ConvertDuration(time.Second),
			want: Time(time.Second),
		},
		{
			ts:   Time(time.Second + 999*time.Millisecond),
			d:    ConvertDuration(time.Second),
			want: Time(time.Second),
		},
	} {
		t.Run(tt.ts.String(), func(t *testing.T) {
			if want, got := tt.want, tt.ts.Truncate(tt.d); want != got {
				t.Fatalf("unexpected time -want/+got\n\t- %s\n\t%s", want, got)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	for _, tt := range []struct {
		s    string
		want Duration
	}{
		{
			s: `1mo`,
			want: Duration{
				months: 1,
			},
		},
		{
			s: `1m`,
			want: Duration{
				nsecs: int64(time.Minute),
			},
		},
		{
			s: `1m30s`,
			want: Duration{
				nsecs: int64(time.Minute + 30*time.Second),
			},
		},
		{
			s: `1y`,
			want: Duration{
				months: 12,
			},
		},
		{
			s: `6mo`,
			want: Duration{
				months: 6,
			},
		},
		{
			s: `1y6mo`,
			want: Duration{
				months: 18,
			},
		},
		{
			s: `52w`,
			want: Duration{
				nsecs: int64(52 * 7 * 24 * time.Hour),
			},
		},
		{
			s: `-5m`,
			want: Duration{
				negative: true,
				nsecs:    int64(5 * time.Minute),
			},
		},
		{
			s: `-1y`,
			want: Duration{
				negative: true,
				months:   12,
			},
		},
		{
			s: `1d`,
			want: Duration{
				nsecs: int64(24 * time.Hour),
			},
		},
		{
			s: `1mo3d`,
			want: Duration{
				months: 1,
				nsecs:  int64(3 * 24 * time.Hour),
			},
		},
		{
			s: `1d12h`,
			want: Duration{
				nsecs: int64(36 * time.Hour),
			},
		},
		{
			s:    `0ns`,
			want: Duration{},
		},
		{
			s: `500ms`,
			want: Duration{
				nsecs: int64(500 * time.Millisecond),
			},
		},
		{
			s: `300us`,
			want: Duration{
				nsecs: int64(300 * time.Microsecond),
			},
		},
		{
			s: `300µs`,
			want: Duration{
				nsecs: int64(300 * time.Microsecond),
			},
		},
	} {
		t.Run(tt.s, func(t *testing.T) {
			got, err := ParseDuration(tt.s)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.Equal(tt.want) {
				t.Fatalf("unexpected duration value -want/+got:\n\t- %s\n\t+ %s", tt.want, got)
			}
		})
	}
}

func TestDuration_String(t *testing.T) {
	for _, tt := range []struct {
		d    Duration
		want string
	}{
		{
			d: Duration{
				months: 1,
			},
			want: `1mo`,
		},
		{
			d: Duration{
				nsecs: int64(time.Minute),
			},
			want: `1m`,
		},
		{
			d: Duration{
				nsecs: int64(time.Minute + 30*time.Second),
			},
			want: `1m30s`,
		},
		{
			d: Duration{
				months: 12,
			},
			want: `1y`,
		},
		{
			d: Duration{
				months: 6,
			},
			want: `6mo`,
		},
		{
			d: Duration{
				months: 18,
			},
			want: `1y6mo`,
		},
		{
			d: Duration{
				nsecs: int64(52 * 7 * 24 * time.Hour),
			},
			want: `52w`,
		},
		{
			d: Duration{
				negative: true,
				nsecs:    int64(5 * time.Minute),
			},
			want: `-5m`,
		},
		{
			d: Duration{
				negative: true,
				months:   12,
			},
			want: `-1y`,
		},
		{
			d: Duration{
				nsecs: int64(24 * time.Hour),
			},
			want: `1d`,
		},
		{
			d: Duration{
				months: 1,
				nsecs:  int64(3 * 24 * time.Hour),
			},
			want: `1mo3d`,
		},
		{
			d: Duration{
				nsecs: int64(36 * time.Hour),
			},
			want: `1d12h`,
		},
		{
			d:    Duration{},
			want: `0ns`,
		},
		{
			d: Duration{
				nsecs: int64(500 * time.Millisecond),
			},
			want: `500ms`,
		},
		{
			d: Duration{
				nsecs: int64(300 * time.Microsecond),
			},
			want: `300us`,
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			if got, want := tt.d.String(), tt.want; got != want {
				t.Fatalf("unexpected duration string -want/+got:\n\t- %q\n\t+ %q", want, got)
			}
		})
	}
}
