package app

import (
	"ollama-desktop/internal/ollama"
	"sync"
)

var downloader = DownLoader{}

type DownloadItem struct {
	Pull ollama.PullRequest
}

type DownLoader struct {
	ch      chan bool
	counter uint
	lock    sync.Mutex
}

func (d *DownLoader) Add() {

}

func GetDownLoader() *DownLoader {
	_once.Do(func() {
		downloader = &DownLoader{ch: make(chan bool)}
	})
	return downloader
}
