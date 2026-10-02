package encoder

import (
	"strings"
	"testing"
)

func TestEncoder(t *testing.T) {
	gen := New()

	code, err := gen.Encode()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	if len(code) != codeLength {
		t.Errorf(
			"expected code length %d, got %d",
			codeLength,
			len(code),
		)
	}

	for _, char := range code {
		if !strings.ContainsRune(alphabet, char) {
			t.Errorf("code contains invalid character: %q", char)
		}
	}
}
