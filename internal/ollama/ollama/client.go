package ollama

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"net/url"
	"ollama-desktop/internal/config"
	"ollama-desktop/internal/ollama"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var noModelError = errors.New("No Model")

type Client struct {
	base *url.URL
	http *http.Client
}

func checkError(resp *http.Response, body []byte) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	return ollama.StatusError{StatusCode: resp.StatusCode, Status: resp.Status, ErrorMessage: string(body)}
}

func NewClient() *Client {
	base, _ := url.Parse("https://ollama.com")
	var proxy func(*http.Request) (*url.URL, error)
	if config.Config.Proxy != nil {
		proxy = http.ProxyURL(config.Config.Proxy.ToUrl())
	}
	return &Client{
		base: base,
		http: &http.Client{
			Timeout: 30 * time.Second, // 设置超时时间为 30 秒
			Transport: &http.Transport{
				Proxy: proxy,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // 不验证证书
				},
			},
		},
	}
}

func (c *Client) do(ctx context.Context, path string, reqData map[string]string) ([]byte, error) {
	requestURL := c.base.JoinPath(path)
	if len(reqData) > 0 {
		rawQuery := ""
		for name, value := range reqData {
			if rawQuery != "" {
				rawQuery += "&"
			}
			rawQuery += name + "=" + url.QueryEscape(value)
		}
		requestURL.RawQuery = rawQuery
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", fmt.Sprintf("ollama-desktop/%s (%s %s) Go/%s", config.BuildVersion, runtime.GOARCH, runtime.GOOS, runtime.Version()))

	respObj, err := c.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer respObj.Body.Close()

	respBody, err := io.ReadAll(respObj.Body)
	if err != nil {
		return nil, err
	}

	if err := checkError(respObj, respBody); err != nil {
		return nil, err
	}
	return respBody, nil
}

// 解析模型名称、是否归档
func parseNameArchive(s *goquery.Selection) (string, bool) {
	children := s.Children()
	name := strings.TrimSpace(children.Eq(0).Text())
	if children.Size() < 2 {
		return name, false
	}
	// 存在名称后面存在标签的情况，比如falcon
	for i := 1; i < children.Size(); i++ {
		if strings.Contains(children.Eq(i).Text(), "Archive") {
			return name, true
		}
	}
	return name, false
}

// 解析下载次数、标签数、更新时间
func parsePullTagCountAndUpdated(spans *goquery.Selection) (string, int, string) {
	pullCount := ""
	tagCount := 0
	updated := ""
	spans.Each(func(_ int, info *goquery.Selection) {
		text := strings.ReplaceAll(strings.TrimSpace(info.Text()), "\t", " ")
		if strings.HasSuffix(text, "Pulls") {
			pullCount = strings.TrimSpace(text[:len(text)-5])
		} else if strings.HasSuffix(text, "Pull") {
			pullCount = strings.TrimSpace(text[:len(text)-4])
		} else if strings.HasSuffix(text, "Tag") {
			tagCount, _ = strconv.Atoi(strings.TrimSpace(text[:len(text)-3]))
		} else if strings.HasSuffix(text, "Tags") {
			tagCount, _ = strconv.Atoi(strings.TrimSpace(text[:len(text)-4]))
		} else if strings.HasPrefix(text, "Updated") {
			updated = strings.TrimSpace(text[7:])
		}
	})
	return pullCount, tagCount, updated
}

func parseTags(spans *goquery.Selection) []string {
	var tags []string
	spans.Each(func(_ int, tag *goquery.Selection) {
		tags = append(tags, strings.TrimSpace(tag.Text()))
	})
	return tags
}

func (c *Client) SearchPreview(ctx context.Context, q string) ([]*ollama.SimpleModelInfo, error) {
	respBody, err := c.do(ctx, "/search-preview", map[string]string{"q": q})
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return nil, err
	}
	var list []*ollama.SimpleModelInfo
	doc.Find("ul > li > a").Each(func(i int, item *goquery.Selection) {
		children := item.Children()
		name, archive := parseNameArchive(children.First())
		list = append(list, &ollama.SimpleModelInfo{
			Name:        name,
			Archive:     archive,
			Description: strings.TrimSpace(children.Last().Text()),
		})
	})
	return list, nil
}

func (c *Client) Search(ctx context.Context, request *ollama.SearchRequest) (*ollama.SearchResponse, error) {
	if request.P <= 0 {
		request.P = 1
	}
	respBody, err := c.do(ctx, "/search", map[string]string{
		"q": request.Q,
		"p": strconv.Itoa(request.P),
		"c": request.C,
	})
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return nil, err
	}
	var list []*ollama.ModelInfo

	// 解析模型信息
	doc.Find("ul.grid > li > a").Each(func(_ int, item *goquery.Selection) {
		name, archive := parseNameArchive(item.Find("h2.flex.items-center").First())
		description := strings.TrimSpace(item.Find("div.space-y-2 > p.break-words").First().Text())
		tags := parseTags(item.Find("div.space-y-2 > div.space-x-2 > span"))
		pullCount, tagCount, updated := parsePullTagCountAndUpdated(item.Find("div.space-y-2 > p.space-x-5 > span"))
		list = append(list, &ollama.ModelInfo{
			Name:        name,
			Archive:     archive,
			Description: description,
			Tags:        tags,
			PullCount:   pullCount,
			TagCount:    tagCount,
			UpdateTime:  updated,
		})
	})
	pageCount := 0
	// 解析导航页码
	doc.Find("nav > ul > li > a").Each(func(_ int, item *goquery.Selection) {
		text := strings.TrimSpace(item.Text())
		if "Previous" == text || "Next" == text {
			return
		}
		if page, err := strconv.Atoi(text); err == nil {
			if page > pageCount {
				pageCount = page
			}
		}
	})

	return &ollama.SearchResponse{
		Query:     request.Q,
		Page:      request.P,
		PageCount: pageCount,
		Items:     list,
	}, nil
}

func (c *Client) Library(ctx context.Context, request *ollama.LibraryRequest) ([]*ollama.ModelInfo, error) {
	if request.Sort == "" {
		request.Sort = "featured"
	}
	respBody, err := c.do(ctx, "/library", map[string]string{
		"q":    request.Q,
		"sort": request.Sort,
	})
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return nil, err
	}
	var list []*ollama.ModelInfo

	// 解析模型信息
	doc.Find("ul.grid > li > a").Each(func(_ int, item *goquery.Selection) {
		name, archive := parseNameArchive(item.Find("div.flex.items-center.mb-3").First())
		description := strings.TrimSpace(item.Find("div.space-y-2 > p.break-words").First().Text())
		tags := parseTags(item.Find("div.space-y-2 > div.space-x-2 > span"))
		pullCount, tagCount, updated := parsePullTagCountAndUpdated(item.Find("div.space-y-2 > p.space-x-5 > span"))
		list = append(list, &ollama.ModelInfo{
			Name:        name,
			Archive:     archive,
			Description: description,
			Tags:        tags,
			PullCount:   pullCount,
			TagCount:    tagCount,
			UpdateTime:  updated,
		})
	})

	return list, err
}

func (c *Client) ModelTags(ctx context.Context, model string) (*ollama.ModelTagsResponse, error) {
	respBody, err := c.do(ctx, fmt.Sprintf("/library/%s/tags", model), nil)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return nil, err
	}

	name, archive := parseNameArchive(doc.Find("div > main > section > div > h1").First())
	description := strings.TrimSpace(doc.Find("div > main > section > h2.break-words").First().Text())
	tags := parseTags(doc.Find("div > main > section > div.space-x-2 > span"))
	pullCount, tagCount, updated := parsePullTagCountAndUpdated(doc.Find("div > main > section > p.space-x-5 > span"))
	modelInfo := &ollama.ModelInfo{
		Name:        name,
		Archive:     archive,
		Description: description,
		Tags:        tags,
		PullCount:   pullCount,
		TagCount:    tagCount,
		UpdateTime:  updated,
	}

	response := &ollama.ModelTagsResponse{
		Model: modelInfo,
	}

	tagsNode := doc.Find("section > div > div > div.px-4.py-3 > div")

	if tagsNode.Size() < 1 {
		modelInfo.TagCount = 0
		return response, nil
	}

	var modelTags []*ollama.ModelTag
	for i := 0; i < tagsNode.Size(); i++ {
		tagNode := tagsNode.Eq(i)

		line1Node := tagNode.Find("div.space-x-2").First()
		name := strings.TrimSpace(line1Node.Find("a.group").Text())
		latest := strings.Contains(line1Node.Find("span.px-2").Text(), "latest")

		infos := strings.Split(tagNode.Find("div.space-x-1 > span").Text(), "•")
		tag := &ollama.ModelTag{
			Name:       name,
			Latest:     latest,
			Id:         strings.TrimSpace(infos[0]),
			Size:       strings.TrimSpace(infos[1]),
			UpdateTime: strings.TrimSpace(strings.TrimSpace(infos[2])[7:]),
		}
		modelTags = append(modelTags, tag)
	}
	response.Tags = modelTags
	modelInfo.TagCount = len(modelTags)

	return response, err
}

func (c *Client) ModelInfo(ctx context.Context, model, tag string) (*ollama.ModelInfoResponse, error) {
	if tag != "" {
		tag = ":" + tag
	}
	respBody, err := c.do(ctx, fmt.Sprintf("/library/%s%s", model, tag), nil)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return nil, err
	}

	name, archive := parseNameArchive(doc.Find("div > main > section > div > h1").First())
	description := strings.TrimSpace(doc.Find("div > main > section > h2.break-words").First().Text())
	tags := parseTags(doc.Find("div > main > section > div.space-x-2 > span"))
	pullCount, tagCount, updated := parsePullTagCountAndUpdated(doc.Find("div > main > section > p.space-x-5 > span"))
	modelInfo := &ollama.ModelInfo{
		Name:        name,
		Archive:     archive,
		Description: description,
		Tags:        tags,
		PullCount:   pullCount,
		TagCount:    tagCount,
		UpdateTime:  updated,
	}

	var modelTags []*ollama.ModelTag
	tagFunc := func(_ int, item *goquery.Selection) {
		leftNode := item.Find("div.flex.space-x-2").First()
		name := strings.TrimSpace(leftNode.Find("span.truncate").Text())
		latest := strings.Contains(leftNode.Find("span.px-2").Text(), "latest")

		item.Find("div.space-x-2")
		size := strings.TrimSpace(item.Find("span.text-neutral-400").First().Text())
		tag := &ollama.ModelTag{
			Name:       name,
			Latest:     latest,
			Id:         "",
			Size:       size,
			UpdateTime: "",
		}
		modelTags = append(modelTags, tag)
	}

	doc.Find("#primary-tags > a").Each(tagFunc)
	doc.Find("#secondary-tags > a").Each(tagFunc)
	modelInfo.TagCount = len(modelTags)

	readme := doc.Find("div#textareaInput > textarea#editor").Eq(0).Text()

	response := &ollama.ModelInfoResponse{
		Model:  modelInfo,
		Tags:   modelTags,
		Readme: readme,
	}
	return response, err
}
