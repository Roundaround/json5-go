package lexer

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Roundaround/json5-go/token"
)

func New(source string) *Lexer {
	l := &Lexer{source: source}
	l.readChar()
	return l
}

type Lexer struct {
	source       string
	position     int
	readPosition int
	ch           rune
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	inc := true

	l.skipWhitespace()

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
		Offset:  l.position,
	}
}

func (l *Lexer) token(kind token.Kind, literal string) token.Token {
	return token.Token{
		Kind:    kind,
		Literal: literal,
		Offset:  l.position,
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.source) {
		l.ch = 0
	} else {
		r, size := utf8.DecodeRuneInString(l.source[l.readPosition:])
		l.ch = r
		l.position = l.readPosition
		l.readPosition += size
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.source) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.readPosition:])
	return r
}

func (l *Lexer) readString() (string, error) {
	q := l.ch
	pos := l.position
	l.readChar()

	for l.ch != q {
		l.readChar()

		if l.ch == 0 {
			return "", errors.New("unterminated string")
		}

		if l.ch == '\\' {
			peek := l.peekChar()
			if peek == '\r' {
				l.readChar()
				if l.peekChar() == '\n' {
					l.readChar()
				}
			}
			if peek == '\n' || peek == '\u2028' || peek == '\u2029' {
				l.readChar()
			}
			continue
		}

		if isLineTerminator(l.ch) {
			return "", errors.New("unterminated string")
		}
	}

	l.readChar()
	return unescapeString(l.source[pos:l.position])
}

func (l *Lexer) readCommentToken() token.Token {
	switch l.peekChar() {
	case '/':
		return l.token(token.LINE_COMMENT, l.readLineComment())
	case '*':
		return l.token(token.BLOCK_COMMENT, l.readBlockComment())
	default:
		return l.rtoken(token.ILLEGAL)
	}
}

func (l *Lexer) readLineComment() string {
	pos := l.position
	for !isLineTerminator(l.ch) {
		l.readChar()
	}
	return l.source[pos:l.position]
}

func (l *Lexer) readBlockComment() string {
	pos := l.position
	for !(l.ch == '*' && l.peekChar() == '/') {
		l.readChar()
	}
	l.readChar()
	l.readChar()
	return l.source[pos:l.position]
}

func (l *Lexer) readNumberToken() token.Token {
	sign := ""
	if l.ch == '-' || l.ch == '+' {
		sign = string(l.ch)
		l.readChar()
	}

	if l.ch == '0' && l.peekChar() == 'x' {
		return l.token(token.HEX_NUMBER, l.readHexNumber(sign))
	}

	return l.token(token.DECIMAL_NUMBER, l.readDecimalNumber(sign))
}

func (l *Lexer) readHexNumber(sign string) string {
	pos := l.position
	l.readChar()
	l.readChar()
	for isHexDigit(l.ch) {
		l.readChar()
	}
	return sign + l.source[pos:l.position]
}

func (l *Lexer) readDecimalNumber(sign string) string {
	pos := l.position

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

	return sign + l.source[pos:l.position]
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
	pos := l.position
	for isIdentifierPart(l.ch) {
		l.readChar()
	}
	return l.source[pos:l.position]
}

func isLineTerminator(ch rune) bool {
	return ch == '\r' || ch == '\n' || ch == '\u2028' || ch == '\u2029'
}

func countLineTerminatorRunes(ch rune, peek rune) int {
	switch ch {
	case '\r':
		if peek == '\n' {
			return 2
		}
		return 1
	case '\n', '\u2028', '\u2029':
		return 1
	default:
		return 0
	}
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
