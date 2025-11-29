package response

import "testing"

func TestOK(t *testing.T) {
	resp := OK("data")

	if resp.Code != 0 {
		t.Errorf("Expected code 0, got %d", resp.Code)
	}

	if resp.Message != "ok" {
		t.Errorf("Expected 'ok', got '%s'", resp.Message)
	}
}

func TestPage(t *testing.T) {
	data := []int{1, 2, 3}
	resp := Page(data, 100, 1, 20)

	if resp.Total != 100 {
		t.Errorf("Expected total 100, got %d", resp.Total)
	}

	if resp.Page != 1 {
		t.Errorf("Expected page 1, got %d", resp.Page)
	}
}

func TestWithRequestID(t *testing.T) {
	resp := OK("data").WithRequestID("req-123")

	if resp.RequestID != "req-123" {
		t.Errorf("Expected 'req-123', got '%s'", resp.RequestID)
	}
}
