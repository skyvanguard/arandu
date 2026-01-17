package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteScreenshotToFile(t *testing.T) {
	// Create test data
	testData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header

	// Clean up before and after test
	defer os.RemoveAll("./tmp")
	os.RemoveAll("./tmp")

	filename, err := writeScreenshotToFile(testData)
	if err != nil {
		t.Fatalf("writeScreenshotToFile() error = %v", err)
	}

	// Verify filename format (should end with .png)
	if !strings.HasSuffix(filename, ".png") {
		t.Errorf("filename = %q, want suffix .png", filename)
	}

	// Verify file was created
	filepath := filepath.Join("./tmp/browser", filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Errorf("file was not created at %s", filepath)
	}

	// Verify file content
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if len(content) != len(testData) {
		t.Errorf("file content length = %d, want %d", len(content), len(testData))
	}

	for i, b := range content {
		if b != testData[i] {
			t.Errorf("file content[%d] = %v, want %v", i, b, testData[i])
		}
	}
}

func TestWriteScreenshotToFile_EmptyData(t *testing.T) {
	// Clean up before and after test
	defer os.RemoveAll("./tmp")
	os.RemoveAll("./tmp")

	filename, err := writeScreenshotToFile([]byte{})
	if err != nil {
		t.Fatalf("writeScreenshotToFile() error = %v", err)
	}

	// File should be created even with empty data
	if filename == "" {
		t.Error("filename should not be empty")
	}
}

func TestWriteScreenshotToFile_LargeData(t *testing.T) {
	// Clean up before and after test
	defer os.RemoveAll("./tmp")
	os.RemoveAll("./tmp")

	// Create 1MB of test data
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	filename, err := writeScreenshotToFile(largeData)
	if err != nil {
		t.Fatalf("writeScreenshotToFile() error = %v", err)
	}

	// Verify file was created with correct size
	filepath := filepath.Join("./tmp/browser", filename)
	info, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if info.Size() != int64(len(largeData)) {
		t.Errorf("file size = %d, want %d", info.Size(), len(largeData))
	}
}

func TestWriteScreenshotToFile_DirectoryCreation(t *testing.T) {
	// Clean up before and after test
	defer os.RemoveAll("./tmp")
	os.RemoveAll("./tmp")

	// Ensure directory doesn't exist
	if _, err := os.Stat("./tmp/browser"); !os.IsNotExist(err) {
		t.Fatal("directory should not exist before test")
	}

	_, err := writeScreenshotToFile([]byte{1, 2, 3})
	if err != nil {
		t.Fatalf("writeScreenshotToFile() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat("./tmp/browser"); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}
