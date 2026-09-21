package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/quoctann/content-hub/internal/domain"
)

type ImageKitStore struct {
	client     *http.Client
	endpoint   string
	privateKey string
	folder     string
}

func NewImageKitStore(endpoint, privateKey, folder string, timeout time.Duration) *ImageKitStore {
	return &ImageKitStore{client: &http.Client{Timeout: timeout}, endpoint: endpoint, privateKey: privateKey, folder: folder}
}

func (s *ImageKitStore) Upload(ctx context.Context, input domain.UploadInput) (domain.UploadResult, error) {
	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, reader)
	if err != nil {
		_ = reader.Close()
		return domain.UploadResult{}, fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	req.SetBasicAuth(s.privateKey, "")
	writeErr := make(chan error, 1)
	go func() {
		defer close(writeErr)
		defer writer.Close()
		part, err := multipartWriter.CreateFormFile("file", filepath.Base(input.Filename))
		if err == nil {
			_, err = io.Copy(part, input.Reader)
		}
		if err == nil {
			err = multipartWriter.WriteField("fileName", filepath.Base(input.Filename))
		}
		if err == nil && s.folder != "" {
			err = multipartWriter.WriteField("folder", s.folder)
		}
		if err == nil {
			err = multipartWriter.Close()
		}
		if err != nil {
			_ = writer.CloseWithError(err)
			writeErr <- err
		}
	}()
	resp, err := s.client.Do(req)
	if writeError := <-writeErr; writeError != nil {
		return domain.UploadResult{}, fmt.Errorf("write upload: %w", writeError)
	}
	if err != nil {
		return domain.UploadResult{}, fmt.Errorf("upload to ImageKit: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.UploadResult{}, fmt.Errorf("ImageKit returned %s", resp.Status)
	}
	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return domain.UploadResult{}, fmt.Errorf("decode ImageKit response: %w", err)
	}
	if result.URL == "" {
		return domain.UploadResult{}, fmt.Errorf("ImageKit response did not include a URL")
	}
	return domain.UploadResult{URL: result.URL}, nil
}
