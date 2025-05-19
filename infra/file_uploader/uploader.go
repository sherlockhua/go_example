package file_uploader

import (
	"context"
)

type FileUploadConfig struct {
	AccessKeyId     string
	AccessKeySecret string
	RoleArn         string
	BucketName      string
	StsEndpoint     string
	OssEndpoint     string
	Region          string
	ExpiredSeconds  int64
	AllowBucket     []string
	ServiceName     string
	CallbackUrl     string
}

type FileUploaderToken struct {
	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
	BucketName      string
	Region          string
	Endpoint        string
	ObjectKey       string
	CallbackInfo    CallbackInfo
}

type FileInfo struct {
	FileName string
	FileSize int64
	FileType string
}

type CallbackBody struct {
	FileName  string
	FileSize  int64
	FileType  string
	ObjectKey string
	UserId    string
}

type CallbackInfo struct {
	CallbackUrl  string
	CallbackBody string
}

type FileUploader interface {
	GetFileUploadAccessKey(ctx context.Context, userId int64, fileInfo *FileInfo) (*FileUploaderToken, error)
}
