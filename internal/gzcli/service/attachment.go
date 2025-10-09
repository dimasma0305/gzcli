package service

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
)

// AttachmentService handles attachment business logic
type AttachmentService struct {
	attachmentRepo repository.AttachmentRepository
	challengeRepo  repository.ChallengeRepository
}

// AttachmentServiceConfig holds configuration for AttachmentService
type AttachmentServiceConfig struct {
	AttachmentRepo repository.AttachmentRepository
	ChallengeRepo  repository.ChallengeRepository
}

// NewAttachmentService creates a new AttachmentService
func NewAttachmentService(config AttachmentServiceConfig) *AttachmentService {
	return &AttachmentService{
		attachmentRepo: config.AttachmentRepo,
		challengeRepo:  config.ChallengeRepo,
	}
}

// Create creates a new attachment
func (s *AttachmentService) Create(challengeID int, attachment gzapi.CreateAttachmentForm) error {
	return s.attachmentRepo.Create(challengeID, attachment)
}

// Delete deletes an attachment
func (s *AttachmentService) Delete(challengeID int, attachmentID int) error {
	return s.attachmentRepo.Delete(challengeID, attachmentID)
}

// List returns all attachments for a challenge
func (s *AttachmentService) List(challengeID int) ([]gzapi.Attachment, error) {
	return s.attachmentRepo.List(challengeID)
}