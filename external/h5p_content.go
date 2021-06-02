package external

type H5PContent struct {
	ContentID string   `json:"content_id"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	FileType  FileType `json:"fileType"`
	H5PID     string   `json:"h5p_id"`
}

type FileType string

const (
	FileTypeImage        FileType = "Image"
	FileTypeVideo        FileType = "Video"
	FileTypeAudio        FileType = "Audio"
	FileTypeDocument     FileType = "Document"
	FileTypeH5P          FileType = "H5P"
	FileTypeH5PExtension FileType = "H5PExtension"
)
