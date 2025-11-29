package utils

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- String Tests ---

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"camelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"simple", "simple"},
		{"HTTPServer", "h_t_t_p_server"},
	}

	for _, tt := range tests {
		result := ToSnakeCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToSnakeCase(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"snake_case", "snakeCase"},
		{"kebab-case", "kebabCase"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		result := ToCamelCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToCamelCase(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestMaskPhone(t *testing.T) {
	result := MaskPhone("13812345678")
	if result != "138****5678" {
		t.Errorf("MaskPhone = %s, want 138****5678", result)
	}
}

func TestMaskEmail(t *testing.T) {
	result := MaskEmail("test@example.com")
	if result != "t***@example.com" {
		t.Errorf("MaskEmail = %s, want t***@example.com", result)
	}
}

// --- Slice Tests ---

func TestContains(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	if !Contains(slice, 3) {
		t.Error("Contains should return true for 3")
	}
	if Contains(slice, 6) {
		t.Error("Contains should return false for 6")
	}
}

func TestUnique(t *testing.T) {
	slice := []int{1, 2, 2, 3, 3, 3}
	result := Unique(slice)
	if len(result) != 3 {
		t.Errorf("Unique length = %d, want 3", len(result))
	}
}

func TestFilter(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := Filter(slice, func(n int) bool { return n%2 == 0 })
	if len(result) != 2 {
		t.Errorf("Filter length = %d, want 2", len(result))
	}
}

func TestMap(t *testing.T) {
	slice := []int{1, 2, 3}
	result := Map(slice, func(n int) int { return n * 2 })
	expected := []int{2, 4, 6}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Map[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestGroupBy(t *testing.T) {
	type Item struct {
		Category string
		Name     string
	}
	items := []Item{
		{"A", "a1"}, {"A", "a2"}, {"B", "b1"},
	}
	result := GroupBy(items, func(i Item) string { return i.Category })
	if len(result["A"]) != 2 {
		t.Errorf("GroupBy A = %d, want 2", len(result["A"]))
	}
}

func TestChunk(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := Chunk(slice, 2)
	if len(result) != 3 {
		t.Errorf("Chunk length = %d, want 3", len(result))
	}
}

// --- Map Tests ---

func TestKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	keys := Keys(m)
	if len(keys) != 2 {
		t.Errorf("Keys length = %d, want 2", len(keys))
	}
}

func TestMerge(t *testing.T) {
	m1 := map[string]int{"a": 1}
	m2 := map[string]int{"b": 2, "a": 3}
	result := Merge(m1, m2)
	if result["a"] != 3 {
		t.Errorf("Merge a = %d, want 3", result["a"])
	}
	if result["b"] != 2 {
		t.Errorf("Merge b = %d, want 2", result["b"])
	}
}

// --- Ptr Tests ---

func TestPtr(t *testing.T) {
	p := Ptr(42)
	if *p != 42 {
		t.Errorf("Ptr = %d, want 42", *p)
	}
}

func TestValOr(t *testing.T) {
	var p *int
	result := ValOr(p, 10)
	if result != 10 {
		t.Errorf("ValOr = %d, want 10", result)
	}

	v := 42
	result = ValOr(&v, 10)
	if result != 42 {
		t.Errorf("ValOr = %d, want 42", result)
	}
}

func TestOptional(t *testing.T) {
	opt := Some(42)
	if !opt.IsPresent() {
		t.Error("Some should be present")
	}
	if opt.Get() != 42 {
		t.Errorf("Get = %d, want 42", opt.Get())
	}

	none := None[int]()
	if none.IsPresent() {
		t.Error("None should not be present")
	}
	if none.OrElse(10) != 10 {
		t.Errorf("OrElse = %d, want 10", none.OrElse(10))
	}
}

// --- Time Tests ---

func TestStartOfDay(t *testing.T) {
	now := time.Now()
	start := StartOfDay(now)
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Error("StartOfDay should be 00:00:00")
	}
}

func TestDaysBetween(t *testing.T) {
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	days := DaysBetween(t1, t2)
	if days != 9 {
		t.Errorf("DaysBetween = %d, want 9", days)
	}
}

// --- Number Tests ---

func TestMinMax(t *testing.T) {
	if Min(1, 2) != 1 {
		t.Error("Min(1,2) should be 1")
	}
	if Max(1, 2) != 2 {
		t.Error("Max(1,2) should be 2")
	}
}

func TestClamp(t *testing.T) {
	if Clamp(5, 0, 10) != 5 {
		t.Error("Clamp(5,0,10) should be 5")
	}
	if Clamp(-1, 0, 10) != 0 {
		t.Error("Clamp(-1,0,10) should be 0")
	}
	if Clamp(15, 0, 10) != 10 {
		t.Error("Clamp(15,0,10) should be 10")
	}
}

func TestSum(t *testing.T) {
	result := Sum(1, 2, 3, 4, 5)
	if result != 15 {
		t.Errorf("Sum = %d, want 15", result)
	}
}

// --- Async Tests ---

func TestRetry(t *testing.T) {
	attempts := 0
	err := Retry(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return errors.New("fail")
		}
		return nil
	}, RetryConfig{MaxAttempts: 5, Delay: time.Millisecond})

	if err != nil {
		t.Errorf("Retry should succeed, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("Retry attempts = %d, want 3", attempts)
	}
}

func TestParallel(t *testing.T) {
	results := make([]int, 3)
	err := Parallel(
		func() error { results[0] = 1; return nil },
		func() error { results[1] = 2; return nil },
		func() error { results[2] = 3; return nil },
	)

	if err != nil {
		t.Errorf("Parallel error: %v", err)
	}
	if results[0] != 1 || results[1] != 2 || results[2] != 3 {
		t.Errorf("Parallel results = %v", results)
	}
}

// --- ID Tests ---

func TestUUID(t *testing.T) {
	id := UUID()
	if len(id) != 36 {
		t.Errorf("UUID length = %d, want 36", len(id))
	}
}

func TestSnowflake(t *testing.T) {
	id1 := Snowflake()
	id2 := Snowflake()
	if id1 >= id2 {
		t.Error("Snowflake should be increasing")
	}
}

func TestShortID(t *testing.T) {
	id := ShortID(10)
	if len(id) != 10 {
		t.Errorf("ShortID length = %d, want 10", len(id))
	}
}

func TestNanoID(t *testing.T) {
	id := NanoID()
	if len(id) != 21 {
		t.Errorf("NanoID length = %d, want 21", len(id))
	}
}
