package lexer

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Roundaround/json5-go/token"
)

func New(source string) *Lexer {
	l := &Lexer{
		source: source,
		col:    -1,
	}
	l.readChar()
	return l
}

type Lexer struct {
	source  string
	pos     int
	readPos int
	line    int
	col     int
	ch      rune
	next    rune
	tokPos  tokenPos
}

type tokenPos struct {
	offset int
	line   int
	column int
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	inc := true

	l.skipWhitespace()

	l.tokPos = tokenPos{
		offset: l.pos,
		line:   l.line,
		column: l.col,
	}

	switch l.ch {
	case '{':
		tok = l.rtoken(token.LEFT_BRACE)
	case '}':
		tok = l.rtoken(token.RIGHT_BRACE)
	case '[':
		tok = l.rtoken(token.LEFT_BRACKET)
	case ']':
		tok = l.rtoken(token.RIGHT_BRACKET)
	case ',':
		tok = l.rtoken(token.COMMA)
	case ':':
		tok = l.rtoken(token.COLON)
	case 0:
		tok = l.rtoken(token.EOF)
	case '"', '\'':
		inc = false
		literal, err := l.readString()
		if err != nil {
			tok = l.rtoken(token.ILLEGAL)
		} else {
			tok = l.token(token.QUOTED_STRING, literal)
		}
	case '/':
		inc = false
		tok = l.readCommentToken()
	default:
		inc = false
		if isNumberStart(l.ch) {
			tok = l.readNumberToken()
		} else if isIdentifierStart(l.ch) {
			tok = l.readIdentifierToken()
		} else {
			tok = l.rtoken(token.ILLEGAL)
		}
	}

	if inc {
		l.readChar()
	}

	return tok
}

func (l *Lexer) rtoken(kind token.Kind) token.Token {
	return token.Token{
		Kind:    kind,
		Literal: string(l.ch),
		Offset:  l.tokPos.offset,
		Line:    l.tokPos.line + 1,
		Column:  l.tokPos.column + 1,
	}
}

func (l *Lexer) token(kind token.Kind, literal string) token.Token {
	return token.Token{
		Kind:    kind,
		Literal: literal,
		Offset:  l.tokPos.offset,
		Line:    l.tokPos.line + 1,
		Column:  l.tokPos.column + 1,
	}
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.source) {
		l.ch = 0
		l.col = 0
	} else {
		r, size := utf8.DecodeRuneInString(l.source[l.readPos:])
		l.ch = r
		l.pos = l.readPos
		l.readPos += size

		snext := 0
		if l.readPos >= len(l.source) {
			l.next = 0
		} else {
			l.next, snext = utf8.DecodeRuneInString(l.source[l.readPos:])
		}

		if isLineTerminator(l.ch) {
			l.line++
			l.col = -1
			if l.ch == '\r' && l.next == '\n' {
				// Treat CRLF as a single character
				l.readPos += snext
			}
		} else {
			l.col += len(string(l.ch))
		}
	}
}

func (l *Lexer) readString() (string, error) {
	q := l.ch
	pos := l.pos
	l.readChar()

	for l.ch != q {
		l.readChar()

		if l.ch == 0 {
			return "", errors.New("unterminated string")
		}

		if l.ch == '\\' {
			if isLineTerminator(l.next) {
				l.readChar()
			}
			continue
		}

		if isLineTerminator(l.ch) {
			return "", errors.New("unterminated string")
		}
	}

	l.readChar()
	return unescapeString(l.source[pos:l.pos])
}

func (l *Lexer) readCommentToken() token.Token {
	switch l.next {
	case '/':
		return l.token(token.LINE_COMMENT, l.readLineComment())
	case '*':
		return l.token(token.BLOCK_COMMENT, l.readBlockComment())
	default:
		return l.rtoken(token.ILLEGAL)
	}
}

func (l *Lexer) readLineComment() string {
	pos := l.pos
	for !isLineTerminator(l.ch) {
		l.readChar()
	}
	return l.source[pos:l.pos]
}

func (l *Lexer) readBlockComment() string {
	pos := l.pos
	for !(l.ch == '*' && l.next == '/') {
		l.readChar()
	}
	l.readChar()
	l.readChar()
	return l.source[pos:l.pos]
}

func (l *Lexer) readNumberToken() token.Token {
	sign := ""
	if l.ch == '-' || l.ch == '+' {
		sign = string(l.ch)
		l.readChar()
	}

	if l.ch == '0' && l.next == 'x' {
		return l.token(token.HEX_NUMBER, l.readHexNumber(sign))
	}

	return l.token(token.DECIMAL_NUMBER, l.readDecimalNumber(sign))
}

func (l *Lexer) readHexNumber(sign string) string {
	pos := l.pos
	l.readChar()
	l.readChar()
	for isHexDigit(l.ch) {
		l.readChar()
	}
	return sign + l.source[pos:l.pos]
}

func (l *Lexer) readDecimalNumber(sign string) string {
	pos := l.pos

	// integer part ([0-9]*)
	for isDigit(l.ch) {
		l.readChar()
	}

	// fraction part (\.[0-9]*)
	if l.ch == '.' {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// exponent part ([eE][+-]?[0-9]+)
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return sign + l.source[pos:l.pos]
}

func (l *Lexer) readIdentifierToken() token.Token {
	literal := l.readIdentifier()
	kind, ok := token.LookupKeyword(literal)
	if ok {
		return l.token(kind, literal)
	}
	return l.token(token.UNQUOTED_STRING, literal)
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isIdentifierPart(l.ch) {
		l.readChar()
	}
	return l.source[pos:l.pos]
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || isLineTerminator(ch)
}

func isLineTerminator(ch rune) bool {
	return ch == '\r' || ch == '\n' || ch == '\u2028' || ch == '\u2029'
}

func isNumberStart(ch rune) bool {
	return (ch >= '0' && ch <= '9') || ch == '-' || ch == '+' || ch == '.'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '$'
}

func isIdentifierPart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '$'
}

func unescapeString(s string) (string, error) {
	var buf strings.Builder
	pos := 0

	for pos < len(s) {
		r, size := utf8.DecodeRuneInString(s[pos:])
		if r == utf8.RuneError {
			return "", errors.New("invalid UTF-8 sequence")
		}

		pos += size

		if r == '\\' {
			pr, psize := utf8.DecodeRuneInString(s[pos:])
			if pr == utf8.RuneError {
				return "", errors.New("invalid UTF-8 sequence")
			}

			switch pr {
			case '\\':
				buf.WriteRune('\\')
				pos += psize
			case '/':
				buf.WriteRune('/')
				pos += psize
			case '"':
				buf.WriteRune('"')
				pos += psize
			case '\'':
				buf.WriteRune('\'')
				pos += psize
			case 'n':
				buf.WriteRune('\n')
				pos += psize
			case 'r':
				buf.WriteRune('\r')
				pos += psize
			case 't':
				buf.WriteRune('\t')
				pos += psize
			case 'b':
				buf.WriteRune('\b')
				pos += psize
			case 'f':
				buf.WriteRune('\f')
				pos += psize
			case 'u', 'U':
				// Convert \uXXXX to the actual rune
				hex := s[pos+1 : pos+5]
				ru, err := strconv.ParseUint(hex, 16, 32)
				if err != nil {
					return "", errors.New("invalid Unicode escape sequence")
				}
				buf.WriteRune(rune(ru))
				pos += psize + 4
			case '\n', '\r', '\u2028', '\u2029':
				buf.WriteRune(pr)
				pos += psize
			default:
				// Invalid escape sequence - leave as is
				buf.WriteRune(r)
			}
		} else {
			buf.WriteRune(r)
		}
	}

	return buf.String(), nil
}
