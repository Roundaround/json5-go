package lexer

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Roundaround/json5-go/token"
)

type expectedToken struct {
	kind    token.Kind
	literal string
	line    int
	column  int
}

func (e expectedToken) String() string {
	return fmt.Sprintf("%s %q", e.kind, e.literal)
}

func TestLexer_NextToken(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		linebreak string
	}{
		{
			name:      "lf",
			filename:  "testdata/test.json5",
			linebreak: "\n",
		},
		{
			name:      "crlf",
			filename:  "testdata/crlf.json5",
			linebreak: "\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := []expectedToken{
				{token.LEFT_BRACE, "{", 1, 1},
				{token.LINE_COMMENT, "// line comment", 2, 3},
				{token.BLOCK_COMMENT, "/* block comment */", 3, 3},
				{token.UNQUOTED_STRING, "unquoted", 4, 3},
				{token.COLON, ":", 4, 11},
				{token.QUOTED_STRING, "\"and you can quote me on that\"", 4, 13},
				{token.COMMA, ",", 4, 43},
				{token.UNQUOTED_STRING, "singleQuotes", 5, 3},
				{token.COLON, ":", 5, 15},
				{token.QUOTED_STRING, "'I can use \"double quotes\" here'", 5, 17},
				{token.COMMA, ",", 5, 49},
				{token.UNQUOTED_STRING, "escapeSequences", 6, 3},
				{token.COLON, ":", 6, 18},
				{token.QUOTED_STRING, "\"\n\r\t\b\fǨ\"", 6, 20},
				{token.COMMA, ",", 6, 38},
				{token.UNQUOTED_STRING, "lineBreaks", 7, 3},
				{token.COLON, ":", 7, 13},
				{token.QUOTED_STRING, fmt.Sprintf("\"Look, Mom! %sNo \\n's!\"", tt.linebreak), 7, 15},
				{token.COMMA, ",", 8, 11},
				{token.UNQUOTED_STRING, "hexadecimal", 9, 3},
				{token.COLON, ":", 9, 14},
				{token.HEX_NUMBER, "0xdecaf", 9, 16},
				{token.COMMA, ",", 9, 23},
				{token.UNQUOTED_STRING, "leadingDecimalPoint", 10, 3},
				{token.COLON, ":", 10, 22},
				{token.DECIMAL_NUMBER, ".8675309", 10, 24},
				{token.COMMA, ",", 10, 32},
				{token.UNQUOTED_STRING, "andTrailing", 10, 34},
				{token.COLON, ":", 10, 45},
				{token.DECIMAL_NUMBER, "8675309.", 10, 47},
				{token.COMMA, ",", 10, 55},
				{token.UNQUOTED_STRING, "positiveSign", 11, 3},
				{token.COLON, ":", 11, 15},
				{token.DECIMAL_NUMBER, "+1", 11, 17},
				{token.COMMA, ",", 11, 19},
				{token.UNQUOTED_STRING, "trailingComma", 12, 3},
				{token.COLON, ":", 12, 16},
				{token.QUOTED_STRING, "'in objects'", 12, 18},
				{token.COMMA, ",", 12, 30},
				{token.UNQUOTED_STRING, "andIn", 12, 32},
				{token.COLON, ":", 12, 37},
				{token.LEFT_BRACKET, "[", 12, 39},
				{token.QUOTED_STRING, "'arrays'", 12, 40},
				{token.COMMA, ",", 12, 48},
				{token.RIGHT_BRACKET, "]", 12, 49},
				{token.COMMA, ",", 12, 50},
				{token.QUOTED_STRING, "\"backwardsCompatible\"", 13, 3},
				{token.COLON, ":", 13, 24},
				{token.QUOTED_STRING, "\"with JSON\"", 13, 26},
				{token.COMMA, ",", 13, 37},
				{token.RIGHT_BRACE, "}", 14, 1},
				{token.EOF, string(byte(0)), 15, 1},
			}

			source, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			lexer := New(string(source))
			failed := 0
			messages := []struct {
				i   int
				msg string
			}{}

			for i := range expected {
				tok := lexer.NextToken()
				if tok.Kind != expected[i].kind || tok.Literal != expected[i].literal || tok.Line != expected[i].line || tok.Column != expected[i].column {
					messages = append(messages, struct {
						i   int
						msg string
					}{i, fmt.Sprintf("expected %s (Ln %d, Col %d), got %s (Ln %d, Col %d)", expected[i].String(), expected[i].line, expected[i].column, tok.String(), tok.Line, tok.Column)})
					failed++
				}

				if failed >= 5 {
					break
				}
			}

			if failed > 0 {
				var msg strings.Builder
				msg.WriteString(fmt.Sprintf("%d or more tokens incorrect:\n", failed))
				for _, message := range messages {
					msg.WriteString(fmt.Sprintf("  %d: %s\n", message.i, message.msg))
				}
				t.Error(msg.String())
			}
		})
	}
}
