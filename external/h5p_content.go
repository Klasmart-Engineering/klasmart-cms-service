package external

type H5PContent struct {
	ParentID     string   `json:"parent_id"` // add: 2021-08-14
	ContentID    string   `json:"content_id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	FileType     FileType `json:"fileType"`
	H5PID        string   `json:"h5p_id"`
	SubContentID string   `json:"subcontent_id"`
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
