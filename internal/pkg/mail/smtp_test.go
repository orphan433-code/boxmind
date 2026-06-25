package mail

import (
	"strings"
	"testing"
)

func TestBuildMessage(t *testing.T) {
	t.Parallel()

	msg := string(buildMessage(
		"Boxmind <noreply@boxmind.test>",
		"user@example.com",
		"Код для входа в Boxmind",
		"123456",
	))

	if !strings.Contains(msg, "To: user@example.com") {
		t.Fatalf("missing recipient: %q", msg)
	}
	if !strings.Contains(msg, "From: Boxmind <noreply@boxmind.test>") {
		t.Fatalf("missing from: %q", msg)
	}
	if !strings.Contains(msg, "123456") {
		t.Fatalf("missing body: %q", msg)
	}
}

func TestParseAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		want string
	}{
		{raw: "noreply@boxmind.test", want: "noreply@boxmind.test"},
		{raw: "Boxmind <noreply@boxmind.test>", want: "noreply@boxmind.test"},
	}

	for _, tt := range tests {
		got, err := parseAddress(tt.raw)
		if err != nil {
			t.Fatalf("parseAddress(%q): %v", tt.raw, err)
		}
		if got != tt.want {
			t.Fatalf("parseAddress(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}
