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

func NewClient() (*Client, error) {
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
	}, nil
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
	doc.Find("ul li a").Each(func(i int, item *goquery.Selection) {
		children := item.Children()
		list = append(list, &ollama.SimpleModelInfo{
			Name:        strings.TrimSpace(children.First().Text()),
			Description: strings.TrimSpace(children.Last().Text()),
		})
	})
	return list, nil
}

func (c *Client) Search(ctx context.Context, q string, p int) (*ollama.SearchResponse, error) {
	if p <= 0 {
		p = 1
	}
	respBody, err := c.do(ctx, "/search", map[string]string{
		"q": q,
		"p": strconv.Itoa(p),
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
	doc.Find("ul.grid li a").Each(func(_ int, item *goquery.Selection) {
		children := item.Children()
		name := strings.TrimSpace(children.First().Text())
		// 部分数据不一定存在
		content := children.Last()
		description := strings.TrimSpace(content.Find("p.break-words").Eq(0).Text())
		var tags []string
		content.Find("div.space-x-2 > span").Each(func(_ int, tag *goquery.Selection) {
			tags = append(tags, strings.TrimSpace(tag.Text()))
		})
		pullCount := ""
		tagCount := 0
		updated := ""
		content.Find("p.my-2 > span").Each(func(_ int, info *goquery.Selection) {
			text := strings.ReplaceAll(strings.TrimSpace(info.Text()), "\t", " ")
			if strings.HasSuffix(text, "Pulls") || strings.HasSuffix(text, "Pull") {
				pullCount = strings.TrimSpace(text[0 : len(text)-5])
			} else if strings.HasSuffix(text, "Tag") || strings.HasSuffix(text, "Tags") {
				tagCount, _ = strconv.Atoi(strings.TrimSpace(text[0 : len(text)-4]))
			} else if strings.HasPrefix(text, "Updated") {
				updated = strings.TrimSpace(text[7:])
			}
		})
		list = append(list, &ollama.ModelInfo{
			Name:        name,
			Description: description,
			Tags:        tags,
			PullCount:   pullCount,
			TagCount:    tagCount,
			UpdateTime:  updated,
		})
	})
	pageCount := 0
	// 解析导航页码
	doc.Find("nav ul li a").Each(func(_ int, item *goquery.Selection) {
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
		Query:     q,
		Page:      p,
		PageCount: pageCount,
		Items:     list,
	}, nil
}

func (c *Client) Library(ctx context.Context, q, sort string) ([]*ollama.ModelInfo, error) {
	if sort == "" {
		sort = "featured"
	}
	respBody, err := c.do(ctx, "/library", map[string]string{
		"q":    q,
		"sort": sort,
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
	doc.Find("ul.grid li a").Each(func(_ int, item *goquery.Selection) {
		children := item.Children()
		name := strings.TrimSpace(children.First().Text())
		// 部分数据不一定存在
		content := children.Last()
		description := strings.TrimSpace(content.Find("p.break-words").Eq(0).Text())
		var tags []string
		content.Find("div.space-x-2 > span").Each(func(_ int, tag *goquery.Selection) {
			tags = append(tags, strings.TrimSpace(tag.Text()))
		})
		pullCount := ""
		tagCount := 0
		updated := ""
		content.Find("p.my-2 > span").Each(func(_ int, info *goquery.Selection) {
			text := strings.ReplaceAll(strings.TrimSpace(info.Text()), "\t", " ")
			if strings.HasSuffix(text, "Pulls") || strings.HasSuffix(text, "Pull") {
				pullCount = strings.TrimSpace(text[0 : len(text)-5])
			} else if strings.HasSuffix(text, "Tag") || strings.HasSuffix(text, "Tags") {
				tagCount, _ = strconv.Atoi(strings.TrimSpace(text[0 : len(text)-4]))
			} else if strings.HasPrefix(text, "Updated") {
				updated = strings.TrimSpace(text[7:])
			}
		})
		list = append(list, &ollama.ModelInfo{
			Name:        name,
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

	sections := doc.Find("div main section")
	// 解析模型信息
	if sections.Size() != 2 {
		return nil, errors.New("No Model")
	}

	node := sections.First().Children()

	modelInfo := &ollama.ModelInfo{
		Name:        strings.TrimSpace(node.Eq(0).Text()),
		Description: strings.TrimSpace(node.Eq(1).Text()),
	}

	var tags []string
	node.Eq(2).Find("span").Each(func(_ int, tag *goquery.Selection) {
		tags = append(tags, strings.TrimSpace(tag.Text()))
	})
	modelInfo.Tags = tags

	node = node.Eq(3).Children()
	text := strings.TrimSpace(node.Eq(0).Text())
	modelInfo.PullCount = strings.TrimSpace(text[0 : len(text)-5])
	text = strings.TrimSpace(node.Eq(1).Text())
	modelInfo.UpdateTime = strings.TrimSpace(text[7:])

	response := &ollama.ModelTagsResponse{
		Model: modelInfo,
	}

	node = sections.Last().Find("div.px-4.py-3")

	if node.Size() <= 1 {
		modelInfo.TagCount = 0
		return response, nil
	}
	var modelTags []*ollama.ModelTag
	for i := 1; i < node.Size(); i++ {
		tagNode := node.Eq(i)
		println(tagNode.Text())
	}
	response.Tags = modelTags

	return response, err
}
