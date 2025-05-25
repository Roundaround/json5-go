package path

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func newParser(source string) *parser {
	p := &parser{source: source}
	p.readChar()
	return p
}

type parser struct {
	source  string
	pos     int
	readPos int
	ch      rune
	next    rune
}

func (p *parser) nextSegment() (*Segment, error) {
	switch p.ch {
	case 0:
		return nil, nil
	case '.':
		if p.pos == 0 {
			return nil, p.errf("expected key or index, got %s", quotech(p.ch))
		}
		if !isIdentifierStart(p.next) {
			return nil, p.erratf(p.readPos, "expected key, got %s", quotech(p.next))
		}
		p.readChar() // skip '.'
		return Key(p.readIdentifier()), nil
	case '[':
		// TODO: Can I pull out the common code here?

		if isDigit(p.next) {
			p.readChar() // skip '['
			num, err := p.readNumber()
			if err != nil {
				return nil, err
			}
			if p.ch != ']' {
				return nil, p.errf("expected ']', got %s", quotech(p.ch))
			}
			p.readChar() // skip ']'
			return Index(num), nil
		}

		if p.next == '"' || p.next == '\'' {
			q := p.next
			p.readChar() // skip '['
			p.readChar() // skip quote
			ident := p.readIdentifier()
			if p.ch != q {
				return nil, p.errf("expected %s, got %s", quotech(q), quotech(p.ch))
			}
			p.readChar() // skip quote
			if p.ch != ']' {
				return nil, p.errf("expected ']', got %s", quotech(p.ch))
			}
			p.readChar() // skip ']'
			return Key(ident), nil
		}

		if isIdentifierStart(p.next) {
			// TODO: Do I actually want to allow unquoted keys inside brackets?
			p.readChar() // skip '['
			ident := p.readIdentifier()
			if p.ch != ']' {
				return nil, p.errf("expected ']', got %s", quotech(p.ch))
			}
			p.readChar() // skip ']'
			return Key(ident), nil
		}

		return nil, p.errf("expected key or index, got %s", quotech(p.ch))
	default:
		if isIdentifierStart(p.ch) {
			return Key(p.readIdentifier()), nil
		}

		var msg string
		if isDigit(p.ch) {
			if isDigit(p.next) {
				msg = "indexes must be in square brackets"
			} else if isIdentifierPart(p.next) {
				msg = "keys cannot start with a digit"
			}
		}
		if msg == "" {
			msg = "expected key or index"
		}

		return nil, p.errf("%s, got %s", msg, quotech(p.ch))
	}
}

func (p *parser) readChar() {
	if p.readPos >= len(p.source) {
		p.pos = p.readPos
		p.ch = 0
		return
	}

	r, size := utf8.DecodeRuneInString(p.source[p.readPos:])
	p.ch = r
	p.pos = p.readPos
	p.readPos += size

	if p.readPos >= len(p.source) {
		p.next = 0
	} else {
		p.next, _ = utf8.DecodeRuneInString(p.source[p.readPos:])
	}
}

func (p *parser) errf(format string, args ...any) *PathError {
	return p.erratf(p.pos, format, args...)
}

func (p *parser) erratf(pos int, format string, args ...any) *PathError {
	return &PathError{fmt.Errorf(format, args...), pos, p.source}
}

func (p *parser) readIdentifier() string {
	pos := p.pos
	for isIdentifierPart(p.ch) {
		p.readChar()
	}
	return p.source[pos:p.pos]
}

func (p *parser) readNumber() (int, error) {
	pos := p.pos
	for isDigit(p.ch) {
		p.readChar()
	}

	str := p.source[pos:p.pos]
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0, p.erratf(pos, "expected index, got %q", str)
	}
	return num, nil
}

type PathError struct {
	err  error
	pos  int
	path string
}

func (e *PathError) Error() string {
	return fmt.Sprintf("invalid path: %v", e.err)
}

func (e *PathError) Unwrap() []error {
	return []error{e.err}
}

func (e *PathError) Annotate() string {
	msg := "invalid path: "
	pad := strings.Repeat(" ", len(msg)+e.pos)
	return fmt.Sprintf("%s%s\n%s^ %v", msg, e.path, pad, e.err)
}

func isIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '$'
}

func isIdentifierPart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '$'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func quotech(ch rune) string {
	if ch == '\'' {
		return fmt.Sprintf("%q", string(ch))
	}
	return fmt.Sprintf("%q", ch)
}
