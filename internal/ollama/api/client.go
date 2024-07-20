package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"ollama-desktop/internal/config"
	"ollama-desktop/internal/ollama"
	"ollama-desktop/internal/ollama/format"
	"runtime"
)

// Client encapsulates client state for interacting with the ollama
// service. Use [ClientFromEnvironment] to create new Clients.
type Client struct {
	base *url.URL
	http *http.Client
}

func checkError(resp *http.Response, body []byte) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	apiError := ollama.StatusError{StatusCode: resp.StatusCode}

	err := json.Unmarshal(body, &apiError)
	if err != nil {
		// Use the full body as the message if we fail to decode a response.
		apiError.ErrorMessage = string(body)
	}

	return apiError
}

// ClientFromEnvironment creates a new [Client] using configuration from the
// environment variable OLLAMA_HOST, which points to the network host and
// port on which the ollama service is listenting. The format of this variable
// is:
//
//	<scheme>://<host>:<port>
//
// If the variable is not specified, a default ollama host and port will be
// used.
func ClientFromConfig() *Client {
	ollamaHost := config.Config.Ollama.Host

	return &Client{
		base: &url.URL{
			Scheme: ollamaHost.Scheme,
			Host:   net.JoinHostPort(ollamaHost.Host, ollamaHost.Port),
		},
		http: http.DefaultClient,
	}
}

func (c *Client) do(ctx context.Context, method, path string, reqData, respData any) error {
	var reqBody io.Reader
	var data []byte
	var err error

	switch reqData := reqData.(type) {
	case io.Reader:
		// reqData is already an io.Reader
		reqBody = reqData
	case nil:
		// noop
	default:
		data, err = json.Marshal(reqData)
		if err != nil {
			return err
		}

		reqBody = bytes.NewReader(data)
	}

	requestURL := c.base.JoinPath(path)
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), reqBody)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", fmt.Sprintf("ollama-desktop/%s (%s %s) Go/%s", config.BuildVersion, runtime.GOARCH, runtime.GOOS, runtime.Version()))

	respObj, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer respObj.Body.Close()

	respBody, err := io.ReadAll(respObj.Body)
	if err != nil {
		return err
	}

	if err := checkError(respObj, respBody); err != nil {
		return err
	}

	if len(respBody) > 0 && respData != nil {
		if err := json.Unmarshal(respBody, respData); err != nil {
			return err
		}
	}
	return nil
}

const maxBufferSize = 512 * format.KiloByte

func (c *Client) stream(ctx context.Context, method, path string, data any, fn func([]byte) error) error {
	var buf *bytes.Buffer
	if data != nil {
		bts, err := json.Marshal(data)
		if err != nil {
			return err
		}

		buf = bytes.NewBuffer(bts)
	}

	requestURL := c.base.JoinPath(path)
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), buf)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/x-ndjson")
	request.Header.Set("User-Agent", fmt.Sprintf("ollama-desktop/%s (%s %s) Go/%s", config.BuildVersion, runtime.GOARCH, runtime.GOOS, runtime.Version()))

	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	scanner := bufio.NewScanner(response.Body)
	// increase the buffer size to avoid running out of space
	scanBuf := make([]byte, 0, maxBufferSize)
	scanner.Buffer(scanBuf, maxBufferSize)
	for scanner.Scan() {
		var errorResponse struct {
			Error string `json:"error,omitempty"`
		}

		bts := scanner.Bytes()
		if err := json.Unmarshal(bts, &errorResponse); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		if errorResponse.Error != "" {
			return fmt.Errorf(errorResponse.Error)
		}

		if response.StatusCode >= http.StatusBadRequest {
			return ollama.StatusError{
				StatusCode:   response.StatusCode,
				Status:       response.Status,
				ErrorMessage: errorResponse.Error,
			}
		}

		if err := fn(bts); err != nil {
			return err
		}
	}

	return nil
}

// GenerateResponseFunc is a function that [Client.Generate] invokes every time
// a response is received from the service. If this function returns an error,
// [Client.Generate] will stop generating and return this error.
type GenerateResponseFunc func(ollama.GenerateResponse) error

// Generate generates a response for a given prompt. The req parameter should
// be populated with prompt details. fn is called for each response (there may
// be multiple responses, e.g. in case streaming is enabled).
func (c *Client) Generate(ctx context.Context, req *ollama.GenerateRequest, fn GenerateResponseFunc) error {
	return c.stream(ctx, http.MethodPost, "/api/generate", req, func(bts []byte) error {
		var resp ollama.GenerateResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}

		return fn(resp)
	})
}

// ChatResponseFunc is a function that [Client.Chat] invokes every time
// a response is received from the service. If this function returns an error,
// [Client.Chat] will stop generating and return this error.
type ChatResponseFunc func(ollama.ChatResponse) error

// Chat generates the next message in a chat. [ChatRequest] may contain a
// sequence of messages which can be used to maintain chat history with a model.
// fn is called for each response (there may be multiple responses, e.g. if case
// streaming is enabled).
func (c *Client) Chat(ctx context.Context, req *ollama.ChatRequest, fn ChatResponseFunc) error {
	return c.stream(ctx, http.MethodPost, "/api/chat", req, func(bts []byte) error {
		var resp ollama.ChatResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}

		return fn(resp)
	})
}

// PullProgressFunc is a function that [Client.Pull] invokes every time there
// is progress with a "pull" request sent to the service. If this function
// returns an error, [Client.Pull] will stop the process and return this error.
type PullProgressFunc func(ollama.ProgressResponse) error

// Pull downloads a model from the ollama library. fn is called each time
// progress is made on the request and can be used to display a progress bar,
// etc.
func (c *Client) Pull(ctx context.Context, req *ollama.PullRequest, fn PullProgressFunc) error {
	return c.stream(ctx, http.MethodPost, "/api/pull", req, func(bts []byte) error {
		var resp ollama.ProgressResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}

		return fn(resp)
	})
}

// PushProgressFunc is a function that [Client.Push] invokes when progress is
// made.
// It's similar to other progress function types like [PullProgressFunc].
type PushProgressFunc func(ollama.ProgressResponse) error

// Push uploads a model to the model library; requires registering for ollama.ai
// and adding a public key first. fn is called each time progress is made on
// the request and can be used to display a progress bar, etc.
func (c *Client) Push(ctx context.Context, req *ollama.PushRequest, fn PushProgressFunc) error {
	return c.stream(ctx, http.MethodPost, "/api/push", req, func(bts []byte) error {
		var resp ollama.ProgressResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}

		return fn(resp)
	})
}

// CreateProgressFunc is a function that [Client.Create] invokes when progress
// is made.
// It's similar to other progress function types like [PullProgressFunc].
type CreateProgressFunc func(ollama.ProgressResponse) error

// Create creates a model from a [Modelfile]. fn is a progress function that
// behaves similarly to other methods (see [Client.Pull]).
//
// [Modelfile]: https://github.com/ollama/ollama/blob/main/docs/modelfile.md
func (c *Client) Create(ctx context.Context, req *ollama.CreateRequest, fn CreateProgressFunc) error {
	return c.stream(ctx, http.MethodPost, "/api/create", req, func(bts []byte) error {
		var resp ollama.ProgressResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}

		return fn(resp)
	})
}

// List lists models that are available locally.
func (c *Client) List(ctx context.Context) (*ollama.ListResponse, error) {
	var lr ollama.ListResponse
	if err := c.do(ctx, http.MethodGet, "/api/tags", nil, &lr); err != nil {
		return nil, err
	}
	return &lr, nil
}

// List running models.
func (c *Client) ListRunning(ctx context.Context) (*ollama.ProcessResponse, error) {
	var lr ollama.ProcessResponse
	if err := c.do(ctx, http.MethodGet, "/api/ps", nil, &lr); err != nil {
		return nil, err
	}
	return &lr, nil
}

// Copy copies a model - creating a model with another name from an existing
// model.
func (c *Client) Copy(ctx context.Context, req *ollama.CopyRequest) error {
	if err := c.do(ctx, http.MethodPost, "/api/copy", req, nil); err != nil {
		return err
	}
	return nil
}

// Delete deletes a model and its data.
func (c *Client) Delete(ctx context.Context, req *ollama.DeleteRequest) error {
	if err := c.do(ctx, http.MethodDelete, "/api/delete", req, nil); err != nil {
		return err
	}
	return nil
}

// Show obtains model information, including details, modelfile, license etc.
func (c *Client) Show(ctx context.Context, req *ollama.ShowRequest) (*ollama.ShowResponse, error) {
	var resp ollama.ShowResponse
	if err := c.do(ctx, http.MethodPost, "/api/show", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Hearbeat checks if the server has started and is responsive; if yes, it
// returns nil, otherwise an error.
func (c *Client) Heartbeat(ctx context.Context) error {
	if err := c.do(ctx, http.MethodHead, "/", nil, nil); err != nil {
		return err
	}
	return nil
}

// Embed generates embeddings from a model.
func (c *Client) Embed(ctx context.Context, req *ollama.EmbedRequest) (*ollama.EmbedResponse, error) {
	var resp ollama.EmbedResponse
	if err := c.do(ctx, http.MethodPost, "/api/embed", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Embeddings generates embeddings from a model.
func (c *Client) Embeddings(ctx context.Context, req *ollama.EmbeddingRequest) (*ollama.EmbeddingResponse, error) {
	var resp ollama.EmbeddingResponse
	if err := c.do(ctx, http.MethodPost, "/api/embeddings", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateBlob creates a blob from a file on the server. digest is the
// expected SHA256 digest of the file, and r represents the file.
func (c *Client) CreateBlob(ctx context.Context, digest string, r io.Reader) error {
	return c.do(ctx, http.MethodPost, fmt.Sprintf("/api/blobs/%s", digest), r, nil)
}

// Version returns the Ollama server version as a string.
func (c *Client) Version(ctx context.Context) (string, error) {
	var version struct {
		Version string `json:"version"`
	}

	if err := c.do(ctx, http.MethodGet, "/api/version", nil, &version); err != nil {
		return "", err
	}

	return version.Version, nil
}
