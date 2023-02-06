package session

import (
	"testing"

	"github.com/matryer/is"
)

func Test_escape(t *testing.T) {
	tests := []struct {
		name     string
		buf      string
		want     string
		sameSize bool
	}{
		{
			name: "should escape \\/ correctly",
			buf:  "hello/world",
			want: "hello\\/world",
		},
		{
			name: "should escape sole \\/ correctly",
			buf:  "/",
			want: "\\/",
		},
		{
			name: "should escape ending \\/ correctly",
			buf:  "hello/",
			want: "hello\\/",
		},
		{
			name: "should escape starting \\/ correctly",
			buf:  "/hello",
			want: "\\/hello",
		},
		{
			name: "should escape \\\\ correctly",
			buf:  "hello\\world",
			want: "hello\\\\world",
		},
		{
			name: "should escape sole \\\\ correctly",
			buf:  "\\",
			want: "\\\\",
		},
		{
			name: "should escape ending \\\\ correctly",
			buf:  "hello\\",
			want: "hello\\\\",
		},
		{
			name: "should escape starting \\\\ correctly",
			buf:  "\\hello",
			want: "\\\\hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			got := Escape([]byte(tt.buf))
			is.Equal(string(got), tt.want)
			is.Equal((len(got) == len([]byte(tt.buf))), tt.sameSize)
		})
	}
}

func Test_unescape(t *testing.T) {
	tests := []struct {
		name     string
		buf      string
		want     string
		sameSize bool
	}{
		{
			name: "should unescape \\/ correctly",
			buf:  "hello\\/world",
			want: "hello/world",
		},
		{
			name: "should unescape sole \\/ correctly",
			buf:  "\\/",
			want: "/",
		},
		{
			name: "should unescape ending \\/ correctly",
			buf:  "hello\\/",
			want: "hello/",
		},
		{
			name: "should unescape starting \\/ correctly",
			buf:  "\\/hello",
			want: "/hello",
		},
		{
			name:     "should NOT unescape single backslash",
			buf:      "\\hello",
			want:     "\\hello",
			sameSize: true,
		},
		{
			name: "should unescape \\\\ correctly",
			buf:  "hello\\\\world",
			want: "hello\\world",
		},
		{
			name: "should unescape sole \\\\ correctly",
			buf:  "\\\\",
			want: "\\",
		},
		{
			name: "should unescape ending \\\\ correctly",
			buf:  "hello\\\\",
			want: "hello\\",
		},
		{
			name: "should unescape starting \\\\ correctly",
			buf:  "\\\\hello",
			want: "\\hello",
		},
		{
			name:     "should NOT unescape single backslash",
			buf:      "\\hello",
			want:     "\\hello",
			sameSize: true,
		},
		{
			name: "should convert the line with multiple escapes",
			buf:  `foo\/bar\/baz\nfoo\\bar\\baz`,
			want: `foo/bar/baz\nfoo\bar\baz`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			got := Unescape([]byte(tt.buf))
			is.Equal(string(got), tt.want)
			is.Equal((len(got) == len([]byte(tt.buf))), tt.sameSize)
		})
	}
}
