package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	ollama2 "ollama-desktop/internal/ollama"
	"ollama-desktop/internal/ollama/api"
	"sync"
)

var downloader = DownLoader{}

const (
	pullWait    = 1
	pulling     = 2
	pullSuccess = 3
	pullError   = -1
)

type DownloadItem struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	// 进度条名称
	Names []string `json:"names"`
	// 进度条数据
	Bars   map[string]ollama2.ProgressResponse `json:"bars"`
	cancel context.CancelFunc                  `json:"-"`
}

type DownLoader struct {
	tasks map[string]*DownloadItem
	lock  sync.Mutex
}

func (d *DownLoader) Pull(requestStr string) error {
	request := &ollama2.PullRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	if request.Model == "" && request.Name != "" {
		request.Model = request.Name
	}
	if request.Model == "" {
		return errors.New("model must not be empty")
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.tasks == nil {
		d.tasks = make(map[string]*DownloadItem)
	}
	if _, ok := d.tasks[request.Model]; ok {
		return nil
	}
	item := &DownloadItem{
		Model:    request.Model,
		Insecure: request.Insecure,
		Names:    nil,
		Bars:     make(map[string]ollama2.ProgressResponse),
	}
	d.tasks[request.Model] = item
	go d.pull(request, item)
	return nil
}

func (d *DownLoader) pull(request *ollama2.PullRequest, item *DownloadItem) {
	ctx, cancel := context.WithCancel(app.ctx)
	item.cancel = cancel
	d.eventsEmit(request.Model, pullWait, item)
	err := api.ClientFromConfig().Pull(ctx, request, func(response ollama2.ProgressResponse) error {
		d.lock.Lock()
		defer d.lock.Unlock()
		status := ""
		if response.Digest != "" {
			if _, ok := item.Bars[response.Digest]; !ok {
				item.Names = append(item.Names, response.Digest)
			}
			item.Bars[response.Digest] = response
		} else if status != response.Status {
			status = response.Status
			item.Names = append(item.Names, status)
		}
		d.eventsEmit(request.Model, pulling, item)
		return nil
	})
	if err != nil {
		d.eventsEmit(request.Model, pullError, item)
	} else {
		d.eventsEmit(request.Model, pullSuccess, item)
	}
	delete(d.tasks, request.Model)
}

func (d *DownLoader) eventsEmit(model string, status int, item *DownloadItem) {
	d.lock.Lock()
	defer d.lock.Unlock()
	runtime.EventsEmit(app.ctx, "pull_"+model, status, item)
}

func (d *DownLoader) Cancel(model string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.tasks == nil {
		return
	}
	if item, ok := d.tasks[model]; ok {
		if item.cancel != nil {
			item.cancel()
		}
	}
	return
}

func (d *DownLoader) List() []*DownloadItem {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.tasks == nil {
		return nil
	}
	var list []*DownloadItem
	for _, item := range d.tasks {
		list = append(list, item)
	}
	return list
}
