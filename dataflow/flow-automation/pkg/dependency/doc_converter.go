package dependency

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/store"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

const dfsDocPrefix = "dfs://"

const (
	pdfMarker   = "%PDF-"
	maxPeekSize = 8 * 1024
)

type GotenbergCallbackRequest struct {
	TaskID      string
	DocID       string
	FileName    string
	ContentType string
	Body        io.Reader
	Size        int64
}

type DocumentConverter interface {
	ExtractFullText(ctx context.Context, docID string) (map[string]any, error)
	ConvertToPDF(ctx context.Context, taskID, docID string) error
	HandleGotenbergCallback(ctx context.Context, req *GotenbergCallbackRequest) (map[string]any, error)
}

type documentObjectStorage interface {
	GetAvaildOSS(ctx context.Context) (string, error)
	UploadFile(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader, size int64) error
	GetDownloadURL(ctx context.Context, ossID, key string, expires int64, internalRequest bool, opts ...drivenadapters.OssOpt) (string, error)
}

type documentConverter struct {
	flowFileDao          rds.FlowFileDao
	flowStorageDao       rds.FlowStorageDao
	objectStorage        documentObjectStorage
	textExtractor        drivenadapters.PlainTextExtractor
	pdfConverter         drivenadapters.PDFConverter
	rawHTTPClient        *http.Client
	gtbgcallbackURL      string
	gtbgcallbackErrorURL string
}

type resolvedFlowFile struct {
	file    *rds.FlowFile
	storage *rds.FlowStorage
}

var nextDocumentIDFunc = store.NextID

var (
	documentConverterOnce sync.Once
	documentConverterIns  DocumentConverter
)

func NewDocumentConverter() DocumentConverter {
	documentConverterOnce.Do(func() {
		config := common.NewConfig()
		gtbgcallbackURL := fmt.Sprintf("http://%s:%s/api/automation/v1/gotenberg/callback/success", config.ContentAutomation.PrivateHost, config.ContentAutomation.PrivatePort)
		gtbgcallbackErrorURL := fmt.Sprintf("http://%s:%s/api/automation/v1/gotenberg/callback/error", config.ContentAutomation.PrivateHost, config.ContentAutomation.PrivatePort)
		documentConverterIns = &documentConverter{
			flowFileDao:          rds.GetFlowFileDao(),
			flowStorageDao:       rds.GetFlowStorageDao(),
			objectStorage:        drivenadapters.NewOssGateWay(),
			textExtractor:        drivenadapters.NewTikaPlainTextExtractor(),
			pdfConverter:         drivenadapters.NewGotenberg(),
			rawHTTPClient:        drivenadapters.NewOtelRawHTTPClient(),
			gtbgcallbackURL:      gtbgcallbackURL,
			gtbgcallbackErrorURL: gtbgcallbackErrorURL,
		}
	})

	return documentConverterIns
}

func (c *documentConverter) ExtractFullText(ctx context.Context, docID string) (map[string]any, error) {
	source, err := c.resolveFlowFile(ctx, docID)
	if err != nil {
		return nil, err
	}

	reader, _, err := c.openSourceReader(ctx, source)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	text, err := c.textExtractor.ExtractPlainText(ctx, source.file.Name, reader)
	if err != nil {
		return nil, err
	}

	fileID, err := c.persistDerivedFile(
		ctx,
		source,
		replaceFileExt(source.file.Name, ".txt"),
		"text/plain; charset=utf-8",
		bytes.NewReader([]byte(text)),
		int64(len(text)),
	)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"file_id": fileID,
		"text":    text,
	}, nil
}

func (c *documentConverter) ConvertToPDF(ctx context.Context, taskID, docID string) error {
	source, err := c.resolveFlowFile(ctx, docID)
	if err != nil {
		return err
	}

	reader, _, err := c.openSourceReader(ctx, source)
	if err != nil {
		return err
	}
	defer reader.Close()

	return c.pdfConverter.ConvertToPDF(ctx, &drivenadapters.GotenbergConvertRequest{
		FileName:        source.file.Name,
		File:            reader,
		WebhookURL:      c.gtbgcallbackURL,
		WebhookErrorURL: c.gtbgcallbackErrorURL,
		WebhookHeaders:  map[string]string{"X-Task-ID": taskID, "X-Source-Doc-ID": docID},
	})
}

func (c *documentConverter) HandleGotenbergCallback(ctx context.Context, req *GotenbergCallbackRequest) (map[string]any, error) {
	if req == nil {
		return nil, fmt.Errorf("gotenberg callback request is nil")
	}
	if req.TaskID == "" {
		return nil, fmt.Errorf("missing X-Task-ID")
	}
	if req.DocID == "" {
		return nil, fmt.Errorf("missing X-Source-Doc-ID")
	}
	if req.Body == nil {
		return nil, fmt.Errorf("callback body is nil")
	}
	if strings.Contains(strings.ToLower(req.ContentType), "application/json") {
		return nil, fmt.Errorf("gotenberg callback content-type %q is not supported", req.ContentType)
	}

	source, err := c.resolveFlowFile(ctx, req.DocID)
	if err != nil {
		return nil, err
	}

	fileName := req.FileName
	if fileName == "" {
		fileName = replaceFileExt(source.file.Name, ".pdf")
	}

	fileID, err := c.persistDerivedFile(ctx, source, fileName, "application/pdf", req.Body, -1)
	if err != nil {
		return nil, err
	}

	return map[string]any{"file_id": fileID}, nil
}

func (c *documentConverter) resolveFlowFile(ctx context.Context, docID string) (*resolvedFlowFile, error) {
	fileID, err := parseDFSDocID(docID)
	if err != nil {
		return nil, err
	}

	file, err := c.flowFileDao.GetByID(ctx, fileID)
	if file == nil {
		return nil, ierrors.NewIError(ierrors.FileNotFound, "", map[string]any{"docid": docID})
	}

	storage, err := c.flowStorageDao.GetByID(ctx, file.StorageID)
	if storage == nil {
		return nil, ierrors.NewIError(ierrors.FileNotFound, "", map[string]any{"docid": docID, "storageid": file.StorageID})
	}

	return &resolvedFlowFile{file: file, storage: storage}, nil
}

func (c *documentConverter) openSourceReader(ctx context.Context, source *resolvedFlowFile) (io.ReadCloser, int64, error) {
	downloadURL, err := c.objectStorage.GetDownloadURL(ctx, source.storage.OssID, source.storage.ObjectKey, 0, true)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.rawHTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("open read stream failed, status=%d body=%s", resp.StatusCode, string(body))
	}

	return resp.Body, resp.ContentLength, nil
}

func (c *documentConverter) persistDerivedFile(
	ctx context.Context,
	source *resolvedFlowFile,
	fileName string,
	contentType string,
	body io.Reader,
	size int64,
) (string, error) {
	uploadReader := body
	uploadSize := size
	var uploadCloser io.Closer

	if size <= 0 {
		reader, bufferedSize, err := bufferToTempFile(body)
		if err != nil {
			return "", err
		}
		uploadReader = reader
		uploadSize = bufferedSize
		uploadCloser = reader
	}
	if uploadCloser != nil {
		defer uploadCloser.Close()
	}

	ossID, err := c.objectStorage.GetAvaildOSS(ctx)
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	storageID, _ := utils.GetUniqueID()
	fileID, _ := utils.GetUniqueID()
	fileName = sanitizeFileName(fileName)
	objectKey := fmt.Sprintf("%s/flow_files/%d/%s", common.NewConfig().Server.StoragePrefix, fileID, fileName)

	if err = c.objectStorage.UploadFile(ctx, ossID, objectKey, true, uploadReader, uploadSize); err != nil {
		return "", err
	}

	if err = c.flowStorageDao.Insert(ctx, &rds.FlowStorage{
		ID:          storageID,
		OssID:       ossID,
		ObjectKey:   objectKey,
		Name:        fileName,
		ContentType: contentType,
		Size:        uint64(uploadSize),
		Status:      rds.FlowStorageStatusNormal,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		return "", err
	}

	if err = c.flowFileDao.Insert(ctx, &rds.FlowFile{
		ID:            fileID,
		DagID:         source.file.DagID,
		DagInstanceID: source.file.DagInstanceID,
		StorageID:     storageID,
		Status:        rds.FlowFileStatusReady,
		Name:          fileName,
		ExpiresAt:     source.file.ExpiresAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%d", dfsDocPrefix, fileID), nil
}

func parseDFSDocID(docID string) (uint64, error) {
	if !strings.HasPrefix(docID, dfsDocPrefix) {
		return 0, fmt.Errorf("invalid dfs doc id: %s", docID)
	}

	fileID, err := strconv.ParseUint(strings.TrimPrefix(docID, dfsDocPrefix), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid dfs doc id: %s", docID)
	}

	return fileID, nil
}

func readAtMost(r io.Reader, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

type autoCleanTempFile struct {
	*os.File
}

func (f *autoCleanTempFile) Close() error {
	name := f.Name()
	err := f.File.Close()
	_ = os.Remove(name)
	return err
}

func bufferToTempFile(r io.Reader) (io.ReadCloser, int64, error) {
	tmp, err := os.CreateTemp("", "doc-converter-*")
	if err != nil {
		return nil, 0, err
	}

	written, err := io.Copy(tmp, r)
	if err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return nil, 0, err
	}

	if _, err = tmp.Seek(0, io.SeekStart); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return nil, 0, err
	}

	return &autoCleanTempFile{File: tmp}, written, nil
}

func sanitizeFileName(fileName string) string {
	baseName := path.Base(fileName)
	baseName = strings.ReplaceAll(baseName, " ", "_")
	if baseName == "." || baseName == "/" || baseName == "" {
		return fmt.Sprintf("file_%d", time.Now().UnixNano())
	}
	return baseName
}

func replaceFileExt(fileName, newExt string) string {
	ext := path.Ext(fileName)
	if ext == "" {
		return fileName + newExt
	}
	return strings.TrimSuffix(fileName, ext) + newExt
}
