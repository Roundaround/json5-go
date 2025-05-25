package path

import (
	"errors"
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	type testCase struct {
		source string
		want   *Path
		err    *PathError
	}

	errorCase := func(source string, errmsg string, pos int) testCase {
		return testCase{
			source: source,
			want:   nil,
			err:    &PathError{errors.New(errmsg), pos, source},
		}
	}

	successCase := func(source string, segments ...Segment) testCase {
		return testCase{
			source: source,
			want:   &Path{segments: segments},
			err:    nil,
		}
	}

	tests := []testCase{
		successCase(""),
		errorCase(
			".",
			"expected key or index, got '.'",
			0,
		),
		errorCase(
			".foo",
			"expected key or index, got '.'",
			0,
		),
		successCase("$"), // Root gets trimmed
		successCase(
			"foo",
			Segment{key: "foo"},
		),
		successCase(
			"[0]",
			Segment{index: 0},
		),
		successCase(
			"['0']",
			Segment{key: "0"},
		),
		successCase(
			"[\"0\"]",
			Segment{key: "0"},
		),
		errorCase(
			"'foo'",
			"expected key or index, got \"'\"",
			0,
		),
		errorCase(
			"\"foo\"",
			"expected key or index, got '\"'",
			0,
		),
		errorCase(
			"foo.'bar'",
			"expected key, got \"'\"",
			4,
		),
		successCase(
			"Foo",
			Segment{key: "Foo"},
		),
		successCase(
			"$foo",
			Segment{key: "$foo"},
		),
		successCase(
			"_foo",
			Segment{key: "_foo"},
		),
		successCase(
			"f$o_o$",
			Segment{key: "f$o_o$"},
		),
		successCase(
			"foo9",
			Segment{key: "foo9"},
		),
		errorCase(
			"9foo",
			"keys cannot start with a digit, got '9'",
			0,
		),
		errorCase(
			"98",
			"indexes must be in square brackets, got '9'",
			0,
		),
		errorCase(
			"98foo",
			"indexes must be in square brackets, got '9'",
			0,
		),
		successCase(
			"foo.bar",
			Segment{key: "foo"},
			Segment{key: "bar"},
		),
		successCase(
			"foo.bar[0]",
			Segment{key: "foo"},
			Segment{key: "bar"},
			Segment{index: 0},
		),
		successCase(
			"foo.bar[3]",
			Segment{key: "foo"},
			Segment{key: "bar"},
			Segment{index: 3},
		),
		successCase(
			"foo.bar[3].baz",
			Segment{key: "foo"},
			Segment{key: "bar"},
			Segment{index: 3},
			Segment{key: "baz"},
		),
		successCase(
			"foo.bar[3][9].baz",
			Segment{key: "foo"},
			Segment{key: "bar"},
			Segment{index: 3},
			Segment{index: 9},
			Segment{key: "baz"},
		),
		successCase(
			"[0][0][0][0]",
			Segment{index: 0},
			Segment{index: 0},
			Segment{index: 0},
			Segment{index: 0},
		),
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.source), func(t *testing.T) {
			got, err := Parse(tt.source)

			if tt.err == nil {
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

				var perr *PathError
				if !errors.As(err, &perr) {
					t.Fatalf("expected *PathError, got %T", err)
				}

				if perr.pos != tt.err.pos {
					t.Fatalf("expected error at pos %d, got %d", tt.err.pos, perr.pos)
				}

				if perr.path != tt.err.path {
					t.Fatalf("expected error to save path %q, got %q", tt.err.path, perr.path)
				}

				if perr.Error() != tt.err.Error() {
					t.Fatalf("expected error %q, got %q", tt.err.Error(), perr.Error())
				}
			}
		})
	}
}
