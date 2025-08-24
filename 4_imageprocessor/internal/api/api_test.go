package api

import (
	"bytes"
	"context"
	"errors"
	"image-processor/internal/domain"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Mock handler for testing
type mockHandler struct {
	uploadErr error
	getErr    error
	deleteErr error
	image     *domain.Image
}

func (m *mockHandler) UploadImage(ctx context.Context, file *domain.File) error {
	return m.uploadErr
}

func (m *mockHandler) GetImage(ctx context.Context, id string) (*domain.Image, error) {
	return m.image, m.getErr
}

func (m *mockHandler) DeleteImage(ctx context.Context, id string) error {
	return m.deleteErr
}

func TestServer_New(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

	if server == nil {
		t.Error("Expected server to be created, got nil")
		return
	}

	if server.h != handler {
		t.Error("Expected handler to be set correctly")
	}

=======
	
	if server == nil {
		t.Error("Expected server to be created, got nil")
	}
	
	if server.h != handler {
		t.Error("Expected handler to be set correctly")
	}
	
>>>>>>> dev
	if server.r == nil {
		t.Error("Expected router to be created")
	}
}

func TestServer_UploadImage_Success(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Create a valid JPEG file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Write JPEG header to make it a valid image
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	part.Write(jpegHeader)
	writer.Close()
<<<<<<< HEAD

	// Create request
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create response recorder
	w := httptest.NewRecorder()

=======
	
	// Create request
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	// Create response recorder
	w := httptest.NewRecorder()
	
>>>>>>> dev
	// Create gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
<<<<<<< HEAD

	// Call the handler
	handlerFunc := server.uploadImage()
	handlerFunc(c)

=======
	
	// Call the handler
	handlerFunc := server.uploadImage()
	handlerFunc(c)
	
>>>>>>> dev
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Check that response contains an ID
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "id") {
		t.Error("Expected response to contain 'id' field")
	}
}

func TestServer_UploadImage_NoFile(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

	// Create request without file
	req := httptest.NewRequest("POST", "/upload", strings.NewReader(""))
	req.Header.Set("Content-Type", "multipart/form-data")

=======
	
	// Create request without file
	req := httptest.NewRequest("POST", "/upload", strings.NewReader(""))
	req.Header.Set("Content-Type", "multipart/form-data")
	
>>>>>>> dev
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
<<<<<<< HEAD

	handlerFunc := server.uploadImage()
	handlerFunc(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

=======
	
	handlerFunc := server.uploadImage()
	handlerFunc(c)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	
>>>>>>> dev
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "error") {
		t.Error("Expected response to contain error message")
	}
}

func TestServer_UploadImage_FileTooLarge(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Create a large file (over 20MB)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "large.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Write content larger than 20MB
	largeContent := make([]byte, 21*1024*1024) // 21MB
	part.Write(largeContent)
	writer.Close()
<<<<<<< HEAD

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

=======
	
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
>>>>>>> dev
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
<<<<<<< HEAD

	handlerFunc := server.uploadImage()
	handlerFunc(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

=======
	
	handlerFunc := server.uploadImage()
	handlerFunc(c)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	
>>>>>>> dev
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "file size is too large") {
		t.Error("Expected response to contain file size error message")
	}
}

func TestServer_UploadImage_InvalidFileType(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Create a file with invalid content type
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Write text content (not an image)
	content := []byte("this is not an image")
	part.Write(content)
	writer.Close()
<<<<<<< HEAD

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

=======
	
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
>>>>>>> dev
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
<<<<<<< HEAD

	handlerFunc := server.uploadImage()
	handlerFunc(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

=======
	
	handlerFunc := server.uploadImage()
	handlerFunc(c)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	
>>>>>>> dev
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "invalid file type") {
		t.Error("Expected response to contain invalid file type error message")
	}
}

func TestServer_UploadImage_HandlerError(t *testing.T) {
	handler := &mockHandler{uploadErr: errors.New("upload failed")}
	server := New(handler)
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Create a valid image file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Write JPEG header to make it a valid image
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	part.Write(jpegHeader)
	writer.Close()
<<<<<<< HEAD

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

=======
	
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
>>>>>>> dev
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
<<<<<<< HEAD

	handlerFunc := server.uploadImage()
	handlerFunc(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

=======
	
	handlerFunc := server.uploadImage()
	handlerFunc(c)
	
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
	
>>>>>>> dev
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "upload failed") {
		t.Error("Expected response to contain handler error message")
	}
}

func TestServer_GetImage_Success(t *testing.T) {
	expectedImage := &domain.Image{
		URL:    "http://example.com/image.jpg",
		Status: domain.StatusCompleted,
	}
<<<<<<< HEAD

	handler := &mockHandler{image: expectedImage}
	server := New(handler)

	// Generate a valid UUID
	id := uuid.New().String()

	req := httptest.NewRequest("GET", "/image/"+id, nil)
	w := httptest.NewRecorder()

=======
	
	handler := &mockHandler{image: expectedImage}
	server := New(handler)
	
	// Generate a valid UUID
	id := uuid.New().String()
	
	req := httptest.NewRequest("GET", "/image/"+id, nil)
	w := httptest.NewRecorder()
	
>>>>>>> dev
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id}}
<<<<<<< HEAD

	handlerFunc := server.getImage()
	handlerFunc(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

=======
	
	handlerFunc := server.getImage()
	handlerFunc(c)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
>>>>>>> dev
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "http://example.com/image.jpg") {
		t.Error("Expected response to contain image URL")
	}
	if !strings.Contains(responseBody, "completed") {
		t.Error("Expected response to contain completed status")
	}
}

func TestServer_GetImage_InvalidID(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

	// Use an invalid UUID
	invalidID := "invalid-uuid"

	req := httptest.NewRequest("GET", "/image/"+invalidID, nil)
	w := httptest.NewRecorder()

=======
	
	// Use an invalid UUID
	invalidID := "invalid-uuid"
	
	req := httptest.NewRequest("GET", "/image/"+invalidID, nil)
	w := httptest.NewRecorder()
	
>>>>>>> dev
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: invalidID}}
<<<<<<< HEAD

	handlerFunc := server.getImage()
	handlerFunc(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

=======
	
	handlerFunc := server.getImage()
	handlerFunc(c)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	
>>>>>>> dev
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "invalid id") {
		t.Error("Expected response to contain invalid id error message")
	}
}

func TestServer_GetImage_HandlerError(t *testing.T) {
	handler := &mockHandler{getErr: errors.New("get failed")}
	server := New(handler)
<<<<<<< HEAD

	id := uuid.New().String()

	req := httptest.NewRequest("GET", "/image/"+id, nil)
	w := httptest.NewRecorder()

=======
	
	id := uuid.New().String()
	
	req := httptest.NewRequest("GET", "/image/"+id, nil)
	w := httptest.NewRecorder()
	
>>>>>>> dev
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id}}
<<<<<<< HEAD

	handlerFunc := server.getImage()
	handlerFunc(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

=======
	
	handlerFunc := server.getImage()
	handlerFunc(c)
	
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
	
>>>>>>> dev
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "get failed") {
		t.Error("Expected response to contain handler error message")
	}
}

func TestServer_DeleteImage_Success(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

	id := uuid.New().String()

	req := httptest.NewRequest("DELETE", "/image/"+id, nil)
	w := httptest.NewRecorder()

=======
	
	id := uuid.New().String()
	
	req := httptest.NewRequest("DELETE", "/image/"+id, nil)
	w := httptest.NewRecorder()
	
>>>>>>> dev
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id}}
<<<<<<< HEAD

	handlerFunc := server.deleteImage()
	handlerFunc(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

=======
	
	handlerFunc := server.deleteImage()
	handlerFunc(c)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
>>>>>>> dev
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, id) {
		t.Error("Expected response to contain deleted image ID")
	}
}

func TestServer_DeleteImage_InvalidID(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

	invalidID := "invalid-uuid"

	req := httptest.NewRequest("DELETE", "/image/"+invalidID, nil)
	w := httptest.NewRecorder()

=======
	
	invalidID := "invalid-uuid"
	
	req := httptest.NewRequest("DELETE", "/image/"+invalidID, nil)
	w := httptest.NewRecorder()
	
>>>>>>> dev
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: invalidID}}
<<<<<<< HEAD

	handlerFunc := server.deleteImage()
	handlerFunc(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

=======
	
	handlerFunc := server.deleteImage()
	handlerFunc(c)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	
>>>>>>> dev
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "invalid id") {
		t.Error("Expected response to contain invalid id error message")
	}
}

func TestServer_DeleteImage_HandlerError(t *testing.T) {
	handler := &mockHandler{deleteErr: errors.New("delete failed")}
	server := New(handler)
<<<<<<< HEAD

	id := uuid.New().String()

	req := httptest.NewRequest("DELETE", "/image/"+id, nil)
	w := httptest.NewRecorder()

=======
	
	id := uuid.New().String()
	
	req := httptest.NewRequest("DELETE", "/image/"+id, nil)
	w := httptest.NewRecorder()
	
>>>>>>> dev
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id}}
<<<<<<< HEAD

	handlerFunc := server.deleteImage()
	handlerFunc(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

=======
	
	handlerFunc := server.deleteImage()
	handlerFunc(c)
	
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
	
>>>>>>> dev
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "delete failed") {
		t.Error("Expected response to contain handler error message")
	}
}

func TestServer_RegisterRoutes(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Test that routes are registered by checking if the server can be created
	if server.r == nil {
		t.Error("Expected router to be initialized")
	}
}

// Helper function to create a valid image file for testing
func createValidImageFile(t *testing.T, filename string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Write JPEG header to make it a valid image
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	part.Write(jpegHeader)
	writer.Close()
<<<<<<< HEAD

=======
	
>>>>>>> dev
	return body, writer.FormDataContentType()
}

func TestServer_UploadImage_ValidJPEG(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

	body, contentType := createValidImageFile(t, "test.jpg")

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", contentType)

=======
	
	body, contentType := createValidImageFile(t, "test.jpg")
	
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", contentType)
	
>>>>>>> dev
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
<<<<<<< HEAD

	handlerFunc := server.uploadImage()
	handlerFunc(c)

=======
	
	handlerFunc := server.uploadImage()
	handlerFunc(c)
	
>>>>>>> dev
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestServer_UploadImage_ValidPNG(t *testing.T) {
	handler := &mockHandler{}
	server := New(handler)
<<<<<<< HEAD

=======
	
>>>>>>> dev
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.png")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
<<<<<<< HEAD

=======
	
>>>>>>> dev
	// Write PNG header to make it a valid image
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	part.Write(pngHeader)
	writer.Close()
<<<<<<< HEAD

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

=======
	
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
>>>>>>> dev
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
<<<<<<< HEAD

	handlerFunc := server.uploadImage()
	handlerFunc(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
=======
	
	handlerFunc := server.uploadImage()
	handlerFunc(c)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
>>>>>>> dev
