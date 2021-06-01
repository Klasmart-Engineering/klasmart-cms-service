package external

type H5PContent struct {
	ContentID string   `json:"content_id"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	FileType  FileType `json:"fileType"`
}

type FileType string

const (
	FileTypeImage        FileType = "1"
	FileTypeVideo        FileType = "2"
	FileTypeAudio        FileType = "3"
	FileTypeDocument     FileType = "4"
	FileTypeH5P          FileType = "5"
	FileTypeH5PExtension FileType = "6"
)
