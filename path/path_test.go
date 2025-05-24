package path

import (
	"fmt"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	type testCase struct {
		source string
		want   *Path
		err    string
	}

	errorCase := func(source string, lines ...string) testCase {
		return testCase{
			source: source,
			want:   nil,
			err:    strings.Join(lines, "\n"),
		}
	}

	successCase := func(source string, segments ...Segment) testCase {
		return testCase{
			source: source,
			want:   &Path{segments: segments},
			err:    "",
		}
	}

	tests := []testCase{
		successCase(""),
		errorCase(
			".",
			"invalid path: .",
			"              ^ expected key or index, got .",
		),
		errorCase(
			".foo",
			"invalid path: .foo",
			"              ^ expected key or index, got .",
		),
		successCase("$"),
		successCase(
			"foo",
			Segment{key: "foo", index: -1},
		),
		successCase(
			"foo.bar",
			Segment{key: "foo", index: -1},
			Segment{key: "bar", index: -1},
		),
		successCase(
			"foo.bar[0]",
			Segment{key: "foo", index: -1},
			Segment{key: "bar", index: -1},
			Segment{index: 0},
		),
		successCase(
			"foo.bar[0].baz",
			Segment{key: "foo", index: -1},
			Segment{key: "bar", index: -1},
			Segment{index: 0},
			Segment{key: "baz", index: -1},
		),
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.source), func(t *testing.T) {
			got, err := Parse(tt.source)

			if tt.err == "" {
				if err != nil {
					t.Fatalf("returned unexpected error %v", err)
				}

				if tt.want != nil && got == nil || tt.want == nil && got != nil {
					t.Fatalf("expected %q, got %q", tt.want.String(), got.String())
				}

				if !got.Equals(tt.want) {
					t.Fatalf("expected %q, got %q", tt.want.String(), got.String())
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				if err.Error() != tt.err {
					t.Fatalf("expected error %q, got %q", tt.err, err.Error())
				}
			}
		})
	}
}
