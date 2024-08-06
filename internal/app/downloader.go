package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	ollama2 "ollama-desktop/internal/ollama"
	"sort"
	"sync"
	"time"
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
	eventModelRefresh = "model_refresh"
)

type DownloadItem struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	// 进度条数据
	Bars     []*ProgressBar     `json:"bars"`
	Canceled bool               `json:"-"`
	cancel   context.CancelFunc `json:"-"`
}

type ProgressBar struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
	Status     string  `json:"status"`
}

func (p *ProgressBar) stop() {
	p.Status = "success"
	p.Percentage = 100
}

func (p *ProgressBar) set(percentage float64) {
	p.Percentage = percentage
	if percentage >= 100 {
		p.Status = "success"
	} else {
		p.Status = ""
	}
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

	cache := make(map[string]*ProgressBar)
	var status string
	var spinner *ProgressBar
	err := ollama.newApiClient().Pull(ctx, request, func(resp ollama2.ProgressResponse) error {
		if resp.Digest != "" {
			if spinner != nil {
				spinner.stop()
			}
			bar, ok := cache[resp.Digest]
			if !ok {
				bar = &ProgressBar{
					Name:       fmt.Sprintf("pulling %s...", resp.Digest[7:19]),
					Percentage: 0,
					Status:     "",
				}
				item.Bars = append(item.Bars, bar)
				cache[resp.Digest] = bar
			}
			percentage := float64(resp.Completed) / float64(resp.Total) * 100
			bar.set(percentage)
		} else if status != resp.Status {
			if spinner != nil {
				spinner.stop()
			}

			status = resp.Status

			spinner = &ProgressBar{
				Name:       status,
				Percentage: 0,
				Status:     "",
			}
			item.Bars = append(item.Bars, spinner)
		}
		d.emit(pullStatusPulling, item)
		return nil
	})

	if err != nil {
		delete(d.tasks, request.Model)
		if !item.Canceled {
			d.emit(pullStatusError, item)
		} else {
			d.emit(pullStatusPulling, item)
		}
	} else {
		if spinner != nil {
			spinner.stop()
		}
		d.emit(pullStatusPulling, item)

		<-time.After(2 * time.Second)

		delete(d.tasks, request.Model)
		if !item.Canceled {
			d.emit(pullStatusSuccess, item)
		} else {
			d.emit(pullStatusPulling, item)
		}
	}
}

func (d *DownLoader) emit(status int, item *DownloadItem) {
	d.lock.Lock()
	defer d.lock.Unlock()
	runtime.EventsEmit(app.ctx, pullEventList, d.List())
	switch status {
	case pullStatusSuccess:
		runtime.EventsEmit(app.ctx, pullEventSuccess, item)
		runtime.EventsEmit(app.ctx, eventModelRefresh)
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
			item.Canceled = true
			item.cancel()
		}
	}
	runtime.EventsEmit(app.ctx, pullEventList, d.List())
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
