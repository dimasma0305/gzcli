package repository

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// GZAPIAttachmentRepository implements AttachmentRepository using GZAPI
type GZAPIAttachmentRepository struct{}

// NewGZAPIAttachmentRepository creates a new GZAPI attachment repository
func NewGZAPIAttachmentRepository() *GZAPIAttachmentRepository {
	return &GZAPIAttachmentRepository{}
}

// Create creates a new attachment
func (r *GZAPIAttachmentRepository) Create(challengeID int, attachment gzapi.CreateAttachmentForm) error {
	// This would need to be implemented based on the GZAPI create attachment functionality
	// For now, return an error as this might not be directly supported
	return fmt.Errorf("create attachment not implemented")
}

// Delete deletes an attachment
func (r *GZAPIAttachmentRepository) Delete(challengeID int, attachmentID int) error {
	// This would need to be implemented based on the GZAPI delete attachment functionality
	return fmt.Errorf("delete attachment not implemented")
}

// List returns all attachments for a challenge
func (r *GZAPIAttachmentRepository) List(challengeID int) ([]gzapi.Attachment, error) {
	// This would need to be implemented based on the GZAPI list attachments functionality
	return nil, fmt.Errorf("list attachments not implemented")
}