package config

import (
	"testing"
)

func TestFlags_Unused(t *testing.T) {
	f := new(Flags)
	f.flags = make(map[string]string)
	f.checked = make(map[string]bool)
	f.add("foo", "bar")
	f.add("boo", "car")

	unused := f.Unused()
	if len(unused) != 2 {
		t.Errorf("Expected 2 unused, foo and boo. Got: %v", unused)
	}

	f.Get("foo")

	unused = f.Unused()
	if len(unused) != 1 {
		t.Errorf("Expected 1 unused, boo. Got: %v", unused)
	}

	f.Get("boo")

	unused = f.Unused()
	if len(unused) != 0 {
		t.Errorf("Expected 0 unused. Got: %v", unused)
	}
}

type s map[string]string

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    s
		wantErr bool
	}{
		{"no flags", nil, nil, false},
		{"empty string", []string{}, nil, false},
		{"invalid flag", []string{"a"}, nil, true},
		{"invalid flag whitespace", []string{" a"}, nil, true},
		{"invalid flags", []string{"a", "b"}, nil, true},
		{"invalid blank", []string{""}, nil, true},
		{"invalid blank one dash", []string{"-"}, nil, true},
		{"invalid blank two dash", []string{"--"}, nil, true},
		{"mixed invalid flags", []string{"a", "-b"}, nil, true},
		{"valid one dash", []string{"-foo", "bar"}, s{"foo": "bar"}, false},
		{"valid one dash multiword", []string{"-foo", "bar bar bar"}, s{"foo": "bar bar bar"}, false},
		{"valid case", []string{"-FOO", "bar"}, s{"foo": "bar"}, false},
		{"valid one dash, equals", []string{"-foo=bar"}, s{"foo": "bar"}, false},
		{"valid one dash, equals multiword", []string{"-foo=bar bar bar"}, s{"foo": "bar bar bar"}, false},
		{"valid two dash", []string{"--foo", "bar"}, s{"foo": "bar"}, false},
		{"valid two dash, equals", []string{"--foo=bar"}, s{"foo": "bar"}, false},
		{"valid one dash bool", []string{"-foo"}, s{"foo": "true"}, false},
		{"valid two dash bool", []string{"--foo"}, s{"foo": "true"}, false},
		{"invalid double single", []string{"-foo", "-foo"}, nil, true},
		{"invalid double double", []string{"--foo", "--foo"}, nil, true},
		{"invalid double mixed", []string{"-foo", "--foo"}, nil, true},
		{"invalid double case", []string{"-foo", "-FOO"}, nil, true},
		{"valid blank", []string{"-foo="}, s{"foo": ""}, false},
		{"valid mixed 1", []string{"--foo", "-a=b"}, s{"foo": "true", "a": "b"}, false},
		{"valid mixed 2", []string{"--foo=bar", "-a==b"}, s{"foo": "bar", "a": "=b"}, false},
		{"valid mixed 3", []string{"--foo", "-a=b", "-boo", "car"}, s{"foo": "true", "a": "b", "boo": "car"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(tt.want) > 0 {
				for k, v := range tt.want {
					got, ok := got.Get(k)
					if !ok {
						t.Errorf("Flags.parse() did not save flag %v", k)
					} else if got != v {
						t.Errorf("Flags.parse() flag %v: wanted %v, got %v", k, v, got)
					}
				}

				for _, f := range got.Unused() {
					t.Errorf("Error in test, unused flag: %s", f)
				}
			}
		})
	}
}
