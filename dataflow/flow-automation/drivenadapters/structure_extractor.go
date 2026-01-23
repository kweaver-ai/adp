package drivenadapters

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/ide-go-lib/telemetry/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type StructureExtractor interface {
	FileParse(ctx context.Context, fileUrl, fileName string) (*FileParseResultItem, []*ContentItem, error)
}

type structureExtractor struct {
	baseURL string
	config  *common.StructureExtractor
	client  *http.Client
}

var (
	structureExtractorIns  StructureExtractor
	structureextractorOnce sync.Once
)

func NewStructureExtractor() StructureExtractor {
	structureextractorOnce.Do(func() {
		config := common.NewConfig().StructureExtractor
		structureExtractorIns = &structureExtractor{
			baseURL: fmt.Sprintf("http://%s:%v", config.PrivateHost, config.PrivatePort),
			config:  &config,
			client: &http.Client{
				Transport: otelhttp.NewTransport(&http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}),
				Timeout: 30 * time.Minute, // 整体请求超时（文件解析可能耗时较长）
			},
		}
	})

	return structureExtractorIns
}

type FileParseResultItem struct {
	Images      map[string]string `json:"images"`
	MdContent   string            `json:"md_content"`
	ContentList string            `json:"content_list"`
}

type ContentItem struct {
	// 通用字段 (所有类型都有)
	Type    string `json:"type"`     // text, image, table, equation, code, list, header, footer, etc.
	PageIdx int    `json:"page_idx"` // 页码,从 0 开始
	Bbox    [4]int `json:"bbox"`     // [x0, y0, x1, y1], 归一化到 0-1000

	// 文本类型字段
	Text      string `json:"text,omitempty"`       // 文本内容
	TextLevel *int   `json:"text_level,omitempty"` // 标题层级,0=正文,1-N=标题

	// 图片类型字段
	ImgPath       string   `json:"img_path,omitempty"`       // 图片路径
	ImageCaption  []string `json:"image_caption,omitempty"`  // 图片描述
	ImageFootnote []string `json:"image_footnote,omitempty"` // 图片脚注

	// 公式类型字段
	TextFormat string `json:"text_format,omitempty"` // "latex"

	// 表格类型字段
	TableCaption  []string `json:"table_caption,omitempty"`  // 表格标题
	TableFootnote []string `json:"table_footnote,omitempty"` // 表格脚注
	TableBody     string   `json:"table_body,omitempty"`     // HTML 格式表格

	// 代码类型字段 (VLM 后端)
	SubType     string   `json:"sub_type,omitempty"`     // "code", "algorithm", "text", "ref_text"
	CodeCaption []string `json:"code_caption,omitempty"` // 代码标题
	CodeBody    string   `json:"code_body,omitempty"`    // 代码内容
	GuessLang   string   `json:"guess_lang,omitempty"`   // 编程语言

	// 列表类型字段 (VLM 后端)
	ListItems []string `json:"list_items,omitempty"` // 列表项
}

type FileParseResult struct {
	Results map[string]*FileParseResultItem `json:"results"`
}

func (s *structureExtractor) FileParse(ctx context.Context, fileUrl, fileName string) (*FileParseResultItem, []*ContentItem, error) {
	log := traceLog.WithContext(ctx)

	// 1. 先下载文件（建立连接，获取 response body）
	fileResp, err := s.client.Get(fileUrl)
	if err != nil {
		log.Warnf("FileParse download file err: %s, url: %s", err.Error(), fileUrl)
		return nil, nil, err
	}
	defer fileResp.Body.Close()

	// 2. 使用 io.Pipe 进行流式传输（避免大文件占用过多内存）
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// 先获取 Content-Type（必须在写入之前获取，确保 boundary 正确）
	contentType := writer.FormDataContentType()

	// 用于捕获 goroutine 中的错误
	errChan := make(chan error, 1)

	go func() {
		var writeErr error
		defer func() {
			writer.Close() // 先关闭 writer
			if writeErr != nil {
				pw.CloseWithError(writeErr)
			} else {
				pw.Close()
			}
			errChan <- writeErr
		}()

		// 写入表单字段（用 map+循环简化逻辑）
		fields := map[string]string{
			"output_dir":          s.config.OutputDir,
			"backend":             s.config.Backend,
			"server_url":          s.config.ServerUrl,
			"lang_list":           "ch",
			"return_md":           "true",
			"return_content_list": "true",
			"return_images":       "false",
		}
		for k, v := range fields {
			if writeErr = writer.WriteField(k, v); writeErr != nil {
				return
			}
		}

		// 创建文件表单字段
		part, err := writer.CreateFormFile("files", fileName)
		if err != nil {
			writeErr = err
			return
		}

		// 流式复制文件内容
		_, writeErr = io.Copy(part, fileResp.Body)
	}()

	// 3. 发送请求（会从 pipe reader 读取数据）
	target := fmt.Sprintf("%s/file_parse", s.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", target, pr)
	if err != nil {
		pr.Close() // 关闭 reader，让 goroutine 退出
		<-errChan  // 等待 goroutine 结束
		log.Warnf("FileParse create request err: %s", err.Error())
		return nil, nil, err
	}

	req.Header.Set("Content-Type", contentType)

	log.Infof("FileParse sending request to %s, Content-Type: %s, fileName: %s", target, contentType, fileName)
	resp, err := s.client.Do(req)

	// 等待写入 goroutine 完成，获取写入错误
	writeErr := <-errChan
	if writeErr != nil {
		log.Warnf("FileParse write goroutine error: %s", writeErr.Error())
		if err == nil {
			err = writeErr
		}
	}

	if err != nil {
		traceLog.WithContext(ctx).Warnf("FileParse err: %s", err.Error())
		return nil, nil, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("FileParse err: %s", err.Error())
		return nil, nil, err
	}

	respCode := resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = errors.ExHTTPError{
			Body:   string(data),
			Status: respCode,
		}
		return nil, nil, err
	}

	var res FileParseResult
	err = json.Unmarshal(data, &res)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("FileParse err: %s", err.Error())
		return nil, nil, err
	}

	for _, item := range res.Results {
		var contentList []*ContentItem
		if err := json.Unmarshal([]byte(item.ContentList), &contentList); err != nil {
			return item, nil, nil
		}

		return item, contentList, nil
	}

	return nil, nil, nil
}
