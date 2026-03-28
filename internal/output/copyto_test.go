package output

import "testing"

func TestExpandTemplate(t *testing.T) {
	vars := CopyToVars{
		OutputDir: "/vault/references",
		Date:      "2026-03-28",
		DateAdded: "2026-03-25",
		Title:     "Test Article Title",
		SHA8:      "a1b2c3d4",
		Source:    "safari",
		Domain:    "example.com",
		DomainDir: "example_com",
		Type:      "summary",
	}

	got := ExpandTemplate("{output_dir}/by-source/{source}/{domain_dir}", vars)
	want := "/vault/references/by-source/safari/example_com"
	if got != want {
		t.Errorf("ExpandTemplate path = %q, want %q", got, want)
	}

	got = ExpandTemplate("{date}__{title}__{type}.md", vars)
	want = "2026-03-28__Test Article Title__summary.md"
	if got != want {
		t.Errorf("ExpandTemplate filename = %q, want %q", got, want)
	}
}

func TestSanitizeTitleForDisplay(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello/World: Test", "Hello World Test"},
		{"日本語テスト", "日本語テスト"},
		{`A "quoted" <title>`, "A quoted title"},
	}
	for _, tt := range tests {
		got := SanitizeTitleForDisplay(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeTitleForDisplay(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
