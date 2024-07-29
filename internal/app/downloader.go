package app

import (
	"encoding/json"
	ollama2 "ollama-desktop/internal/ollama"
	"sync"
)

var downloader = DownLoader{
	ch: make(chan bool, 3),
}

type DownloadItem struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	// 进度条名称
	Names []string `json:"names"`
	// 进度条数据
	Bars map[string]ollama2.ProgressResponse `json:"bars"`
}

type DownLoader struct {
	ch   chan bool
	lock sync.Mutex
}

func (d *DownLoader) Add(requestStr string) error {
	request := &ollama2.PullRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	// 限制协程数量
	d.ch <- true
	go func() {
		defer func() {
			<-d.ch
		}()
	}()
	return nil
}
