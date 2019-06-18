package config

import (
	"strings"
	"testing"
)

func TestWordWrap(t *testing.T) {
	type args struct {
		s      string
		limit  int
		prefix string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"blank", args{"", 0, ""}, ""},
		{"one char zero limit", args{"a", 0, ""}, "a"},
		{"one word zero limit", args{"abc", 0, ""}, "abc"},
		{"two words zero limit", args{"foo bar", 0, ""}, "foo\nbar"},
		{"two words smaller limit", args{"foo bar", 6, ""}, "foo\nbar"},
		{"two words equal limit", args{"foo bar", 7, ""}, "foo bar"},
		{"two words bigger limit", args{"foo bar", 8, ""}, "foo bar"},
		{"one char prefix", args{"a", 5, "b"}, "ba"},
		{"one char tab prefix", args{"a", 5, "  "}, "  a"},
		{"two words tab prefix", args{"foo bar", 5, "  "}, "  foo\n  bar"},
		{"longer sentence zero limit", args{"The quick brown fox jumps over the lazy dog", 0, ""}, "The\nquick\nbrown\nfox\njumps\nover\nthe\nlazy\ndog"},
		{"longer sentence 1 limit", args{"The quick brown fox jumps over the lazy dog", 1, ""}, "The\nquick\nbrown\nfox\njumps\nover\nthe\nlazy\ndog"},
		{"longer sentence 8 limit", args{"The quick brown fox jumps over the lazy dog", 8, ""}, "The\nquick\nbrown\nfox\njumps\nover the\nlazy dog"},
		{"longer sentence 9 limit", args{"The quick brown fox jumps over the lazy dog", 9, ""}, "The quick\nbrown fox\njumps\nover the\nlazy dog"},
		{"longer sentence zero limit prefix 2", args{"The quick brown fox jumps over the lazy dog", 0, "  "}, "  The\n  quick\n  brown\n  fox\n  jumps\n  over\n  the\n  lazy\n  dog"},
		{"longer sentence zero limit prefix ab", args{"The quick brown fox jumps over the lazy dog", 0, "ab"}, "abThe\nabquick\nabbrown\nabfox\nabjumps\nabover\nabthe\nablazy\nabdog"},
		{"longer sentence 15 limit prefix ab", args{"The quick brown fox jumps over the lazy dog", 15, "ab"}, "abThe quick\nabbrown fox\nabjumps over\nabthe lazy dog"},
		{"multiline zero limit no prefix", args{"foo bar\nboo car\nzoo far", 0, ""}, "foo\nbar\nboo\ncar\nzoo\nfar"},
		{"multiline equal limit no prefix", args{"foo bar\nboo car\nzoo far", 7, ""}, "foo bar\nboo car\nzoo far"},
		{"multiline zero limit 1 prefix", args{"foo bar\nboo car\nzoo far", 0, " "}, " foo\n bar\n boo\n car\n zoo\n far"},
		{"multiline equal limit 1 prefix", args{"foo bar\nboo car\nzoo far", 7, " "}, " foo\n bar\n boo\n car\n zoo\n far"},
		{"multiline equal+1 limit 1 prefix", args{"foo bar\nboo car\nzoo far", 8, " "}, " foo bar\n boo car\n zoo far"},
		{"multiline preserve indent", args{"The quick brown fox jumps over the lazy dog\n  The quick brown fox jumps over the lazy dog\n   The quick brown fox jumps over the lazy dog", 15, ""}, "The quick brown\nfox jumps over\nthe lazy dog\n  The quick\n  brown fox\n  jumps over\n  the lazy dog\n   The quick\n   brown fox\n   jumps over\n   the lazy dog"},
		{"multiline preserve indent four prefix", args{"The quick brown fox jumps over the lazy dog\n  The quick brown fox jumps over the lazy dog\n   The quick brown fox jumps over the lazy dog", 15, "    "}, "    The quick\n    brown fox\n    jumps over\n    the lazy\n    dog\n      The quick\n      brown fox\n      jumps\n      over the\n      lazy dog\n       The\n       quick\n       brown\n       fox\n       jumps\n       over the\n       lazy dog"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WordWrap(tt.args.s, tt.args.limit, tt.args.prefix); got != tt.want {
				t.Errorf("WordWrap() = %v, want %v", strings.Replace(got, "\n", "\\n", -1), strings.Replace(tt.want, "\n", "\\n", -1))
			}
		})
	}
}

func Test_lcFirst(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{""}, ""},
		{"lowercase", args{"foo"}, "foo"},
		{"uppercase", args{"FOO"}, "fOO"},
		{"regular case", args{"Foo"}, "foo"},
		{"camel case", args{"fooBar"}, "fooBar"},
		{"pascal case", args{"FooBar"}, "fooBar"},
		{"special p2p ", args{"P2PFooBar"}, "p2pFooBar"},
		{"two fords", args{"FOO BAR"}, "fOO BAR"},
		{"special api", args{"APIFooBar"}, "apiFooBar"},
		{"special api bad", args{"APiFooBar"}, "aPiFooBar"},
		{"special Db", args{"DBFooBar"}, "dbFooBar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lcFirst(tt.args.s); got != tt.want {
				t.Errorf("lcFirst() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prettyEnum(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"blank", args{""}, ""},
		{"one", args{"a"}, "a"},
		{"two words", args{"foo bar"}, "foo bar"},
		{"two enums", args{"foo,bar"}, "foo, bar"},
		{"three enums", args{"foo,bar,boo"}, "foo, bar, boo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prettyEnum(tt.args.s); got != tt.want {
				t.Errorf("prettyEnum() = %v, want %v", got, tt.want)
			}
		})
	}
}
