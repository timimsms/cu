package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTableFormatter_Format(t *testing.T) {
	t.Run("formats slice of maps", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		data := []map[string]string{
			{"id": "1", "name": "John", "email": "john@example.com"},
			{"id": "2", "name": "Jane", "email": "jane@example.com"},
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "email")
		assert.Contains(t, output, "John")
		assert.Contains(t, output, "jane@example.com")
		// Should have separator line
		assert.Contains(t, output, "---")
	})

	t.Run("formats empty slice with ShowEmpty", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{
			Writer:    &buf,
			ShowEmpty: true,
		}

		data := []map[string]string{}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "No items found")
	})

	t.Run("formats empty slice without ShowEmpty", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{
			Writer:    &buf,
			ShowEmpty: false,
		}

		data := []map[string]string{}

		err := formatter.Format(data)
		assert.NoError(t, err)
		assert.Empty(t, buf.String())
	})

	t.Run("formats slice without headers", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{
			Writer:   &buf,
			NoHeader: true,
		}

		data := []map[string]string{
			{"id": "1", "name": "John"},
			{"id": "2", "name": "Jane"},
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		// Should not have headers
		assert.NotContains(t, output, "id\tname")
		// But should have data
		assert.Contains(t, output, "John")
		assert.Contains(t, output, "Jane")
	})

	t.Run("formats map data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		data := map[string]interface{}{
			"id":     123,
			"name":   "Test User",
			"active": true,
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "KEY")
		assert.Contains(t, output, "VALUE")
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "123")
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "Test User")
		assert.Contains(t, output, "active")
		assert.Contains(t, output, "true")
	})

	t.Run("formats map without headers", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{
			Writer:   &buf,
			NoHeader: true,
		}

		data := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		// Should not have headers
		assert.NotContains(t, output, "KEY\tVALUE")
		// But should have data
		assert.Contains(t, output, "key1")
		assert.Contains(t, output, "value1")
	})

	t.Run("formats struct data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		type TestStruct struct {
			ID     int       `json:"id"`
			Name   string    `json:"name"`
			Active bool      `json:"active"`
			Tags   []string  `json:"tags"`
			Time   time.Time `json:"time"`
		}

		data := TestStruct{
			ID:     1,
			Name:   "Test",
			Active: true,
			Tags:   []string{"tag1", "tag2"},
			Time:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "FIELD")
		assert.Contains(t, output, "VALUE")
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "1")
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "Test")
		assert.Contains(t, output, "active")
		assert.Contains(t, output, "true")
		assert.Contains(t, output, "tags")
		assert.Contains(t, output, "[tag1 tag2]")
	})

	t.Run("formats pointer to struct", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		type TestStruct struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		data := &TestStruct{
			ID:   1,
			Name: "Test",
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "1")
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "Test")
	})

	t.Run("formats simple types", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		// String
		err := formatter.Format("simple string")
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "simple string")

		// Number
		buf.Reset()
		err = formatter.Format(42)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "42")

		// Boolean
		buf.Reset()
		err = formatter.Format(true)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "true")
	})

	t.Run("formats slice of structs", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		type Person struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		data := []Person{
			{ID: 1, Name: "Alice", Age: 30},
			{ID: 2, Name: "Bob", Age: 25},
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "age")
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "30")
		assert.Contains(t, output, "Bob")
		assert.Contains(t, output, "25")
	})

	t.Run("formats struct with unexported fields", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		type TestStruct struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			internal string // unexported, should be skipped
		}

		data := TestStruct{
			ID:       1,
			Name:     "Test",
			internal: "hidden",
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "name")
		assert.NotContains(t, output, "internal")
		assert.NotContains(t, output, "hidden")
	})

	t.Run("uses default writer when nil", func(t *testing.T) {
		formatter := &TableFormatter{Writer: nil}

		// Should not panic
		err := formatter.Format("test")
		assert.NoError(t, err)
	})

	t.Run("formats struct with no json tags", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		type TestStruct struct {
			ID   int
			Name string
		}

		data := TestStruct{
			ID:   1,
			Name: "Test",
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "ID")
		assert.Contains(t, output, "1")
		assert.Contains(t, output, "Name")
		assert.Contains(t, output, "Test")
	})

	t.Run("formats struct with json tag options", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		type TestStruct struct {
			ID       int    `json:"id,omitempty"`
			Name     string `json:"name"`
			Internal string `json:"-"` // Should be skipped
			Renamed  string `json:"custom_name"`
		}

		data := TestStruct{
			ID:       1,
			Name:     "Test",
			Internal: "hidden",
			Renamed:  "value",
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "name")
		assert.NotContains(t, output, "Internal")
		assert.NotContains(t, output, "hidden")
		assert.Contains(t, output, "custom_name")
		assert.Contains(t, output, "value")
	})

	t.Run("formats time values", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		type TestStruct struct {
			Created time.Time `json:"created"`
		}

		testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		data := TestStruct{
			Created: testTime,
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "created")
		// Time should be formatted
		assert.Contains(t, output, "2023")
	})

	t.Run("formats nested structs", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		type Address struct {
			Street string `json:"street"`
			City   string `json:"city"`
		}

		type Person struct {
			Name    string  `json:"name"`
			Address Address `json:"address"`
		}

		data := Person{
			Name: "John",
			Address: Address{
				Street: "123 Main St",
				City:   "New York",
			},
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "John")
		assert.Contains(t, output, "address")
		// Nested struct should be formatted as a string representation
		assert.Contains(t, output, "123 Main St")
	})

	t.Run("formats relative time", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		// Test data with recent time
		now := time.Now()
		data := []map[string]interface{}{
			{
				"id":      "1",
				"created": now.Add(-5 * time.Minute),
			},
			{
				"id":      "2",
				"created": now.Add(-2 * time.Hour),
			},
			{
				"id":      "3",
				"created": now.Add(-48 * time.Hour),
			},
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		// Should format times relatively
		assert.Contains(t, output, "ago")
	})
}

func TestTableFormatter_EdgeCases(t *testing.T) {
	t.Run("handles nil values in map", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		data := map[string]interface{}{
			"key1": "value1",
			"key2": nil,
			"key3": "value3",
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "key1")
		assert.Contains(t, output, "value1")
		assert.Contains(t, output, "key2")
		assert.Contains(t, output, "<nil>")
	})

	t.Run("handles empty struct", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		type EmptyStruct struct{}
		data := EmptyStruct{}

		err := formatter.Format(data)
		assert.NoError(t, err)

		// Should have headers but no data rows
		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.LessOrEqual(t, len(lines), 2) // Headers and separator only
	})

	t.Run("handles struct with all unexported fields", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		type PrivateStruct struct {
			internal1 string
			internal2 int
		}

		data := PrivateStruct{
			internal1: "hidden",
			internal2: 42,
		}

		err := formatter.Format(data)
		assert.NoError(t, err)

		// Should not expose private fields
		output := buf.String()
		assert.NotContains(t, output, "hidden")
		assert.NotContains(t, output, "42")
	})
}
