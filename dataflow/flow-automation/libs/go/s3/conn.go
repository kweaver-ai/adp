package s3

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Connection struct {
	client          *awsS3.Client
	Name            string `json:"name"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	BucketName      string `json:"bucketName"`
	SkipVerify      bool   `json:"skipVerify"`
	DisableHTTPS    bool   `json:"disableHTTPS"`
}

func (c *S3Connection) InitClient() error {
	if c.client != nil {
		return nil
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(c.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			c.AccessKeyID,
			c.SecretAccessKey,
			"",
		)),
	)

	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	var s3Opts []func(*awsS3.Options)

	if c.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *awsS3.Options) {
			o.BaseEndpoint = aws.String(c.Endpoint)
			o.UsePathStyle = true
			o.EndpointOptions.DisableHTTPS = c.DisableHTTPS
		})
	}

	// https 模式下跳过证书验证
	if c.SkipVerify && !c.DisableHTTPS {
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		s3Opts = append(s3Opts, func(o *awsS3.Options) {
			o.HTTPClient = httpClient
		})
	}

	// 创建 S3 客户端
	c.client = awsS3.NewFromConfig(cfg, s3Opts...)

	return nil
}

// GetDownloadURL 生成一个限时的预签名下载链接
func (c *S3Connection) GetDownloadURL(ctx context.Context, key string, expires int64) (string, error) {
	if err := c.InitClient(); err != nil {
		return "", err
	}

	// 1. 创建预签名客户端
	presignClient := s3.NewPresignClient(c.client)

	// 2. 构建请求参数
	// expires 单位通常为秒
	duration := time.Duration(expires) * time.Second

	// 3. 调用 PresignGetObject
	presignedRequest, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// 返回生成的 URL
	return presignedRequest.URL, nil
}

func (c *S3Connection) DeleteObject(ctx context.Context, key string) error {
	if c.client == nil {
		return fmt.Errorf("s3 client is not initialized")
	}

	_, err := c.client.DeleteObject(ctx, &awsS3.DeleteObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(key),
	})

	return err
}

func (c *S3Connection) UploadObject(ctx context.Context, key string, file io.Reader, size int64) error {
	if c.client == nil {
		return fmt.Errorf("s3 client is not initialized")
	}

	_, err := c.client.PutObject(ctx, &awsS3.PutObjectInput{
		Bucket:        aws.String(c.BucketName),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: aws.Int64(size),
	})

	return err
}

func (c *S3Connection) GetObjectSize(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("s3 client is not initialized")
	}

	resp, err := c.client.HeadObject(ctx, &awsS3.HeadObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return 0, fmt.Errorf("failed to head object: %w", err)
	}

	return *resp.ContentLength, nil
}

func (c *S3Connection) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if c.client == nil {
		return nil, fmt.Errorf("s3 client is not initialized")
	}

	resp, err := c.client.GetObject(ctx, &awsS3.GetObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return resp.Body, nil
}
