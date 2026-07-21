package shared

import "testing"

func TestNormalizeNumericID(t *testing.T) {
	id, ok := NormalizeNumericID(" 001 ")
	if !ok || id != "1" {
		t.Fatalf("expected normalized positive id, got id=%q ok=%v", id, ok)
	}

	for _, value := range []string{"", "0", "-1", "abc"} {
		if id, ok := NormalizeNumericID(value); ok || id != "" {
			t.Fatalf("expected invalid id for %q, got id=%q ok=%v", value, id, ok)
		}
	}
}

func TestNormalizePagination(t *testing.T) {
	limit, offset := NormalizePagination(0, -5, 20, 100)
	if limit != 20 || offset != 0 {
		t.Fatalf("expected defaults, got limit=%d offset=%d", limit, offset)
	}

	limit, offset = NormalizePagination(150, 10, 20, 100)
	if limit != 100 || offset != 10 {
		t.Fatalf("expected capped limit, got limit=%d offset=%d", limit, offset)
	}
}
