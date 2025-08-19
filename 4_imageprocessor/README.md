# Image Processor

A modern web application for uploading, processing, and managing images with a beautiful frontend interface.

## Features

- **Image Upload**: Drag and drop or click to upload JPEG/PNG images (up to 20MB)
- **Status Tracking**: Real-time status checking for image processing
- **Image Management**: View processed images and delete them
- **Modern UI**: Responsive design with beautiful gradients and animations
- **RESTful API**: Clean API endpoints for programmatic access

## Frontend

The application includes a modern, responsive frontend built with vanilla HTML, CSS, and JavaScript. The frontend provides:

### Upload Section
- Drag and drop interface for easy file uploads
- File type validation (JPEG/PNG only)
- File size validation (20MB limit)
- Progress indication during upload
- Success/error notifications

### Status Check Section
- Check processing status by image ID
- Real-time status updates
- Display of processed image URLs
- Image preview for completed uploads
- Delete functionality for processed images

### Features
- **Responsive Design**: Works on desktop, tablet, and mobile devices
- **Modern UI**: Beautiful gradients, smooth animations, and intuitive interface
- **Error Handling**: Clear error messages and validation feedback
- **Accessibility**: Keyboard navigation and screen reader support

## API Endpoints

### Upload Image
```
POST /upload
Content-Type: multipart/form-data

Parameters:
- file: Image file (JPEG/PNG, max 20MB)

Response:
{
  "id": "uuid-string"
}
```

### Get Image Status
```
GET /image/{id}

Response:
{
  "url": "http://example.com/image.jpg",  // Only for completed images
  "status": "pending|running|completed|failed"
}
```

### Delete Image
```
DELETE /image/{id}

Response:
{
  "id": "uuid-string"
}
```

## Status Types

- **pending**: Image is queued for processing
- **running**: Image is currently being processed
- **completed**: Image processing is complete, URL is available
- **failed**: Image processing failed

## Getting Started

1. **Start the Backend Services**:
   ```bash
   docker-compose up -d
   ```

2. **Access the Frontend**:
   Open your browser and navigate to `http://localhost:8080`

3. **Upload an Image**:
   - Drag and drop an image file or click to browse
   - Wait for the upload to complete
   - Note the image ID returned

4. **Check Status**:
   - Enter the image ID in the status check section
   - Click "Check Status" to see the current processing status

5. **View Results**:
   - Once processing is complete, the image URL will be displayed
   - Click the URL to view the processed image
   - Use the delete button to remove the image

## Development

### Running Tests
```bash
go test ./...
```

### Building the Application
```bash
go build ./cmd/processor
go build ./cmd/editor
```

### Code Structure
```
4_imageprocessor/
├── cmd/                    # Application entry points
│   ├── processor/         # Main API server
│   └── editor/           # Image processing worker
├── internal/              # Internal application code
│   ├── api/              # HTTP API handlers
│   ├── domain/           # Domain models and interfaces
│   ├── usecase/          # Business logic
│   ├── storage/          # Data storage interfaces
│   ├── broker/           # Message broker interfaces
│   └── editor/           # Image processing logic
├── static/               # Static assets
├── migrations/           # Database migrations
├── index.html           # Frontend application
├── config.yaml          # Configuration file
└── docker-compose.yml   # Docker services
```

## Configuration

The application is configured via `config.yaml`:

```yaml
s3:
  endpoint: localhost:9000
  bucket: images
  base_url: http://localhost:9000
  ssl: false

kafka:
  brokers: localhost:29092

postgres:
  masterdsn: postgres://postgres:${POSTGRES_PASSWORD}@localhost:5432/imageprocessor?sslmode=disable

server:
  addr: 0.0.0.0:8080

editor:
  watermark: ./static/watermark.png
```

## Technologies Used

- **Backend**: Go, Gin, PostgreSQL, MinIO, Kafka
- **Frontend**: HTML5, CSS3, JavaScript (ES6+)
- **Testing**: Go testing framework
- **Containerization**: Docker, Docker Compose

## Browser Support

The frontend supports all modern browsers:
- Chrome 60+
- Firefox 55+
- Safari 12+
- Edge 79+

## License

This project is licensed under the MIT License.