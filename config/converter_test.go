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
			got, got1 := fEnum(tt.args.s, tt.args.set)
			if got != tt.want {
				t.Errorf("enum() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("enum() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_fNetwork(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{""}, false},
		{"main", args{"MAIN"}, true},
		{"local", args{"LOCAL"}, true},
		{"test", args{"TEST"}, true},
		{"testnet", args{"fct_community_test"}, true},
		{"alphanumeral", args{"f00b4r"}, true},
		{"invalid", args{"abc%def"}, false},
		{"invalid", args{""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fNetwork(tt.args.s); got != tt.want {
				t.Errorf("fNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stringFTag(t *testing.T) {
	type args struct {
		f   string
		val string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"f empty", args{"", ""}, true},
		{"url empty", args{"url", ""}, false},
		{"url basic", args{"url", "https://www.factomprotocol.org/"}, false},
		{"url localhost", args{"url", "http://localhost/seed"}, false},
		{"url main seed", args{"url", "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/mainseed.txt"}, false},
		{"url testnet seed", args{"url", "https://raw.githubusercontent.com/FactomProject/communitytestnet/master/seeds/testnetseeds.txt"}, false},
		{"hex64 empty", args{"hex64", ""}, true},
		{"hex64 nonhex", args{"hex64", "z"}, true},
		{"hex64 hex but short", args{"hex64", "c0ffee"}, true},
		{"hex64 zero but short", args{"hex64", "000000000000000000000000000000000000000000000000000000000000000"}, true},
		{"hex64 zero", args{"hex64", "0000000000000000000000000000000000000000000000000000000000000000"}, false},
		{"hex64 all", args{"hex64", "0123456789ABCDEF0123456789abcdef0123456789ABCDEF0123456789abcdef"}, false},
		{"hex64 f", args{"hex64", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}, false},
		{"hex64 F", args{"hex64", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"}, false},
		{"hex64 f-oneoff 1", args{"hex64", "fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffg"}, true},
		{"hex64 f-oneoff 2", args{"hex64", "ffffffffffffffgfffffffffffffffffffffffffffffffffffffffffffffffff"}, true},
		{"hex64 f-oneoff 3", args{"hex64", "gfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}, true},
		{"hex64 f-oneoff 4", args{"hex64", "ffffffffffzfffffffffffffffffffffffffffffffffffffffffffffffffffff"}, true},
		{"sha256 of test", args{"hex64", "9F86D081884C7D659A2FEAA0C55AD015A3BF4F1B2B0B822CD15D6C15B0F00A08"}, false},
		{"network", args{"network", "fnetwork_are_tested_separately_in_this_file"}, false},
		{"network empty", args{"network", ""}, true},
		{"alpha empty", args{"alpha", ""}, false},
		{"alpha all", args{"alpha", "abcdefghijklmonpqrstuvwxyABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"}, false},
		{"alpha nonalpha", args{"alpha", "$"}, true},
		{"alpha nonalpha 2", args{"alpha", "abc$"}, true},
		{"alpha nonalpha 3", args{"alpha", "99.99"}, true},
		{"alpha nonalpha 4", args{"alpha", "normal text"}, true},
		{"ipport empty", args{"ipport", ""}, true},
		{"ipport localhost", args{"ipport", "localhost"}, true},
		{"ipport localhost w port", args{"ipport", "localhost:80"}, false},
		{"ipport loopback", args{"ipport", "127.0.0.1"}, true},
		{"ipport loopback w :", args{"ipport", "127.0.0.1:"}, true},
		{"ipport loopback w zero port", args{"ipport", "127.0.0.1:0"}, false},
		{"ipport loopback w port", args{"ipport", "127.0.0.1:80"}, false},
		{"ipport mainseed node", args{"ipport", "52.17.183.121:8108"}, false},
		{"ipport mainseed node 2", args{"ipport", "34.248.202.6:8108"}, false},
		{"ipport mainseed node 3", args{"ipport", "52.17.183.121:8108 "}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := stringFTag(tt.args.f, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("stringFTag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_intFTag(t *testing.T) {
	type args struct {
		f   string
		val string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"empty f", args{"", ""}, 0, true},
		{"time empty", args{"time", ""}, 0, true}, // ftime is already tested separately earlier
		{"time 0", args{"time", "0"}, 0, false},
		{"time text 1", args{"time", "one"}, 0, true},
		{"time 0s", args{"time", "0s"}, 0, false},
		{"time 0m", args{"time", "0m"}, 0, false},
		{"time 0h", args{"time", "0h"}, 0, false},
		{"time 0d", args{"time", "0d"}, 0, false},
		{"time 1s", args{"time", "1s"}, 1, false},
		{"time 1m", args{"time", "1m"}, 60, false},
		{"time 1h", args{"time", "1h"}, 3600, false},
		{"time 1d", args{"time", "1d"}, 86400, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := intFTag(tt.args.f, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("intFTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("intFTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
