//nolint:revive // Function and variable names follow project conventions
package challenge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/fileutil"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/log"
)

type assetsCache struct {
	once    sync.Once
	loadErr error

	mu     sync.RWMutex
	byHash map[string]gzapi.FileInfo
}

var assetsCacheByAPI sync.Map

func cacheKeyForAPI(api *gzapi.GZAPI) string {
	if api == nil || api.Creds == nil {
		return ""
	}
	return fmt.Sprintf("%s|%s", strings.TrimSpace(api.Url), api.Creds.Username)
}

func getAssetsCache(api *gzapi.GZAPI) *assetsCache {
	key := cacheKeyForAPI(api)
	if key == "" {
		return nil
	}

	if cached, ok := assetsCacheByAPI.Load(key); ok {
		if c, ok2 := cached.(*assetsCache); ok2 {
			return c
		}
	}

	newCache := &assetsCache{
		byHash: make(map[string]gzapi.FileInfo),
	}
	actual, _ := assetsCacheByAPI.LoadOrStore(key, newCache)
	if c, ok := actual.(*assetsCache); ok {
		return c
	}
	return newCache
}

func (c *assetsCache) ensureLoaded(api *gzapi.GZAPI) error {
	if c == nil {
		return nil
	}
	c.once.Do(func() {
		assets, err := api.GetAssets()
		if err != nil {
			c.loadErr = fmt.Errorf("failed to get assets: %w", err)
			return
		}
		c.mu.Lock()
		for i := range assets {
			c.byHash[assets[i].Hash] = assets[i]
		}
		c.mu.Unlock()
	})
	return c.loadErr
}

func (c *assetsCache) get(hash string) (*gzapi.FileInfo, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.byHash[hash]
	if !ok {
		return nil, false
	}
	out := v
	return &out, true
}

func (c *assetsCache) set(file gzapi.FileInfo) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.byHash[file.Hash] = file
	c.mu.Unlock()
}

func HandleChallengeAttachments(challengeConf config.ChallengeYaml, challengeData *gzapi.Challenge, api *gzapi.GZAPI) error {
	log.DebugH3("Processing attachments for challenge: %s", challengeConf.Name)

	switch {
	case challengeConf.Provide != nil:
		log.DebugH3("Challenge %s has attachment: %s", challengeConf.Name, *challengeConf.Provide)

		switch {
		case strings.HasPrefix(*challengeConf.Provide, "http"):
			log.DebugH3("Creating remote attachment for %s: %s", challengeConf.Name, *challengeConf.Provide)
			if err := challengeData.CreateAttachment(gzapi.CreateAttachmentForm{
				AttachmentType: "Remote",
				RemoteUrl:      *challengeConf.Provide,
			}); err != nil {
				log.Error("Failed to create remote attachment for %s: %v", challengeConf.Name, err)
				return fmt.Errorf("remote attachment creation failed for %s: %w", challengeConf.Name, err)
			}
			log.DebugH3("Successfully created remote attachment for %s", challengeConf.Name)
		default:
			log.DebugH3("Processing local attachment for %s: %s", challengeConf.Name, *challengeConf.Provide)
			return HandleLocalAttachment(challengeConf, challengeData, api)
		}
	case challengeData.Attachment != nil:
		log.DebugH3("Removing existing attachment for %s", challengeConf.Name)
		if err := challengeData.CreateAttachment(gzapi.CreateAttachmentForm{
			AttachmentType: "None",
		}); err != nil {
			log.Error("Failed to remove attachment for %s: %v", challengeConf.Name, err)
			return fmt.Errorf("attachment removal failed for %s: %w", challengeConf.Name, err)
		}
		log.DebugH3("Successfully removed attachment for %s", challengeConf.Name)
	default:
		log.DebugH3("No attachment processing needed for %s", challengeConf.Name)
	}

	log.DebugH3("Attachment processing completed for %s", challengeConf.Name)
	return nil
}

func HandleLocalAttachment(challengeConf config.ChallengeYaml, challengeData *gzapi.Challenge, api *gzapi.GZAPI) error {
	log.DebugH3("Creating local attachment for %s", challengeConf.Name)

	zipFilename := "dist.zip"
	// Write zip to temp dir to avoid triggering watcher events inside challenge dir
	zipOutput := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s", fileutil.NormalizeFileName(challengeConf.Name), zipFilename))
	attachmentPath := filepath.Join(challengeConf.Cwd, *challengeConf.Provide)

	// Artifact path that will be used for upload/uniqueness processing
	var artifactPath string
	var artifactBase string

	log.DebugH3("Checking attachment path: %s", attachmentPath)
	if info, err := os.Stat(attachmentPath); err != nil || info.IsDir() {
		log.DebugH3("Creating zip file for %s from: %s", challengeConf.Name, attachmentPath)
		if err := fileutil.ZipSource(attachmentPath, zipOutput); err != nil {
			log.Error("Failed to create zip for %s: %v", challengeConf.Name, err)
			return fmt.Errorf("zip creation failed for %s: %w", challengeConf.Name, err)
		}
		log.DebugH3("Successfully created zip file: %s", zipOutput)
		// Use the temp zip directly as the artifact, do not write into challenge directory
		artifactPath = zipOutput
		artifactBase = filepath.Base(zipOutput)
	} else {
		log.DebugH3("Using existing file: %s", attachmentPath)
		artifactPath = attachmentPath
		artifactBase = filepath.Base(attachmentPath)
	}

	artifactHash, err := fileutil.GetFileHashHex(artifactPath)
	if err != nil {
		return fmt.Errorf("failed to hash attachment for %s: %w", challengeConf.Name, err)
	}

	// Skip all copy/upload work when the challenge already points at the same file hash.
	if challengeData.Attachment != nil && strings.Contains(challengeData.Attachment.Url, artifactHash) {
		log.DebugH3("Attachment for %s is unchanged (hash: %s)", challengeConf.Name, artifactHash)
		if strings.HasSuffix(zipOutput, ".zip") {
			_ = os.Remove(zipOutput)
		}
		return nil
	}

	// Create a unique attachment file name while preserving extension
	ext := filepath.Ext(artifactBase)
	nameNoExt := strings.TrimSuffix(artifactBase, ext)
	sanitizedBase := fileutil.NormalizeFileName(fmt.Sprintf("%s_%s", challengeConf.Name, nameNoExt))
	uniqueFilename := sanitizedBase + ext
	uniqueFilePath := filepath.Join(os.TempDir(), uniqueFilename)

	log.DebugH3("Creating unique attachment file: %s", uniqueFilePath)

	// Copy the artifact and append challenge metadata to make it unique
	if err := CreateUniqueAttachmentFile(artifactPath, uniqueFilePath, challengeConf.Name); err != nil {
		log.Error("Failed to create unique attachment file for %s: %v", challengeConf.Name, err)
		return fmt.Errorf("unique file creation failed for %s: %w", challengeConf.Name, err)
	}

	log.DebugH3("Creating/checking assets for %s", challengeConf.Name)
	fileinfo, err := CreateAssetsIfNotExistOrDifferentWithHash(uniqueFilePath, artifactHash, api)
	if err != nil {
		_ = os.Remove(uniqueFilePath) // Clean up on error
		log.Error("Failed to create/check assets for %s: %v", challengeConf.Name, err)
		return fmt.Errorf("asset creation failed for %s: %w", challengeConf.Name, err)
	}
	log.DebugH3("Asset info for %s: Hash=%s, Name=%s", challengeConf.Name, fileinfo.Hash, fileinfo.Name)

	// Check if the challenge already has the same attachment hash
	if challengeData.Attachment != nil && strings.Contains(challengeData.Attachment.Url, fileinfo.Hash) {
		log.DebugH3("Attachment for %s is unchanged (hash: %s)", challengeConf.Name, fileinfo.Hash)
	} else {
		var attachmentUrl string
		if challengeData.Attachment != nil {
			attachmentUrl = challengeData.Attachment.Url
		}
		log.DebugH3("Updating attachment for %s (hash: %s, current: %s)", challengeConf.Name, fileinfo.Hash, attachmentUrl)

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
			log.DebugH3("Successfully created local attachment for %s", challengeConf.Name)
		}
	}

	// Clean up temporary files
	if strings.HasSuffix(zipOutput, ".zip") {
		log.DebugH3("Cleaning up temporary zip file: %s", zipOutput)
		_ = os.Remove(zipOutput)
	}

	// Clean up the unique file after successful upload
	log.DebugH3("Cleaning up unique attachment file: %s", uniqueFilePath)
	_ = os.Remove(uniqueFilePath)

	log.DebugH3("Local attachment processing completed for %s", challengeConf.Name)
	return nil
}

// CreateUniqueAttachmentFile creates a unique version of the attachment file by appending metadata
func CreateUniqueAttachmentFile(srcPath, dstPath, challengeName string) error {
	_ = challengeName // kept for backward-compatible signature; uniqueness must not change bytes
	return fileutil.CopyFile(srcPath, dstPath)
}

// CreateAssetsIfNotExistOrDifferent creates assets if they don't exist or are different
func CreateAssetsIfNotExistOrDifferent(filePath string, api *gzapi.GZAPI) (*gzapi.FileInfo, error) {
	hash, err := fileutil.GetFileHashHex(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file hash: %w", err)
	}
	return CreateAssetsIfNotExistOrDifferentWithHash(filePath, hash, api)
}

// CreateAssetsIfNotExistOrDifferentWithHash creates assets if they don't exist or are different,
// using a precomputed hash to avoid re-hashing the same file in hot paths.
func CreateAssetsIfNotExistOrDifferentWithHash(filePath, hash string, api *gzapi.GZAPI) (*gzapi.FileInfo, error) {
	if hash == "" {
		return nil, fmt.Errorf("file hash cannot be empty")
	}

	cache := getAssetsCache(api)
	if err := cache.ensureLoaded(api); err != nil {
		return nil, err
	}

	if existing, ok := cache.get(hash); ok {
		return existing, nil
	}

	// Asset doesn't exist, create it
	newAssets, err := api.CreateAssets(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	if len(newAssets) == 0 {
		return nil, fmt.Errorf("asset creation returned empty result")
	}

	cache.set(newAssets[0])
	return &newAssets[0], nil
}
