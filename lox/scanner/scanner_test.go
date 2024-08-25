package scanner

import (
	"testing"

	"github.com/Tan2Pi/golox/lox"
	"github.com/Tan2Pi/golox/lox/tokens"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals // okay to use in tests
var testData = `
// Unicode characters are allowed in comments.
//
// Latin 1 Supplement: £§¶ÜÞ
// Latin Extended-A: ĐĦŋœ
// Latin Extended-B: ƂƢƩǁ
// Other stuff: ឃᢆ᯽₪ℜ↩⊗┺░
// Emoji: ☃☺♣

print "ok"; // expect: ok
`

func TestScannerUnicode(t *testing.T) {
	t.Setenv(lox.EnvDebug, "true")
	a := assert.New(t)
	t.Logf("source len: %v\n", len([]rune(testData)))
	s := New(testData)
	scannedTokens := s.ScanTokens()
	a.NotEmpty(scannedTokens)
	a.Equal(tokens.Print.String(), scannedTokens[0].Type.String())
	expectedTokens := []tokens.Token{
		{
			Type:    tokens.Print,
			Lexeme:  "print",
			Literal: nil,
			Line:    0,
		},
		{
			Type:    tokens.String,
			Lexeme:  "\"ok\"",
			Literal: "ok",
			Line:    0,
		},
		{
			Type:    tokens.Semicolon,
			Lexeme:  ";",
			Literal: nil,
			Line:    0,
		},
		{
			Type:    tokens.EOF,
			Lexeme:  "",
			Literal: nil,
			Line:    0,
		},
	}

	require.Len(t, scannedTokens, len(expectedTokens))

	for i, tt := range scannedTokens {
		t.Logf("token at %v = %v", i, tt)
		assert.Equal(t, expectedTokens[i].Type, tt.Type)
		assert.Equal(t, expectedTokens[i].Lexeme, tt.Lexeme)
		assert.Equal(t, expectedTokens[i].Literal, tt.Literal)
	}
}
