package config

import (
	"testing"
)

func Test_fTime(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"empty", args{""}, 0, true},
		{"empty w mod s", args{"s"}, 0, true},
		{"empty w mod h", args{"h"}, 0, true},
		{"empty w mod m", args{"m"}, 0, true},
		{"empty w mod d", args{"d"}, 0, true},
		{"empty w invalid mod", args{"f"}, 0, true},
		{"invalid mod", args{"1f"}, 0, true},
		{"zero seconds", args{"0"}, 0, false},
		{"one second", args{"1"}, 1, false},
		{"negative seconds", args{"-1"}, 0, true},
		{"fifty seconds", args{"50"}, 50, false},
		{"zero seconds w mod", args{"0s"}, 0, false},
		{"fifty seconds w mod", args{"50s"}, 50, false},
		{"zero minutes", args{"0m"}, 0, false},
		{"one minute", args{"1m"}, 60, false},
		{"hundred minutes", args{"100m"}, 6000, false},
		{"zero hours", args{"0h"}, 0, false},
		{"one hours", args{"1h"}, 3600, false},
		{"two hours", args{"2h"}, 7200, false},
		{"zero days", args{"0d"}, 0, false},
		{"one days", args{"1d"}, 86400, false},
		{"two days", args{"2d"}, 86400 * 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fTime(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("fTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_enum(t *testing.T) {
	set := "A,B,AB,ABA,ABAB,FOO,BAR"
	set2 := "      A,B,AB,aba,ABAB,  FOO  ,B A R"
	type args struct {
		s   string
		set string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{"empty choice", args{"", set}, "", false},
		{"empty set", args{"foo", ""}, "", false},
		{"trim & case", args{" a   \n", set}, "A", true},
		{"first", args{"A", set}, "A", true},
		{"second", args{"B", set}, "B", true},
		{"third", args{"B", set}, "B", true},
		{"last", args{"BAR", set}, "BAR", true},
		{"trim list 1", args{"A", set2}, "A", true},
		{"trim list 2", args{"FOO", set2}, "FOO", true},
		{"nonexistent", args{"BAR", set2}, "", false},
		{"case", args{"b a r", set2}, "B A R", true},
		{"partial", args{"BAB", set2}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := enum(tt.args.s, tt.args.set)
			if got != tt.want {
				t.Errorf("enum() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("enum() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
