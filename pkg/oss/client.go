package oss

import (
	"context"
	"fmt"
	"io"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Config struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string
	Directory       string
	OssDomain       string
}

type Client struct {
	bucket    *alioss.Bucket
	directory string
	domain    string
}

func NewClient(cfg Config) (*Client, error) {
	client, err := alioss.New(cfg.Endpoint, cfg.AccessKeyId, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("oss client init: %w", err)
	}
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("oss bucket: %w", err)
	}
	return &Client{
		bucket:    bucket,
		directory: cfg.Directory,
		domain:    cfg.OssDomain,
	}, nil
}

func (c *Client) BuildObjectKey(userID, slug string, version int) string {
	return fmt.Sprintf("%s/%s/%s-%d.zip", c.directory, userID, slug, version)
}

func (c *Client) BuildPublicURL(objectKey string) string {
	return fmt.Sprintf("https://%s/%s", c.domain, objectKey)
}

func (c *Client) UploadFile(_ context.Context, objectKey string, reader io.Reader) (string, error) {
	err := c.bucket.PutObject(objectKey, reader)
	if err != nil {
		return "", fmt.Errorf("oss upload %s: %w", objectKey, err)
	}
	return c.BuildPublicURL(objectKey), nil
}

func (c *Client) GetContent(_ context.Context, objectKey string) ([]byte, error) {
	body, err := c.bucket.GetObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("oss get %s: %w", objectKey, err)
	}
	defer body.Close()
	return io.ReadAll(body)
}
