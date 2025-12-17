package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	ISOUploadDir = "./uploads/isos"
)

// StorageService handles file storage operations
type StorageService struct {
	uploadDir string
}

// NewStorageService creates a new instance of StorageService
func NewStorageService() (*StorageService, error) {
	service := &StorageService{
		uploadDir: ISOUploadDir,
	}
	
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(service.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	
	return service, nil
}

// SaveISO saves an ISO file from a reader to the storage directory
// Uses io.Copy for streaming to avoid loading the entire file into RAM
func (s *StorageService) SaveISO(filename string) (string, error) {
	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(filename), ".iso") {
		return "", fmt.Errorf("invalid file extension. Only .iso files are allowed")
	}
	
	// Generate safe filename
	safeFilename := filepath.Base(filename)
	
	// Create full file path
	filePath := filepath.Join(s.uploadDir, safeFilename)
	
	return filePath, nil
}

// SaveISOFromReader saves an ISO file from an io.Reader to the storage directory
// Uses io.Copy for streaming to avoid loading the entire file into RAM
func (s *StorageService) SaveISOFromReader(filename string, reader io.Reader) (string, error) {
	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(filename), ".iso") {
		return "", fmt.Errorf("invalid file extension. Only .iso files are allowed")
	}
	
	// Generate safe filename
	safeFilename := filepath.Base(filename)
	
	// Create full file path
	filePath := filepath.Join(s.uploadDir, safeFilename)
	
	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	// Stream copy from reader to file (this avoids loading entire file into RAM)
	_, err = io.Copy(file, reader)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	
	return filePath, nil
}

// GetISOPath returns the full path for an ISO file
func (s *StorageService) GetISOPath(filename string) string {
	return filepath.Join(s.uploadDir, filepath.Base(filename))
}

// ListISOs returns a list of ISO files in the upload directory
func (s *StorageService) ListISOs() ([]string, error) {
	var isos []string
	
	entries, err := os.ReadDir(s.uploadDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read upload directory: %w", err)
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".iso") {
			isos = append(isos, entry.Name())
		}
	}
	
	return isos, nil
}

// DeleteISO removes an ISO file from the storage
func (s *StorageService) DeleteISO(filename string) error {
	filePath := s.GetISOPath(filename)
	return os.Remove(filePath)
}