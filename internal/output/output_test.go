package output

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableFormatter(t *testing.T) {
	t.Run("TableFormatter formats data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}

		data := []map[string]string{
			{"id": "123", "name": "test"},
		}

		err := formatter.Format(data)
		assert.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})
}

func TestJSONFormatter(t *testing.T) {
	t.Run("JSONFormatter formats data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &JSONFormatter{Writer: &buf}
		data := map[string]string{"id": "123", "name": "test"}

		err := formatter.Format(data)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "123")
		assert.Contains(t, buf.String(), "test")
	})
}

func TestYAMLFormatter(t *testing.T) {
	t.Run("YAMLFormatter formats data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &YAMLFormatter{Writer: &buf}
		data := map[string]string{"id": "123", "name": "test"}

		err := formatter.Format(data)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "123")
		assert.Contains(t, buf.String(), "test")
	})
}

func TestFormat(t *testing.T) {
	testData := map[string]string{"key": "value"}

	t.Run("formats as JSON", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Format("json", testData)

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "key")
		assert.Contains(t, buf.String(), "value")
	})

	t.Run("formats as YAML", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Format("yaml", testData)

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "key")
		assert.Contains(t, buf.String(), "value")
	})

	t.Run("formats as YML (alias for YAML)", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Format("yml", testData)

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "key")
		assert.Contains(t, buf.String(), "value")
	})

	t.Run("formats as CSV", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Use slice data for CSV
		csvData := []map[string]interface{}{
			{"id": "1", "name": "test"},
			{"id": "2", "name": "test2"},
		}

		err := Format("csv", csvData)

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "test")
	})

	t.Run("formats as Table", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Format("table", testData)

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		assert.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})

	t.Run("returns error for unsupported format", func(t *testing.T) {
		err := Format("unsupported", testData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported output format")
	})

	t.Run("handles case-insensitive format names", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Format("JSON", testData)

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "key")
	})
}

func TestCSVFormatter_Format(t *testing.T) {
	t.Run("formats [][]string data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &CSVFormatter{Writer: &buf}
		
		data := [][]string{
			{"id", "name"},
			{"1", "test"},
			{"2", "test2"},
		}

		err := formatter.Format(data)
		assert.NoError(t, err)
		
		output := buf.String()
		assert.Contains(t, output, "id,name")
		assert.Contains(t, output, "1,test")
		assert.Contains(t, output, "2,test2")
	})

	t.Run("formats []map[string]interface{} data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &CSVFormatter{Writer: &buf}
		
		data := []map[string]interface{}{
			{"id": 1, "name": "test", "active": true},
			{"id": 2, "name": "test2", "active": false},
		}

		err := formatter.Format(data)
		assert.NoError(t, err)
		
		output := buf.String()
		// Headers should be present
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "active")
		// Values should be present
		assert.Contains(t, output, "1")
		assert.Contains(t, output, "test")
		assert.Contains(t, output, "true")
	})

	t.Run("handles empty []map[string]interface{}", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &CSVFormatter{Writer: &buf}
		
		data := []map[string]interface{}{}

		err := formatter.Format(data)
		assert.NoError(t, err)
		assert.Empty(t, buf.String())
	})

	t.Run("formats []map[string]string data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &CSVFormatter{Writer: &buf}
		
		data := []map[string]string{
			{"id": "1", "name": "test"},
			{"id": "2", "name": "test2"},
		}

		err := formatter.Format(data)
		assert.NoError(t, err)
		
		output := buf.String()
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "1")
		assert.Contains(t, output, "test")
	})

	t.Run("handles empty []map[string]string", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &CSVFormatter{Writer: &buf}
		
		data := []map[string]string{}

		err := formatter.Format(data)
		assert.NoError(t, err)
		assert.Empty(t, buf.String())
	})

	t.Run("formats struct data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &CSVFormatter{Writer: &buf}
		
		type TestStruct struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		
		// Single struct should be converted to slice
		data := TestStruct{ID: 1, Name: "test"}

		err := formatter.Format(data)
		assert.NoError(t, err)
		
		output := buf.String()
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "1")
		assert.Contains(t, output, "test")
	})

	t.Run("formats slice of structs", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &CSVFormatter{Writer: &buf}
		
		type TestStruct struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		
		data := []TestStruct{
			{ID: 1, Name: "test1"},
			{ID: 2, Name: "test2"},
		}

		err := formatter.Format(data)
		assert.NoError(t, err)
		
		output := buf.String()
		assert.Contains(t, output, "id")
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "1")
		assert.Contains(t, output, "test1")
		assert.Contains(t, output, "2")
		assert.Contains(t, output, "test2")
	})

	t.Run("handles unsupported data type", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &CSVFormatter{Writer: &buf}
		
		// Unsupported type
		data := 123

		err := formatter.Format(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported CSV data type")
	})
}
