package gemini

import "testing"

func TestExtractJSONTrailingJunk(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "clean object",
			raw:  `{"title":"A"}`,
			want: `{"title":"A"}`,
		},
		{
			name: "trailing backtick",
			raw:  "{\"title\":\"A\"}`",
			want: `{"title":"A"}`,
		},
		{
			name: "trailing code fence and prose",
			raw:  "{\"title\":\"A\",\"tags\":[\"x\"]}\n```\nНадеюсь, помог!",
			want: `{"title":"A","tags":["x"]}`,
		},
		{
			name: "leading json fence",
			raw:  "```json\n{\"title\":\"A\"}\n```",
			want: `{"title":"A"}`,
		},
		{
			name: "nested object",
			raw:  `{"a":{"b":1}} extra`,
			want: `{"a":{"b":1}}`,
		},
		{
			name: "brace inside string",
			raw:  `{"title":"a } b"} trailing`,
			want: `{"title":"a } b"}`,
		},
		{
			name: "escaped quote inside string",
			raw:  `{"title":"a \" } b"} trailing`,
			want: `{"title":"a \" } b"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractJSON(tt.raw); got != tt.want {
				t.Fatalf("extractJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}
