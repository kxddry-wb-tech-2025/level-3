package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"
)

func TestTask_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		task     Task
		expected string
	}{
		{
			name: "pending task",
			task: Task{
				FileName:  "test.jpg",
				Status:    StatusPending,
				CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: `{"file_name":"test.jpg","status":"pending","created_at":"2023-01-01T12:00:00Z"}`,
		},
		{
			name: "completed task",
			task: Task{
				FileName:  "test.png",
				Status:    StatusCompleted,
				CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: `{"file_name":"test.png","status":"completed","created_at":"2023-01-01T12:00:00Z"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.task)
			if err != nil {
				t.Fatalf("Failed to marshal task: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Expected JSON: %s, got: %s", tt.expected, string(data))
			}

			// Test JSON unmarshaling
			var unmarshaledTask Task
			err = json.Unmarshal(data, &unmarshaledTask)
			if err != nil {
				t.Fatalf("Failed to unmarshal task: %v", err)
			}

			if unmarshaledTask.FileName != tt.task.FileName {
				t.Errorf("Expected FileName: %s, got: %s", tt.task.FileName, unmarshaledTask.FileName)
			}
			if unmarshaledTask.Status != tt.task.Status {
				t.Errorf("Expected Status: %s, got: %s", tt.task.Status, unmarshaledTask.Status)
			}
		})
	}
}

func TestImage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		image    Image
		expected string
	}{
		{
			name: "completed image with URL",
			image: Image{
				URL:    "http://example.com/image.jpg",
				Status: StatusCompleted,
			},
			expected: `{"url":"http://example.com/image.jpg","status":"completed"}`,
		},
		{
			name: "pending image without URL",
			image: Image{
				Status: StatusPending,
			},
			expected: `{"status":"pending"}`,
		},
		{
			name: "failed image without URL",
			image: Image{
				Status: StatusFailed,
			},
			expected: `{"status":"failed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.image)
			if err != nil {
				t.Fatalf("Failed to marshal image: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Expected JSON: %s, got: %s", tt.expected, string(data))
			}

			// Test JSON unmarshaling
			var unmarshaledImage Image
			err = json.Unmarshal(data, &unmarshaledImage)
			if err != nil {
				t.Fatalf("Failed to unmarshal image: %v", err)
			}

			if unmarshaledImage.URL != tt.image.URL {
				t.Errorf("Expected URL: %s, got: %s", tt.image.URL, unmarshaledImage.URL)
			}
			if unmarshaledImage.Status != tt.image.Status {
				t.Errorf("Expected Status: %s, got: %s", tt.image.Status, unmarshaledImage.Status)
			}
		})
	}
}

func TestFile_ReadSeekClose(t *testing.T) {
	content := []byte("test file content")
	reader := bytes.NewReader(content)
	
	// Create a ReadSeekCloser from bytes.Reader
	readSeekCloser := struct {
		io.Reader
		io.Seeker
		io.Closer
	}{
		Reader: reader,
		Seeker: reader,
		Closer: io.NopCloser(reader),
	}
	
	file := &File{
		Name:        "test.txt",
		Data:        readSeekCloser,
		Size:        int64(len(content)),
		ContentType: "text/plain",
	}

	// Test reading
	buf := make([]byte, len(content))
	n, err := file.Data.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Failed to read file: %v", err)
	}

	if n != len(content) {
		t.Errorf("Expected to read %d bytes, got %d", len(content), n)
	}

	if string(buf) != string(content) {
		t.Errorf("Expected content: %s, got: %s", string(content), string(buf))
	}

	// Test seeking
	seeker, ok := file.Data.(io.Seeker)
	if !ok {
		t.Fatal("File.Data does not implement io.Seeker")
	}

	// Seek to beginning
	pos, err := seeker.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("Failed to seek to start: %v", err)
	}
	if pos != 0 {
		t.Errorf("Expected position 0, got %d", pos)
	}

	// Test closing
	err = file.Data.Close()
	if err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}
}

func TestStatusConstants(t *testing.T) {
	expectedStatuses := []string{
		StatusPending,
		StatusRunning,
		StatusCompleted,
		StatusFailed,
	}

	for _, status := range expectedStatuses {
		if status == "" {
			t.Errorf("Status constant is empty")
		}
	}

	// Test that all statuses are unique
	statusMap := make(map[string]bool)
	for _, status := range expectedStatuses {
		if statusMap[status] {
			t.Errorf("Duplicate status found: %s", status)
		}
		statusMap[status] = true
	}
}

func TestKafkaMessage_JSONSerialization(t *testing.T) {
	task := Task{
		FileName:  "test.jpg",
		Status:    StatusPending,
		CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	message := KafkaMessage{
		Task: task,
		Commit: func() error {
			return nil
		},
	}

	// Test that Task can be marshaled
	data, err := json.Marshal(message.Task)
	if err != nil {
		t.Fatalf("Failed to marshal task in KafkaMessage: %v", err)
	}

	var unmarshaledTask Task
	err = json.Unmarshal(data, &unmarshaledTask)
	if err != nil {
		t.Fatalf("Failed to unmarshal task in KafkaMessage: %v", err)
	}

	if unmarshaledTask.FileName != task.FileName {
		t.Errorf("Expected FileName: %s, got: %s", task.FileName, unmarshaledTask.FileName)
	}
	if unmarshaledTask.Status != task.Status {
		t.Errorf("Expected Status: %s, got: %s", task.Status, unmarshaledTask.Status)
	}

	// Test commit function
	err = message.Commit()
	if err != nil {
		t.Errorf("Expected commit to succeed, got error: %v", err)
	}
}