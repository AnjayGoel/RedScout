package models_test

import (
	"reflect"
	"testing"

	"redscout/models"
)

func TestParseMonitorLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want models.MonitorEntry
		ok   bool
	}{
		{
			name: "single key command",
			line: `1700000000.000001 [0 127.0.0.1:6379] "GET" "user:1"`,
			want: models.MonitorEntry{DB: 0, Command: "get", Keys: []string{"user:1"}},
			ok:   true,
		},
		{
			name: "escaped quotes and spaces in key",
			line: `1700000000.000001 [3 127.0.0.1:6379] "set" "say \"hi\" there" "v"`,
			want: models.MonitorEntry{DB: 3, Command: "set", Keys: []string{`say "hi" there`}},
			ok:   true,
		},
		{
			name: "mset takes every other arg",
			line: `1700000000.000001 [0 127.0.0.1:6379] "mset" "a" "1" "b" "2"`,
			want: models.MonitorEntry{DB: 0, Command: "mset", Keys: []string{"a", "b"}},
			ok:   true,
		},
		{
			name: "eval takes numkeys keys",
			line: `1700000000.000001 [0 lua] "evalsha" "abc" "2" "k1" "k2" "arg"`,
			want: models.MonitorEntry{DB: 0, Command: "evalsha", Keys: []string{"k1", "k2"}},
			ok:   true,
		},
		{
			name: "keyless command",
			line: `1700000000.000001 [0 127.0.0.1:6379] "select" "3"`,
			want: models.MonitorEntry{DB: 0, Command: "select"},
			ok:   true,
		},
		{
			name: "monitor ack",
			line: "OK",
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := models.ParseMonitorLine(tt.line)
			if ok != tt.ok || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMonitorLine() = %+v, %v; want %+v, %v", got, ok, tt.want, tt.ok)
			}
		})
	}
}
