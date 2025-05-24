package path

import (
	"fmt"
	"slices"
	"strings"
)

const Root = "$"

type Path struct {
	segments []Segment
}

type Segment struct {
	key   string
	index int
}

type number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

type stringish interface {
	AsString() string
}

type integerish interface {
	AsInteger() int
}

func New(segments ...any) (*Path, error) {
	processed := make([]Segment, 0, len(segments))
	for _, segment := range segments {
		s, err := NewSegment(segment)
		if err != nil {
			return nil, err
		}
		processed = append(processed, *s)
	}
	if len(processed) > 0 && processed[0].key == Root {
		// Don't actually store the root segment
		processed = processed[1:]
	}
	return &Path{segments: processed}, nil
}

func Must(segments ...any) *Path {
	p, err := New(segments...)
	if err != nil {
		panic(err)
	}
	return p
}

func Parse(source string) (*Path, error) {
	p := newParser(source)
	segments := make([]Segment, 0)
	for {
		segment, err := p.nextSegment()
		if err != nil {
			return nil, err
		}
		if segment == nil {
			break
		}
		segments = append(segments, *segment)
	}
	if len(segments) > 0 && segments[0].key == Root {
		// Don't actually store the root segment
		segments = segments[1:]
	}
	return &Path{segments: segments}, nil
}

func (p *Path) Segments() []Segment {
	return p.segments
}

func (p *Path) Prepend(segments ...Segment) {
	p.segments = append(segments, p.segments...)
}

func (p *Path) Append(segments ...Segment) {
	p.segments = append(p.segments, segments...)
}

func (p *Path) String() string {
	if p == nil {
		return ""
	}

	var b strings.Builder
	for i, segment := range p.segments {
		if segment.key != "" && i > 0 {
			b.WriteString(".")
		}
		b.WriteString(segment.String())
	}
	return b.String()
}

func (p *Path) Equals(other *Path) bool {
	if p == nil || other == nil {
		return p == other
	}
	return slices.Equal(p.segments, other.segments)
}

func Key(key string) *Segment {
	return &Segment{key: key, index: -1}
}

func Index[N number](index N) *Segment {
	return &Segment{index: int(index), key: ""}
}

func NewSegment(segment any) (*Segment, error) {
	switch v := segment.(type) {
	case string:
		return Key(v), nil
	case int:
		return Index(v), nil
	case int8:
		return Index(v), nil
	case int16:
		return Index(v), nil
	case int32:
		return Index(v), nil
	case int64:
		return Index(v), nil
	case uint:
		return Index(v), nil
	case uint8:
		return Index(v), nil
	case uint16:
		return Index(v), nil
	case uint32:
		return Index(v), nil
	case uint64:
		return Index(v), nil
	case float32:
		return Index(v), nil
	case float64:
		return Index(v), nil
	case Segment:
		return &v, nil
	case *Segment:
		return v, nil
	}

	if stringish, ok := segment.(stringish); ok {
		return Key(stringish.AsString()), nil
	}

	if integerish, ok := segment.(integerish); ok {
		return Index(integerish.AsInteger()), nil
	}

	return nil, fmt.Errorf("invalid segment type: %T", segment)
}

func (s *Segment) String() string {
	if s.index != -1 {
		return fmt.Sprintf("[%d]", s.index)
	}
	return s.key
}
