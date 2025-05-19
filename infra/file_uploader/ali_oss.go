package file_uploader

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts "github.com/alibabacloud-go/sts-20150401/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/sherlockhua/koala/cache"
	"github.com/sherlockhua/koala/logs"
)

type AliFileUploadImpl struct {
	conf  *FileUploadConfig
	cache cache.RedisCache
}

func NewAliFileUploader(conf *FileUploadConfig, cache cache.RedisCache) FileUploader {
	return &AliFileUploadImpl{
		conf:  conf,
		cache: cache,
	}
}

func (u *AliFileUploadImpl) GetFileUploadAccessKey(ctx context.Context, userId int64, fileInfo *FileInfo) (*FileUploaderToken, error) {

	config := &openapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: &u.conf.AccessKeyId,
		// 必填，您的 AccessKey Secret
		AccessKeySecret: &u.conf.AccessKeySecret,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Sts
	config.Endpoint = tea.String(u.conf.StsEndpoint)
	client, err := sts.NewClient(config)
	if err != nil {
		fmt.Printf("create sts client failed, err: %v, conf:%v endpoint:%v\n", err, u.conf, u.conf.StsEndpoint)
		logs.Errorf(ctx, "create sts client failed, err: %v, conf:%v", err, u.conf)
		return nil, err
	}

	assumeRoleRequest := &sts.AssumeRoleRequest{
		DurationSeconds: tea.Int64(u.conf.ExpiredSeconds),
		RoleArn:         tea.String(u.conf.RoleArn),
		RoleSessionName: tea.String(u.conf.ServiceName),
	}
	result, err := client.AssumeRoleWithOptions(assumeRoleRequest, &util.RuntimeOptions{})
	if err != nil {
		fmt.Printf("assume role: %v, conf:%v result:%v\n", err, u.conf, result)
		logs.Errorf(ctx, "assume role: %v, conf:%v", err, u.conf)
		return nil, err
	}

	logs.Infof(ctx, "GetFileUploadAccessKey, assume role: %v, conf:%v", result, u.conf)

	return &FileUploaderToken{
		AccessKeyId:     *result.Body.Credentials.AccessKeyId,
		AccessKeySecret: *result.Body.Credentials.AccessKeySecret,
		SecurityToken:   *result.Body.Credentials.SecurityToken,
		BucketName:      u.conf.BucketName,
		Region:          u.conf.Region,
		Endpoint:        u.conf.OssEndpoint,
		ObjectKey:       u.generateObjectKey(fileInfo),
		CallbackInfo:    u.packCallbackInfo(userId, fileInfo),
	}, nil
}

func (a *AliFileUploadImpl) packCallbackInfo(userId int64, fileInfo *FileInfo) CallbackInfo {

	callbackInfo := CallbackInfo{
		CallbackUrl: a.conf.CallbackUrl,
	}
	body := CallbackBody{
		FileName:  fileInfo.FileName,
		FileSize:  fileInfo.FileSize,
		FileType:  fileInfo.FileType,
		ObjectKey: a.generateObjectKey(fileInfo),
		UserId:    fmt.Sprintf("%d", userId),
	}
	data, _ := json.Marshal(body)
	callbackInfo.CallbackBody = base64.StdEncoding.EncodeToString(data)
	return callbackInfo
}

func (a *AliFileUploadImpl) generateObjectKey(fileInfo *FileInfo) string {
	text := fmt.Sprintf("%s/%s/%d/%d", fileInfo.FileName, fileInfo.FileType, fileInfo.FileSize, time.Now().Unix())
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
