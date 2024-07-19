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
func parsePullTagCountAndUpdated(content *goquery.Selection) (string, int, string) {
	pullCount := ""
	tagCount := 0
	updated := ""
	content.Find("p.my-2 > span").Each(func(_ int, info *goquery.Selection) {
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
		name, archive := parseNameArchive(children.First())
		list = append(list, &ollama.SimpleModelInfo{
			Name:        name,
			Archive:     archive,
			Description: strings.TrimSpace(children.Last().Text()),
		})
	})
	return list, nil
}

func (c *Client) Search(ctx context.Context, q string, p int, cParam string) (*ollama.SearchResponse, error) {
	if p <= 0 {
		p = 1
	}
	respBody, err := c.do(ctx, "/search", map[string]string{
		"q": q,
		"p": strconv.Itoa(p),
		//embedding
		//vision
		//tools
		"c": cParam,
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
		name, archive := parseNameArchive(children.First())
		// 部分数据不一定存在
		content := children.Last()
		description := strings.TrimSpace(content.Find("p.break-words").Eq(0).Text())
		var tags []string
		content.Find("div.space-x-2 > span").Each(func(_ int, tag *goquery.Selection) {
			tags = append(tags, strings.TrimSpace(tag.Text()))
		})
		pullCount, tagCount, updated := parsePullTagCountAndUpdated(content)
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
		name, archive := parseNameArchive(children.First())

		// 部分数据不一定存在
		content := children.Last()
		description := strings.TrimSpace(content.Find("p.break-words").Eq(0).Text())
		var tags []string
		content.Find("div.space-x-2 > span").Each(func(_ int, tag *goquery.Selection) {
			tags = append(tags, strings.TrimSpace(tag.Text()))
		})
		pullCount, tagCount, updated := parsePullTagCountAndUpdated(content)
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

	sections := doc.Find("div > main > section")
	// 解析模型信息
	if sections.Size() != 2 {
		return nil, noModelError
	}

	node := sections.First().Children()
	name, archive := parseNameArchive(node.Eq(0).Children().First())
	modelInfo := &ollama.ModelInfo{
		Name:        name,
		Archive:     archive,
		Description: strings.TrimSpace(node.Eq(1).Text()),
	}

	var tags []string
	node.Eq(2).Find("span").Each(func(_ int, tag *goquery.Selection) {
		tags = append(tags, strings.TrimSpace(tag.Text()))
	})
	modelInfo.Tags = tags

	node = node.Eq(3).Children()
	text := strings.TrimSpace(node.Eq(0).Text())
	modelInfo.PullCount = strings.TrimSpace(text[:len(text)-5])
	text = strings.TrimSpace(node.Eq(1).Text())
	modelInfo.UpdateTime = strings.TrimSpace(text[7:])

	response := &ollama.ModelTagsResponse{
		Model: modelInfo,
	}

	node = sections.Last().Find("div.px-4.py-3 > div")

	if node.Size() < 1 {
		modelInfo.TagCount = 0
		return response, nil
	}

	var modelTags []*ollama.ModelTag
	for i := 0; i < node.Size(); i++ {
		tagNode := node.Eq(i)
		children := tagNode.Children()
		first := children.First()
		last := children.Last()

		children = first.Children()

		infos := strings.Split(last.Text(), "•")
		tag := &ollama.ModelTag{
			Name:       strings.TrimSpace(children.First().Text()),
			Latest:     children.Size() > 1 && strings.TrimSpace(children.Eq(1).Text()) == "latest",
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

func (c *Client) ModelInfo(ctx context.Context, model string) (*ollama.ModelInfoResponse, error) {
	respBody, err := c.do(ctx, fmt.Sprintf("/library/%s", model), nil)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
	if err != nil {
		return nil, err
	}

	sections := doc.Find("div > main > section")
	// 解析模型信息
	if sections.Size() != 2 {
		return nil, noModelError
	}

	node := sections.First().Children()
	name, archive := parseNameArchive(node.Eq(0).Children().First())
	modelInfo := &ollama.ModelInfo{
		Name:        name,
		Archive:     archive,
		Description: strings.TrimSpace(sections.First().Find("h2.break-words").Text()),
	}

	var tags []string
	sections.First().Find("div.space-x-2 > span").Each(func(_ int, tag *goquery.Selection) {
		tags = append(tags, strings.TrimSpace(tag.Text()))
	})
	modelInfo.Tags = tags

	sections.First().Find("p.space-x-5 > span").Each(func(_ int, item *goquery.Selection) {
		text := strings.TrimSpace(item.Text())
		if strings.HasSuffix(text, "Pulls") {
			modelInfo.PullCount = strings.TrimSpace(text[:len(text)-5])
		} else if strings.HasPrefix(text, "Updated") {
			modelInfo.UpdateTime = strings.TrimSpace(text[7:])
		}
	})

	response := &ollama.ModelInfoResponse{
		Model:  modelInfo,
		Readme: doc.Find("div#textareaInput > textarea#editor").Eq(0).Text(),
	}
	return response, err
}
