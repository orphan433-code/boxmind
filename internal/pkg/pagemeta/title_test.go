package pagemeta

import "testing"

func TestCleanPageTitle(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{
			in:   "Ледяная стена — смотреть аниме онлайн",
			want: "Ледяная стена",
		},
		{
			in:   "Breaking Bad | AMC",
			want: "Breaking Bad",
		},
		{
			in:   "Normal title",
			want: "Normal title",
		},
		{
			in:   "- YouTube",
			want: "",
		},
		{
			in:   "Японский рэп микс - YouTube",
			want: "Японский рэп микс",
		},
		{
			in:   "YouTube",
			want: "",
		},
	}

	for _, tt := range tests {
		if got := CleanPageTitle(tt.in); got != tt.want {
			t.Fatalf("CleanPageTitle(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
