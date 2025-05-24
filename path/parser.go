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

type pathError struct {
	pos  int
	msg  string
	path string
}

func (e *pathError) Error() string {
	msg := "invalid path: "
	pad := strings.Repeat(" ", len(msg)+e.pos)
	return fmt.Sprintf("%s%s\n%s^ %s", msg, e.path, pad, e.msg)
}

func (p *parser) nextSegment() (*Segment, error) {
	switch p.ch {
	case 0:
		return nil, nil
	case '.':
		if p.pos == 0 {
			return nil, &pathError{p.pos, "expected key or index, got .", p.source}
		}
		if !isIdentifierStart(p.next) {
			return nil, &pathError{p.pos, fmt.Sprintf("expected key, got %q", p.next), p.source}
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
				return nil, &pathError{p.pos, fmt.Sprintf("expected ], got %q", p.ch), p.source}
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
				return nil, &pathError{p.pos, fmt.Sprintf("expected %q, got %q", q, p.ch), p.source}
			}
			p.readChar() // skip quote
			if p.ch != ']' {
				return nil, &pathError{p.pos, fmt.Sprintf("expected ], got %q", p.ch), p.source}
			}
			p.readChar() // skip ']'
			return Key(ident), nil
		}

		if isIdentifierStart(p.next) {
			// TODO: Do I actually want to allow unquoted keys inside brackets?
			p.readChar() // skip '['
			ident := p.readIdentifier()
			if p.ch != ']' {
				return nil, &pathError{p.pos, fmt.Sprintf("expected ], got %q", p.ch), p.source}
			}
			p.readChar() // skip ']'
			return Key(ident), nil
		}

		return nil, &pathError{p.pos, fmt.Sprintf("expected key or index, got %q", p.ch), p.source}
	default:
		if isIdentifierStart(p.ch) {
			return Key(p.readIdentifier()), nil
		}
		return nil, &pathError{p.pos, fmt.Sprintf("expected key or index, got %q", p.ch), p.source}
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
		return 0, &pathError{pos, fmt.Sprintf("expected index, got %q", str), p.source}
	}
	return num, nil
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
