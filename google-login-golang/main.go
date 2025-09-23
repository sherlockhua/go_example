package main

import (
	"log"
	"net/http"
	"os"

	//	"sync"
	"fmt"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 未找到.env文件，将使用系统环境变量: %v", err)
	}

	// 初始化日志
	logger := log.New(os.Stdout, "GOOGLE_AUTH: ", log.Ldate|log.Ltime|log.Lshortfile)

	// 从环境变量获取配置
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		logger.Fatal("请设置GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET和GOOGLE_REDIRECT_URL环境变量")
	}

	// 定义需要获取的用户权限范围
	scopes := []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
		"openid",
	}

	// 创建GoogleAuth实例
	googleAuth, err := NewGoogleAuth(clientID, clientSecret, redirectURL, scopes, logger)
	if err != nil {
		logger.Fatalf("初始化GoogleAuth失败: %v", err)
	}

	// 设置HTTP路由
	mux := http.NewServeMux()

	// 首页路由
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
			<head>
				<title>Google登录示例</title>
				<style>
					body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
					.login-btn {
						background-color: #4285F4;
						color: white;
						padding: 10px 20px;
						border: none;
						border-radius: 4px;
						font-size: 16px;
						cursor: pointer;
						text-decoration: none;
					}
					.login-btn:hover { background-color: #3367d6; }
				</style>
			</head>
			<body>
				<h1>Google登录示例</h1>
				<a href="/login" class="login-btn">使用Google账号登录</a>
			</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	// 登录路由
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		authURL, err := googleAuth.GenerateAuthURL()
		if err != nil {
			logger.Printf("生成授权URL失败: %v", err)
			http.Error(w, "生成登录链接失败", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, authURL, http.StatusFound)
	})

	// 回调路由
	mux.HandleFunc("/api/v1/google/callback", func(w http.ResponseWriter, r *http.Request) {
		// 获取回调参数
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" || state == "" {
			http.Error(w, "缺少必要的参数", http.StatusBadRequest)
			return
		}

		// 处理登录回调
		user, err := googleAuth.HandleCallback(code, state)
		if err != nil {
			logger.Printf("处理回调失败: %v", err)
			http.Error(w, "登录失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 显示用户信息
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<h1>登录成功！</h1>"))
		w.Write([]byte(fmt.Sprintf("<p>ID: %s</p>", user.ID)))
		w.Write([]byte(fmt.Sprintf("<p>姓名: %s</p>", user.Name)))
		w.Write([]byte(fmt.Sprintf("<p>邮箱: %s</p>", user.Email)))
		if user.Picture != "" {
			w.Write([]byte(fmt.Sprintf("<p><img src=\"%s\" width=\"100\" height=\"100\"></p>", user.Picture)))
		}
	})

	// 健康检查路由
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 配置HTTP服务器
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	logger.Println("服务器启动在 http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("服务器启动失败: %v", err)
	}
}
