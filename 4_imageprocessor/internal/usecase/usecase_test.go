package usecase

import (
	"context"
	"errors"
	"image-processor/internal/domain"
	"testing"
	"time"
)

// Mock implementations for testing
type mockFileStorage struct {
	uploadErr error
	getURLErr error
	getErr    error
	deleteErr error
	url       string
	file      *domain.File
}

func (m *mockFileStorage) Upload(ctx context.Context, file *domain.File) error {
	return m.uploadErr
}

func (m *mockFileStorage) GetURL(ctx context.Context, fileName string) (string, error) {
	return m.url, m.getURLErr
}

func (m *mockFileStorage) Get(ctx context.Context, fileName string) (*domain.File, error) {
	return m.file, m.getErr
}

func (m *mockFileStorage) Delete(ctx context.Context, fileName string) error {
	return m.deleteErr
}

type mockStatusStorage struct {
	addFileErr     error
	updateStatusErr error
	getStatusErr   error
	getFileNameErr error
	deleteFileErr  error
	status         string
	fileName       string
}

func (m *mockStatusStorage) AddFile(ctx context.Context, file *domain.File) error {
	return m.addFileErr
}

func (m *mockStatusStorage) UpdateStatus(ctx context.Context, fileName string, status string) error {
	return m.updateStatusErr
}

func (m *mockStatusStorage) GetStatus(ctx context.Context, id string) (string, error) {
	return m.status, m.getStatusErr
}

func (m *mockStatusStorage) GetFileName(ctx context.Context, id string) (string, error) {
	return m.fileName, m.getFileNameErr
}

func (m *mockStatusStorage) DeleteFile(ctx context.Context, id string) error {
	return m.deleteFileErr
}

type mockTaskSender struct {
	sendTaskErr error
}

func (m *mockTaskSender) SendTask(ctx context.Context, task *domain.Task) error {
	return m.sendTaskErr
}

func TestHandler_UploadImage_Success(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	file := &domain.File{
		Name:        "test.jpg",
		Data:        nil,
		Size:        1024,
		ContentType: "image/jpeg",
	}
	
	err := handler.UploadImage(context.Background(), file)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestHandler_UploadImage_FileStorageError(t *testing.T) {
	fs := &mockFileStorage{uploadErr: errors.New("storage error")}
	ss := &mockStatusStorage{}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	file := &domain.File{
		Name:        "test.jpg",
		Data:        nil,
		Size:        1024,
		ContentType: "image/jpeg",
	}
	
	err := handler.UploadImage(context.Background(), file)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "usecase.UploadImage: storage error" {
		t.Errorf("Expected error message 'usecase.UploadImage: storage error', got: %s", err.Error())
	}
}

func TestHandler_UploadImage_StatusStorageError(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{addFileErr: errors.New("status storage error")}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	file := &domain.File{
		Name:        "test.jpg",
		Data:        nil,
		Size:        1024,
		ContentType: "image/jpeg",
	}
	
	err := handler.UploadImage(context.Background(), file)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "usecase.UploadImage: status storage error" {
		t.Errorf("Expected error message 'usecase.UploadImage: status storage error', got: %s", err.Error())
	}
}

func TestHandler_UploadImage_TaskSenderError(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{}
	ts := &mockTaskSender{sendTaskErr: errors.New("task sender error")}
	
	handler := New(fs, ss, ts)
	
	file := &domain.File{
		Name:        "test.jpg",
		Data:        nil,
		Size:        1024,
		ContentType: "image/jpeg",
	}
	
	err := handler.UploadImage(context.Background(), file)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "usecase.UploadImage: task sender error" {
		t.Errorf("Expected error message 'usecase.UploadImage: task sender error', got: %s", err.Error())
	}
}

func TestHandler_GetImage_Completed(t *testing.T) {
	fs := &mockFileStorage{url: "http://example.com/image.jpg"}
	ss := &mockStatusStorage{
		status:   domain.StatusCompleted,
		fileName: "test.jpg",
	}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	image, err := handler.GetImage(context.Background(), "test-id")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if image.Status != domain.StatusCompleted {
		t.Errorf("Expected status %s, got: %s", domain.StatusCompleted, image.Status)
	}
	
	if image.URL != "http://example.com/image.jpg" {
		t.Errorf("Expected URL %s, got: %s", "http://example.com/image.jpg", image.URL)
	}
}

func TestHandler_GetImage_Pending(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{status: domain.StatusPending}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	image, err := handler.GetImage(context.Background(), "test-id")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if image.Status != domain.StatusPending {
		t.Errorf("Expected status %s, got: %s", domain.StatusPending, image.Status)
	}
	
	if image.URL != "" {
		t.Errorf("Expected empty URL, got: %s", image.URL)
	}
}

func TestHandler_GetImage_Running(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{status: domain.StatusRunning}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	image, err := handler.GetImage(context.Background(), "test-id")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if image.Status != domain.StatusRunning {
		t.Errorf("Expected status %s, got: %s", domain.StatusRunning, image.Status)
	}
	
	if image.URL != "" {
		t.Errorf("Expected empty URL, got: %s", image.URL)
	}
}

func TestHandler_GetImage_Failed(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{status: domain.StatusFailed}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	image, err := handler.GetImage(context.Background(), "test-id")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if image.Status != domain.StatusFailed {
		t.Errorf("Expected status %s, got: %s", domain.StatusFailed, image.Status)
	}
	
	if image.URL != "" {
		t.Errorf("Expected empty URL, got: %s", image.URL)
	}
}

func TestHandler_GetImage_StatusStorageError(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{getStatusErr: errors.New("status storage error")}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	_, err := handler.GetImage(context.Background(), "test-id")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "usecase.GetImage: status storage error" {
		t.Errorf("Expected error message 'usecase.GetImage: status storage error', got: %s", err.Error())
	}
}

func TestHandler_GetImage_GetFileNameError(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{
		status:         domain.StatusCompleted,
		getFileNameErr: errors.New("get filename error"),
	}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	_, err := handler.GetImage(context.Background(), "test-id")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "usecase.GetImage: get filename error" {
		t.Errorf("Expected error message 'usecase.GetImage: get filename error', got: %s", err.Error())
	}
}

func TestHandler_GetImage_GetURLError(t *testing.T) {
	fs := &mockFileStorage{getURLErr: errors.New("get URL error")}
	ss := &mockStatusStorage{
		status:   domain.StatusCompleted,
		fileName: "test.jpg",
	}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	_, err := handler.GetImage(context.Background(), "test-id")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "usecase.GetImage: get URL error" {
		t.Errorf("Expected error message 'usecase.GetImage: get URL error', got: %s", err.Error())
	}
}

func TestHandler_DeleteImage_Success(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	err := handler.DeleteImage(context.Background(), "test-id")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestHandler_DeleteImage_StatusStorageError(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{deleteFileErr: errors.New("status storage error")}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	err := handler.DeleteImage(context.Background(), "test-id")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "usecase.DeleteImage: status storage error" {
		t.Errorf("Expected error message 'usecase.DeleteImage: status storage error', got: %s", err.Error())
	}
}

func TestHandler_DeleteImage_FileStorageError(t *testing.T) {
	fs := &mockFileStorage{deleteErr: errors.New("file storage error")}
	ss := &mockStatusStorage{}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	err := handler.DeleteImage(context.Background(), "test-id")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "usecase.DeleteImage: file storage error" {
		t.Errorf("Expected error message 'usecase.DeleteImage: file storage error', got: %s", err.Error())
	}
}

func TestHandler_New(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{}
	ts := &mockTaskSender{}
	
	handler := New(fs, ss, ts)
	
	if handler == nil {
		t.Error("Expected handler to be created, got nil")
	}
	
	if handler.fs != fs {
		t.Error("Expected file storage to be set correctly")
	}
	
	if handler.ss != ss {
		t.Error("Expected status storage to be set correctly")
	}
	
	if handler.ts != ts {
		t.Error("Expected task sender to be set correctly")
	}
}

func TestHandler_UploadImage_CreatesCorrectTask(t *testing.T) {
	fs := &mockFileStorage{}
	ss := &mockStatusStorage{}
	
	// Create a custom task sender that captures the task
	var capturedTask *domain.Task
	ts := &mockTaskSenderWithCapture{capturedTask: &capturedTask}
	
	handler := New(fs, ss, ts)
	
	file := &domain.File{
		Name:        "test.jpg",
		Data:        nil,
		Size:        1024,
		ContentType: "image/jpeg",
	}
	
	err := handler.UploadImage(context.Background(), file)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if capturedTask == nil {
		t.Error("Expected task to be captured, got nil")
	}
	
	if capturedTask.FileName != "test.jpg" {
		t.Errorf("Expected filename %s, got: %s", "test.jpg", capturedTask.FileName)
	}
	
	if capturedTask.Status != domain.StatusPending {
		t.Errorf("Expected status %s, got: %s", domain.StatusPending, capturedTask.Status)
	}
	
	// Check that CreatedAt is set to a recent time
	if capturedTask.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set, got zero time")
	}
	
	// Check that CreatedAt is within the last second
	if time.Since(capturedTask.CreatedAt) > time.Second {
		t.Error("Expected CreatedAt to be recent")
	}
}

type mockTaskSenderWithCapture struct {
	capturedTask **domain.Task
}

func (m *mockTaskSenderWithCapture) SendTask(ctx context.Context, task *domain.Task) error {
	*m.capturedTask = task
	return nil
}