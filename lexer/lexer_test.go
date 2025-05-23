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
}

func (e expectedToken) String() string {
	return fmt.Sprintf("%s %q", e.kind, e.literal)
}

func TestLexer_NextToken_testdata_test(t *testing.T) {
	expected := []expectedToken{
		{token.LEFT_BRACE, "{"},
		{token.LINE_COMMENT, "// line comment"},
		{token.BLOCK_COMMENT, "/* block comment */"},
		{token.UNQUOTED_STRING, "unquoted"},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "\"and you can quote me on that\""},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "singleQuotes"},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "'I can use \"double quotes\" here'"},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "escapeSequences"},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "\"\n\r\t\b\fǨ\""},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "lineBreaks"},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "\"Look, Mom! \nNo \\n's!\""},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "hexadecimal"},
		{token.COLON, ":"},
		{token.HEX_NUMBER, "0xdecaf"},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "leadingDecimalPoint"},
		{token.COLON, ":"},
		{token.DECIMAL_NUMBER, ".8675309"},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "andTrailing"},
		{token.COLON, ":"},
		{token.DECIMAL_NUMBER, "8675309."},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "positiveSign"},
		{token.COLON, ":"},
		{token.DECIMAL_NUMBER, "+1"},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "trailingComma"},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "'in objects'"},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "andIn"},
		{token.COLON, ":"},
		{token.LEFT_BRACKET, "["},
		{token.QUOTED_STRING, "'arrays'"},
		{token.COMMA, ","},
		{token.RIGHT_BRACKET, "]"},
		{token.COMMA, ","},
		{token.QUOTED_STRING, "\"backwardsCompatible\""},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "\"with JSON\""},
		{token.COMMA, ","},
		{token.RIGHT_BRACE, "}"},
		{token.EOF, string(byte(0))},
	}

	source, err := os.ReadFile("testdata/test.json5")
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
		if tok.Kind != expected[i].kind || tok.Literal != expected[i].literal {
			messages = append(messages, struct {
				i   int
				msg string
			}{i, fmt.Sprintf("expected %s, got %s", expected[i].String(), tok.String())})
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
}

func TestLexer_NextToken_testdata_crlf(t *testing.T) {
	expected := []expectedToken{
		{token.LEFT_BRACE, "{"},
		{token.LINE_COMMENT, "// line comment"},
		{token.BLOCK_COMMENT, "/* block comment */"},
		{token.UNQUOTED_STRING, "unquoted"},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "\"and you can quote me on that\""},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "singleQuotes"},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "'I can use \"double quotes\" here'"},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "escapeSequences"},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "\"\n\r\t\b\fǨ\""},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "lineBreaks"},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "\"Look, Mom! \r\nNo \\n's!\""},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "hexadecimal"},
		{token.COLON, ":"},
		{token.HEX_NUMBER, "0xdecaf"},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "leadingDecimalPoint"},
		{token.COLON, ":"},
		{token.DECIMAL_NUMBER, ".8675309"},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "andTrailing"},
		{token.COLON, ":"},
		{token.DECIMAL_NUMBER, "8675309."},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "positiveSign"},
		{token.COLON, ":"},
		{token.DECIMAL_NUMBER, "+1"},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "trailingComma"},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "'in objects'"},
		{token.COMMA, ","},
		{token.UNQUOTED_STRING, "andIn"},
		{token.COLON, ":"},
		{token.LEFT_BRACKET, "["},
		{token.QUOTED_STRING, "'arrays'"},
		{token.COMMA, ","},
		{token.RIGHT_BRACKET, "]"},
		{token.COMMA, ","},
		{token.QUOTED_STRING, "\"backwardsCompatible\""},
		{token.COLON, ":"},
		{token.QUOTED_STRING, "\"with JSON\""},
		{token.COMMA, ","},
		{token.RIGHT_BRACE, "}"},
		{token.EOF, string(byte(0))},
	}

	source, err := os.ReadFile("testdata/crlf.json5")
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
		if tok.Kind != expected[i].kind || tok.Literal != expected[i].literal {
			messages = append(messages, struct {
				i   int
				msg string
			}{i, fmt.Sprintf("expected %s, got %s", expected[i].String(), tok.String())})
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
}
