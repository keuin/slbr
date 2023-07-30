package pretty

import (
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	type args struct {
		duration time.Duration
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "zero",
			args: args{0},
			want: "00:00:00",
		},
		{
			name: "1s",
			args: args{time.Second},
			want: "00:00:01",
		},
		{
			name: "2s",
			args: args{time.Second * 2},
			want: "00:00:02",
		},
		{
			name: "59s",
			args: args{time.Second * 59},
			want: "00:00:59",
		},
		{
			name: "1m",
			args: args{time.Second * 60},
			want: "00:01:00",
		},
		{
			name: "1m1s",
			args: args{time.Second * 61},
			want: "00:01:01",
		},
		{
			name: "1h",
			args: args{time.Second * 3600},
			want: "01:00:00",
		},
		{
			name: "54h7m13s",
			args: args{time.Hour*54 + time.Minute*7 + time.Second*13},
			want: "54:07:13",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Duration(tt.args.duration); got != tt.want {
				t.Errorf("Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}
