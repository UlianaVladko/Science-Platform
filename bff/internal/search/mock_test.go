package search

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestMakeSnippetPreservesUTF8(t *testing.T) {
	input := strings.Repeat("исследование данных ", 8) + "Экспериментальная проверка правовых практик" + strings.Repeat(" научный текст", 8)
	got := makeSnippet(input, "ПРАВОВЫХ ПРАКТИК")
	if got == "" {
		t.Fatal("makeSnippet returned an empty snippet for a matching query")
	}
	if !utf8.ValidString(got) {
		t.Fatalf("makeSnippet returned invalid UTF-8: %q", got)
	}
	if strings.ContainsRune(got, utf8.RuneError) {
		t.Fatalf("makeSnippet introduced a replacement character: %q", got)
	}
	if !strings.Contains(got, "правовых практик") {
		t.Fatalf("snippet does not contain the matched text: %q", got)
	}
}

func TestMakeSnippetReturnsEmptyForMissingNeedle(t *testing.T) {
	if got := makeSnippet("Русский текст", "missing"); got != "" {
		t.Fatalf("makeSnippet = %q, want empty string", got)
	}
}
