package config

import (
	"testing"

	"github.com/go-ini/ini"
)

func Test_determineNetwork(t *testing.T) {
	i1 := ini.Empty()
	i2, _ := ini.InsensitiveLoad([]byte("[factomd]\nnetwork=foo"))
	f1, _ := ParseFlags([]string{})
	f2, _ := ParseFlags([]string{"-network", "bar"})
	f3, _ := ParseFlags([]string{"-n", "b"})
	type args struct {
		file  *ini.File
		flags *Flags
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"basic empty empty", args{i1, f1}, ""},
		{"basic empty long", args{i1, f2}, "bar"},
		{"basic empty short", args{i1, f3}, "b"},
		{"basic set empty", args{i2, f1}, "foo"},
		{"basic set long", args{i2, f2}, "bar"},
		{"basic set short", args{i2, f3}, "b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineNetwork(tt.args.file, tt.args.flags); got != tt.want {
				t.Errorf("determineNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}
