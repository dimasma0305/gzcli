package gzapi

// FileInfo represents file metadata from the GZCTF platform
type FileInfo struct {
	Hash string `json:"hash"`
	Name string `json:"name"`
}

// CreateAssets uploads asset files to the GZCTF platform
func (cs *GZAPI) CreateAssets(file string) ([]FileInfo, error) {
	var fileInfo []FileInfo
	if err := cs.postMultiPart("/api/assets", file, &fileInfo); err != nil {
		return nil, err
	}
	return fileInfo, nil
}

// GetAssets retrieves all assets from the GZCTF platform
func (cs *GZAPI) GetAssets() ([]FileInfo, error) {
	var data struct {
		Data []FileInfo `json:"data"`
	}
	if err := cs.get("/api/admin/files", &data); err != nil {
		return nil, err
	}
	return data.Data, nil
}
