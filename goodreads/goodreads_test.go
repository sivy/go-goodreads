package goodreads

import (
	"testing"
)

func TestClient(t *testing.T) {
	client := NewClient("some-api-key")
	if client == nil {
		t.Fail()
	}
}
