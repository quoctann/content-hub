package http

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/quoctann/content-hub/internal/domain"
	"github.com/quoctann/content-hub/pkg/config"
	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/server"
)

var uploadMediaTypes = map[string]domain.ContentType{
	"image/jpeg": domain.Image, "image/png": domain.Image, "image/webp": domain.Image, "video/mp4": domain.Video,
}

type UploadHandler struct {
	store  domain.MediaStore
	config config.Upload
	logger logger.ILogger
}

func NewUploadHandler(r server.RouterGroup, store domain.MediaStore, cfg config.Upload, l logger.ILogger) {
	h := &UploadHandler{store: store, config: cfg, logger: l}
	r.GET("/limits", h.Limits)
	r.POST("", h.Upload)
}

func (h *UploadHandler) Limits(c server.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{"enabled": h.config.Enabled, "max_batch_files": h.config.MaxBatchFiles, "max_file_bytes": h.config.MaxFileBytes, "allowed_mime_types": []string{"image/jpeg", "image/png", "image/webp", "video/mp4"}})
}

func (h *UploadHandler) Upload(c server.Context) {
	if !h.config.Enabled || h.store == nil {
		c.JSON(http.StatusNotFound, map[string]string{"error": "uploads are disabled"})
		return
	}
	req := c.Request()
	if req.ContentLength > h.config.MaxFileBytes+1<<20 {
		c.JSON(http.StatusRequestEntityTooLarge, map[string]string{"error": "file is too large"})
		return
	}
	req.Body = http.MaxBytesReader(c.ResponseWriter(), req.Body, h.config.MaxFileBytes+1<<20)
	if err := req.ParseMultipartForm(h.config.MaxFileBytes + 1<<20); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid multipart upload"})
		return
	}
	files := req.MultipartForm.File["file"]
	if len(files) != 1 || len(req.MultipartForm.File) != 1 {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "exactly one file is required"})
		return
	}
	file := files[0]
	if file.Size <= 0 {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "file is empty"})
		return
	}
	if file.Size > h.config.MaxFileBytes {
		c.JSON(http.StatusRequestEntityTooLarge, map[string]string{"error": "file is too large"})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file"})
		return
	}
	defer src.Close()
	probe := make([]byte, 512)
	n, err := io.ReadFull(src, probe)
	if err != nil && err != io.ErrUnexpectedEOF {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file"})
		return
	}
	mediaType, ok := uploadMediaTypes[http.DetectContentType(probe[:n])]
	if !ok {
		c.JSON(http.StatusUnsupportedMediaType, map[string]string{"error": "unsupported file type"})
		return
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file"})
		return
	}
	result, err := h.store.Upload(req.Context(), domain.UploadInput{Reader: src, Filename: safeFilename(file.Filename), ContentType: http.DetectContentType(probe[:n]), Size: file.Size})
	if err != nil {
		h.logger.Error(req.Context(), "media upload failed", logger.Error(err))
		c.JSON(http.StatusBadGateway, map[string]string{"error": "media storage failed"})
		return
	}
	c.JSON(http.StatusCreated, map[string]interface{}{"url": result.URL, "type": mediaType, "mime_type": http.DetectContentType(probe[:n]), "size_bytes": file.Size})
}

func safeFilename(name string) string {
	name = filepath.Base(name)
	if name == "." || name == "" {
		return "upload"
	}
	return strings.ReplaceAll(name, "\x00", "")
}
