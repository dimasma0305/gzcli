package service

import (
	"fmt"
	"strings"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
	"github.com/dimasma0305/gzcli/internal/log"
)

// attachmentService implements AttachmentService
type attachmentService struct {
	attachmentRepo repository.AttachmentRepository
	errorHandler   *ErrorHandler
}

// NewAttachmentService creates a new attachment service
func NewAttachmentService(attachmentRepo repository.AttachmentRepository) AttachmentService {
	return &attachmentService{
		attachmentRepo: attachmentRepo,
		errorHandler:   NewErrorHandler(),
	}
}

// HandleAttachments handles attachments for a challenge
func (s *attachmentService) HandleAttachments(challengeConf config.ChallengeYaml, challenge *gzapi.Challenge) error {
	log.InfoH3("Processing attachments for challenge: %s", challengeConf.Name)

	switch {
	case challengeConf.Provide != nil:
		log.InfoH3("Challenge %s has attachment: %s", challengeConf.Name, *challengeConf.Provide)

		switch {
		case strings.HasPrefix(*challengeConf.Provide, "http"):
			log.InfoH3("Creating remote attachment for %s: %s", challengeConf.Name, *challengeConf.Provide)
			if err := s.Create(challenge.Id, gzapi.CreateAttachmentForm{
				AttachmentType: "Remote",
				RemoteUrl:      *challengeConf.Provide,
			}); err != nil {
				log.Error("Failed to create remote attachment for %s: %v", challengeConf.Name, err)
				return s.errorHandler.Wrap(err, "remote attachment creation failed")
			}
			log.InfoH3("Successfully created remote attachment for %s", challengeConf.Name)
		default:
			log.InfoH3("Processing local attachment for %s: %s", challengeConf.Name, *challengeConf.Provide)
			// This would need to be implemented based on the existing HandleLocalAttachment logic
			return fmt.Errorf("local attachment handling not implemented")
		}
	case challenge.Attachment != nil:
		log.InfoH3("Removing existing attachment for %s", challengeConf.Name)
		if err := s.Create(challenge.Id, gzapi.CreateAttachmentForm{
			AttachmentType: "None",
		}); err != nil {
			log.Error("Failed to remove attachment for %s: %v", challengeConf.Name, err)
			return s.errorHandler.Wrap(err, "attachment removal failed")
		}
		log.InfoH3("Successfully removed attachment for %s", challengeConf.Name)
	default:
		log.InfoH3("No attachment processing needed for %s", challengeConf.Name)
	}

	log.InfoH3("Attachment processing completed for %s", challengeConf.Name)
	return nil
}

// Create creates a new attachment
func (s *attachmentService) Create(challengeID int, attachment gzapi.CreateAttachmentForm) error {
	return s.attachmentRepo.Create(challengeID, attachment)
}

// Delete deletes an attachment
func (s *attachmentService) Delete(challengeID int, attachmentID int) error {
	return s.attachmentRepo.Delete(challengeID, attachmentID)
}

// List returns all attachments for a challenge
func (s *attachmentService) List(challengeID int) ([]gzapi.Attachment, error) {
	return s.attachmentRepo.List(challengeID)
}