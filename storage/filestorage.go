package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

type FileStorage struct {
	path string
}

func (f *FileStorage) OpenStorage(ctx context.Context) error {
	return nil
}
func (f *FileStorage) CloseStorage(ctx context.Context) {

}

func (f *FileStorage) UploadFile(ctx context.Context, partition int, filePath string, fileStream multipart.File) error {
	file, err := os.Create(f.buildFilePath(ctx, partition, filePath))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, fileStream)
	if err != nil {
		return err
	}
	return nil
}
func (f *FileStorage) UploadFileLAN(ctx context.Context, partition int, filePath string, contentType string, fileStream io.Reader) error {
	file, err := os.Create(f.buildFilePath(ctx, partition, filePath))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, fileStream)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) UploadFileBytes(ctx context.Context, partition int, filePath string, fileStream *bytes.Buffer) error{
	file, err := os.Create(f.buildFilePath(ctx, partition, filePath))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, fileStream)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) DownloadFile(ctx context.Context, partition int, filePath string) (io.Reader, error) {
	file, err := os.OpenFile(f.buildFilePath(ctx, partition, filePath), os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}
func (f *FileStorage) ExitsFile(ctx context.Context, partition int, filePath string) bool {
	info, err := os.Stat(f.buildFilePath(ctx, partition, filePath))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
func (f *FileStorage) CopyFile(ctx context.Context, source, target string) error{
	sourcePath := fmt.Sprintf("%s/%s", f.path, source)
	targetPath := fmt.Sprintf("%s/%s", f.path, target)
	f1, err := os.Open(sourcePath)
	if err != nil{
		return err
	}
	f2, err := os.Open(targetPath)
	if err != nil{
		return err
	}
	_, err = io.Copy(f2, f1)
	return err
}

func (f FileStorage) buildFilePath(ctx context.Context, partition int, path string) string {
	return fmt.Sprintf("%s/%d/%s", f.path, partition, path)
}
func (f FileStorage) GetFilePath(ctx context.Context, partition int) string {
	return f.buildFilePath(ctx, partition, "")
}
func (f FileStorage) GetFileTempPath(ctx context.Context, partition int, filePath string) (string, error) {
	return f.buildFilePath(ctx, partition, filePath), nil
}

func (f *FileStorage) GetUploadFileTempPath(ctx context.Context, partition int, fileName string) (string ,error) {
	return f.buildFilePath(ctx, partition, fileName), nil
}

func (f *FileStorage) GetUploadFileTempRawPath(ctx context.Context, tempPath string, fileName string) (string ,error) {
	return fmt.Sprintf("%s/%s/%s", f.path, tempPath, fileName), nil
}

func newFileStorage(path string) IStorage {
	return &FileStorage{path: path}
}
