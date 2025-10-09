package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
	"github.com/dimasma0305/gzcli/internal/log"
)

// GZAPIAttachmentRepository implements AttachmentRepository using GZAPI
type GZAPIAttachmentRepository struct {
	api *gzapi.GZAPI
}

// NewGZAPIAttachmentRepository creates a new GZAPI attachment repository
func NewGZAPIAttachmentRepository(api *gzapi.GZAPI) AttachmentRepository {
	return &GZAPIAttachmentRepository{
		api: api,
	}
}

// UploadAttachment uploads a file as an attachment to a challenge
func (r *GZAPIAttachmentRepository) UploadAttachment(ctx context.Context, challengeID int, filePath string) (*gzapi.Attachment, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.Wrapf(errors.ErrFileNotFound, "file not found: %s", filePath)
	}

	// Read file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file: %s", filePath)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get file info: %s", filePath)
	}

	// Create attachment object
	attachment := &gzapi.Attachment{
		ChallengeID: challengeID,
		FileName:    fileInfo.Name(),
		FileSize:    fileInfo.Size(),
		Content:     fileContent,
	}

	// Upload via API
	uploadedAttachment, err := r.api.UploadAttachment(challengeID, filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to upload attachment for challenge %d", challengeID)
	}

	log.Info("Successfully uploaded attachment %s for challenge %d", fileInfo.Name(), challengeID)
	return uploadedAttachment, nil
}

// DeleteAttachment deletes an attachment
func (r *GZAPIAttachmentRepository) DeleteAttachment(ctx context.Context, attachmentID int) error {
	err := r.api.DeleteAttachment(attachmentID)
	if err != nil {
		return errors.Wrapf(err, "failed to delete attachment %d", attachmentID)
	}

	log.Info("Successfully deleted attachment %d", attachmentID)
	return nil
}

// GetAttachments retrieves all attachments for a challenge
func (r *GZAPIAttachmentRepository) GetAttachments(ctx context.Context, challengeID int) ([]gzapi.Attachment, error) {
	attachments, err := r.api.GetAttachments(challengeID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get attachments for challenge %d", challengeID)
	}

	return attachments, nil
}