package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sherlockhua/go_example/infra/file_uploader"
)

func main() {
	r := gin.Default()

	// 提供静态文件服务 (如果需要)
	// r.Static("/static", "./static")

	// 加载 HTML 模板
	r.LoadHTMLGlob("templates/*")

	// 首页路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "OSS Web Direct Upload",
		})
	})

	// 获取 STS 凭证的 API 路由
	r.GET("/sts", getStsCredentials)

	// 启动 HTTP 服务
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}

}

// getStsCredentials handles the request for STS credentials
func getStsCredentials(c *gin.Context) {
	// 重要提示：在生产环境中，请从环境变量或安全的配置服务中加载这些信息。
	// 请勿在应用程序代码中硬编码凭证。
	uploader := file_uploader.NewAliFileUploader(&file_uploader.FileUploadConfig{
		AccessKeyId:     os.Getenv("ALICLOUD_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("ALICLOUD_ACCESS_KEY_SECRET"),
		RoleArn:         os.Getenv("ALICLOUD_ROLE_ARN"),
		BucketName:      os.Getenv("ALICLOUD_OSS_BUCKET_NAME"),
		Endpoint:        os.Getenv("ALICLOUD_OSS_ENDPOINT"),
		Region:          os.Getenv("ALICLOUD_OSS_REGION"),
		ExpiredSeconds:  3600,
		AllowBucket:     []string{os.Getenv("ALICLOUD_OSS_BUCKET_NAME")},
		ServiceName:     "oss-web-direct-upload",
	})
	result, err := uploader.GetFileUploadAccessKey(c)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取 STS 凭证失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"AccessKeyId":     result.AccessKeyId,
		"AccessKeySecret": result.AccessKeySecret,
		"SecurityToken":   result.SecurityToken,
		"BucketName":      result.BucketName,
		"Region":          result.Region, // Bucket 所在的区域，例如：oss-cn-hangzhou
		//"Endpoint":        ossEndpoint,
	})
}
