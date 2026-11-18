package corpus

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestUppercaseFirstPreservesUnicode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "экспериментальная проверка", want: "Экспериментальная проверка"},
		{input: "analysis", want: "Analysis"},
		{input: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := uppercaseFirst(tt.input); got != tt.want {
				t.Fatalf("uppercaseFirst(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGeneratedTitlesAreValidUTF8(t *testing.T) {
	for _, publication := range Generate().Publications {
		if !utf8.ValidString(publication.Title) {
			t.Errorf("publication %s has invalid UTF-8 title %q", publication.ID, publication.Title)
		}
		if strings.ContainsRune(publication.Title, utf8.RuneError) {
			t.Errorf("publication %s contains Unicode replacement character: %q", publication.ID, publication.Title)
		}
	}
}

func TestReportedPublicationTitleIsGeneratedWithoutCorruption(t *testing.T) {
	want := "Экспериментальная проверка правовых практик"
	for _, publication := range Generate().Publications {
		if publication.Title == want {
			return
		}
	}
	t.Fatalf("generated corpus does not contain expected title %q", want)
}
