package summarize

import "testing"

func TestStripSourceInjection(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "tail note removed",
			input: "提升開發效率與體驗（來源：[ai-souken.com]、[antigravity.google]）",
			want:  "提升開發效率與體驗",
		},
		{
			name:  "multiple notes removed",
			input: "第一段（來源：a）中間文字第二段（來源：b）",
			want:  "第一段中間文字第二段",
		},
		{
			name:  "no note unchanged",
			input: "純文字沒有來源備註",
			want:  "純文字沒有來源備註",
		},
		{
			name:  "no note trailing whitespace trimmed",
			input: "純文字沒有來源備註\n\n  ",
			want:  "純文字沒有來源備註",
		},
		{
			name:  "inline note within paragraph removed",
			input: "開頭（來源：[x.com]）結尾繼續寫",
			want:  "開頭結尾繼續寫",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripSourceInjection(tt.input)
			if got != tt.want {
				t.Errorf("stripSourceInjection(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
