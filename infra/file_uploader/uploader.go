package file_uploader

import (
	"context"
)

type FileUploadConfig struct {
	AccessKeyId     string
	AccessKeySecret string
	RoleArn         string
	BucketName      string
	Endpoint        string
	Region          string
	ExpiredSeconds  int64
	AllowBucket     []string
	ServiceName     string
}

type FileUploaderToken struct {
	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
	BucketName      string
	Region          string
}

type FileUploader interface {
	GetFileUploadAccessKey(ctx context.Context) (*FileUploaderToken, error)
}
