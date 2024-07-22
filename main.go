package main

import (
	"embed"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"io"
	"net/http"
	"ollama-desktop/internal/app"
	"ollama-desktop/internal/config"
	"ollama-desktop/internal/log"
	"runtime"
	"time"
)

//go:embed all:frontend/dist
var assets embed.FS

// 初始化时区
func init() {
	timeZone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Printf("Load time zone error: %s\n", err.Error())
		return
	}
	time.Local = timeZone
}

func main() {
	log.Info().Str("BuildTime", config.BuildTime).Str("BuildVersion", config.BuildVersion).
		Str("Arch", runtime.GOARCH).Str("OS", runtime.GOOS).Str("GoVersion", runtime.Version()).
		Msg("Ollama Desktop")

	// 创建一个HTTP处理器函数
	handler := func(w http.ResponseWriter, r *http.Request) {
		targetHost := r.Header.Get("Target-Host")
		if "ollama" == targetHost {
			targetHost = config.Config.Ollama.Host.String()
		}
		if targetHost == "" {
			http.Error(w, "目标主机不存在", http.StatusBadGateway)
			return
		}
		// 克隆原始请求
		proxyReq, err := http.NewRequest(r.Method, r.Header.Get("Target-Host")+r.RequestURI, r.Body)
		if err != nil {
			http.Error(w, "创建请求出错", http.StatusInternalServerError)
			return
		}
		proxyReq.Header = r.Header

		// 使用默认的HTTP客户端发送请求
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "发送请求到目标服务出错", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 将响应头复制到客户端
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		// 将响应状态码和响应体复制到客户端
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	err := app.StartApp(&assetserver.Options{
		Assets:     assets,
		Handler:    mux,
		Middleware: nil,
	})

	if err != nil {
		log.Error().Err(err).Msg("Run Ollama Desktop Error")
	}
}
