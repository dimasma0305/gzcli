package challenge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/utils"
	"github.com/dimasma0305/gzcli/internal/log"
)

func HandleChallengeAttachments(challengeConf ChallengeYaml, challengeData *gzapi.Challenge, api *gzapi.GZAPI) error {
	log.InfoH3("Processing attachments for challenge: %s", challengeConf.Name)

	if challengeConf.Provide != nil {
		log.InfoH3("Challenge %s has attachment: %s", challengeConf.Name, *challengeConf.Provide)

		if strings.HasPrefix(*challengeConf.Provide, "http") {
			log.InfoH3("Creating remote attachment for %s: %s", challengeConf.Name, *challengeConf.Provide)
			if err := challengeData.CreateAttachment(gzapi.CreateAttachmentForm{
				AttachmentType: "Remote",
				RemoteUrl:      *challengeConf.Provide,
			}); err != nil {
				log.Error("Failed to create remote attachment for %s: %v", challengeConf.Name, err)
				return fmt.Errorf("remote attachment creation failed for %s: %w", challengeConf.Name, err)
			}
			log.InfoH3("Successfully created remote attachment for %s", challengeConf.Name)
		} else {
			log.InfoH3("Processing local attachment for %s: %s", challengeConf.Name, *challengeConf.Provide)
			return HandleLocalAttachment(challengeConf, challengeData, api)
		}
	} else if challengeData.Attachment != nil {
		log.InfoH3("Removing existing attachment for %s", challengeConf.Name)
		if err := challengeData.CreateAttachment(gzapi.CreateAttachmentForm{
			AttachmentType: "None",
		}); err != nil {
			log.Error("Failed to remove attachment for %s: %v", challengeConf.Name, err)
			return fmt.Errorf("attachment removal failed for %s: %w", challengeConf.Name, err)
		}
		log.InfoH3("Successfully removed attachment for %s", challengeConf.Name)
	} else {
		log.InfoH3("No attachment processing needed for %s", challengeConf.Name)
	}

	log.InfoH3("Attachment processing completed for %s", challengeConf.Name)
	return nil
}

func HandleLocalAttachment(challengeConf ChallengeYaml, challengeData *gzapi.Challenge, api *gzapi.GZAPI) error {
	log.InfoH3("Creating local attachment for %s", challengeConf.Name)

	zipFilename := "dist.zip"
	// Write zip to temp dir to avoid triggering watcher events inside challenge dir
	zipOutput := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s", utils.NormalizeFileName(challengeConf.Name), zipFilename))
	attachmentPath := filepath.Join(challengeConf.Cwd, *challengeConf.Provide)

	// Artifact path that will be used for upload/uniqueness processing
	var artifactPath string
	var artifactBase string

	log.InfoH3("Checking attachment path: %s", attachmentPath)
	if info, err := os.Stat(attachmentPath); err != nil || info.IsDir() {
		log.InfoH3("Creating zip file for %s from: %s", challengeConf.Name, attachmentPath)
		if err := utils.ZipSource(attachmentPath, zipOutput); err != nil {
			log.Error("Failed to create zip for %s: %v", challengeConf.Name, err)
			return fmt.Errorf("zip creation failed for %s: %w", challengeConf.Name, err)
		}
		log.InfoH3("Successfully created zip file: %s", zipOutput)
		// Use the temp zip directly as the artifact, do not write into challenge directory
		artifactPath = zipOutput
		artifactBase = filepath.Base(zipOutput)
	} else {
		log.InfoH3("Using existing file: %s", attachmentPath)
		artifactPath = attachmentPath
		artifactBase = filepath.Base(attachmentPath)
	}

	// Create a unique attachment file name while preserving extension
	ext := filepath.Ext(artifactBase)
	nameNoExt := strings.TrimSuffix(artifactBase, ext)
	sanitizedBase := utils.NormalizeFileName(fmt.Sprintf("%s_%s", challengeConf.Name, nameNoExt))
	uniqueFilename := sanitizedBase + ext
	uniqueFilePath := filepath.Join(os.TempDir(), uniqueFilename)

	log.InfoH3("Creating unique attachment file: %s", uniqueFilePath)

	// Copy the artifact and append challenge metadata to make it unique
	if err := CreateUniqueAttachmentFile(artifactPath, uniqueFilePath, challengeConf.Name); err != nil {
		log.Error("Failed to create unique attachment file for %s: %v", challengeConf.Name, err)
		return fmt.Errorf("unique file creation failed for %s: %w", challengeConf.Name, err)
	}

	log.InfoH3("Creating/checking assets for %s", challengeConf.Name)
	fileinfo, err := CreateAssetsIfNotExistOrDifferent(uniqueFilePath, api)
	if err != nil {
		_ = os.Remove(uniqueFilePath) // Clean up on error
		log.Error("Failed to create/check assets for %s: %v", challengeConf.Name, err)
		return fmt.Errorf("asset creation failed for %s: %w", challengeConf.Name, err)
	}
	log.InfoH3("Asset info for %s: Hash=%s, Name=%s", challengeConf.Name, fileinfo.Hash, fileinfo.Name)

	// Check if the challenge already has the same attachment hash
	if challengeData.Attachment != nil && strings.Contains(challengeData.Attachment.Url, fileinfo.Hash) {
		log.InfoH3("Attachment for %s is unchanged (hash: %s)", challengeConf.Name, fileinfo.Hash)
	} else {
		var attachmentUrl string
		if challengeData.Attachment != nil {
			attachmentUrl = challengeData.Attachment.Url
		}
		log.InfoH3("Updating attachment for %s (hash: %s, current: %s)", challengeConf.Name, fileinfo.Hash, attachmentUrl)

		// Try to create the attachment
		err := challengeData.CreateAttachment(gzapi.CreateAttachmentForm{
			AttachmentType: "Local",
			FileHash:       fileinfo.Hash,
		})

		if err != nil {
			log.Error("Failed to create local attachment for %s: %v", challengeConf.Name, err)
			_ = os.Remove(uniqueFilePath) // Clean up on error
			return fmt.Errorf("local attachment creation failed for %s: %w", challengeConf.Name, err)
		} else {
			log.InfoH3("Successfully created local attachment for %s", challengeConf.Name)
		}
	}

	// Clean up temporary files
	if strings.HasSuffix(zipOutput, ".zip") {
		log.InfoH3("Cleaning up temporary zip file: %s", zipOutput)
		_ = os.Remove(zipOutput)
	}

	// Clean up the unique file after successful upload
	log.InfoH3("Cleaning up unique attachment file: %s", uniqueFilePath)
	_ = os.Remove(uniqueFilePath)

	log.InfoH3("Local attachment processing completed for %s", challengeConf.Name)
	return nil
}

// CreateUniqueAttachmentFile creates a unique version of the attachment file by appending metadata
func CreateUniqueAttachmentFile(srcPath, dstPath, challengeName string) error {
	// Copy the original file
	if err := utils.CopyFile(srcPath, dstPath); err != nil {
		return err
	}

	// Append challenge-specific metadata to make the file unique
	file, err := os.OpenFile(dstPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	// Add a comment or metadata that makes this file unique for this challenge
	metadata := fmt.Sprintf("\n# Challenge: %s\n", challengeName)
	_, err = file.WriteString(metadata)
	return err
}

// CreateAssetsIfNotExistOrDifferent creates assets if they don't exist or are different
func CreateAssetsIfNotExistOrDifferent(filePath string, api *gzapi.GZAPI) (*gzapi.FileInfo, error) {
	hash, err := utils.GetFileHashHex(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file hash: %w", err)
	}

	// Try to get existing assets
	assets, err := api.GetAssets()
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}

	// Check if asset with same hash already exists
	for _, asset := range assets {
		if asset.Hash == hash {
			return &asset, nil
		}
	}

	// Asset doesn't exist, create it
	newAssets, err := api.CreateAssets(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	if len(newAssets) == 0 {
		return nil, fmt.Errorf("asset creation returned empty result")
	}

	return &newAssets[0], nil
}
