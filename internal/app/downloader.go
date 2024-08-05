package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	ollama2 "ollama-desktop/internal/ollama"
	"sort"
	"sync"
)

var downloader = DownLoader{}

const (
	pullStatusWait    = 1
	pullStatusPulling = 2
	pullStatusSuccess = 3
	pullStatusError   = -1
	pullEventList     = "pull_list"
	pullEventSuccess  = "pull_success"
	pullEventError    = "pull_error"
)

type DownloadItem struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	// 进度条数据
	Bars   []*ollama2.ProgressResponse `json:"bars"`
	cancel context.CancelFunc          `json:"-"`
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
	if d.tasks == nil {
		d.tasks = make(map[string]*DownloadItem)
	}
	if _, ok := d.tasks[request.Model]; ok {
		return nil
	}
	item := &DownloadItem{
		Model:    request.Model,
		Insecure: request.Insecure,
		Bars:     nil,
	}
	d.tasks[request.Model] = item
	go d.pull(request, item)
	return nil
}

func (d *DownLoader) pull(request *ollama2.PullRequest, item *DownloadItem) {
	ctx, cancel := context.WithCancel(app.ctx)
	item.cancel = cancel
	d.emit(pullStatusWait, item)
	err := ollama.newApiClient().Pull(ctx, request, func(response ollama2.ProgressResponse) error {
		status := ""
		length := len(item.Bars)
		if response.Digest != "" {
			if length == 0 {
				item.Bars = append(item.Bars, &response)
			} else if item.Bars[length-1].Digest != response.Digest {
				item.Bars = append(item.Bars, &response)
			} else {
				item.Bars[length-1] = &response
			}
		} else if status != response.Status {
			status = response.Status
			item.Bars = append(item.Bars, &response)
		}
		d.emit(pullStatusPulling, item)
		return nil
	})
	if err != nil {
		d.emit(pullStatusError, item)
	} else {
		d.emit(pullStatusSuccess, item)
	}
	delete(d.tasks, request.Model)
}

func (d *DownLoader) emit(status int, item *DownloadItem) {
	d.lock.Lock()
	defer d.lock.Unlock()
	runtime.EventsEmit(app.ctx, pullEventList, d.List())
	switch status {
	case pullStatusSuccess:
		runtime.EventsEmit(app.ctx, pullEventSuccess, item)
	case pullStatusError:
		runtime.EventsEmit(app.ctx, pullEventError, item)
	}
}

func (d *DownLoader) Cancel(model string) {
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
	if d.tasks == nil {
		return nil
	}
	var list []*DownloadItem
	for _, item := range d.tasks {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Model < list[j].Model
	})
	return list
}
