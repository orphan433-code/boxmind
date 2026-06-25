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
	}

	for _, tt := range tests {
		if got := CleanPageTitle(tt.in); got != tt.want {
			t.Fatalf("CleanPageTitle(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
