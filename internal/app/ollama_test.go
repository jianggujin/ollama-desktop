package app

import (
	"context"
	"net/http"
	"net/url"
	ollama2 "ollama-desktop/internal/ollama/ollama"
	"testing"
)

func TestApp_OllamaPull(t *testing.T) {
	//fn := func(resp api.ProgressResponse) error {
	//	if resp.Digest != "" {
	//		if spinner != nil {
	//			spinner.Stop()
	//		}
	//
	//		bar, ok := bars[resp.Digest]
	//		if !ok {
	//			bar = progress.NewBar(fmt.Sprintf("pulling %s...", resp.Digest[7:19]), resp.Total, resp.Completed)
	//			bars[resp.Digest] = bar
	//			p.Add(resp.Digest, bar)
	//		}
	//
	//		bar.Set(resp.Completed)
	//	} else if status != resp.Status {
	//		if spinner != nil {
	//			spinner.Stop()
	//		}
	//
	//		status = resp.Status
	//		spinner = progress.NewSpinner(status)
	//		p.Add(status, spinner)
	//	}
	//
	//	return nil
	//}

	//ollama_test.go:19: Status: pulling manifest, Digest: , Total: 0, Completed: 0
	//ollama_test.go:19: Status: pulling 8de95da68dc4, Digest: sha256:8de95da68dc485c0889c205384c24642f83ca18d089559c977ffc6a3972a71a8, Total: 352151968, Completed: 352151968
	//ollama_test.go:19: Status: pulling 62fbfd9ed093, Digest: sha256:62fbfd9ed093d6e5ac83190c86eec5369317919f4b149598d2dbb38900e9faef, Total: 182, Completed: 182
	//ollama_test.go:19: Status: pulling c156170b718e, Digest: sha256:c156170b718ec29139d3653d40ed1986fd92fb7e0959b5c71f3c48f62e6636f4, Total: 11344, Completed: 11344
	//ollama_test.go:19: Status: pulling f02dd72bb242, Digest: sha256:f02dd72bb2423204352eabc5637b44d79d17f109fdb510a7c51455892aa2d216, Total: 59, Completed: 59
	//ollama_test.go:19: Status: pulling 2184ab82477b, Digest: sha256:2184ab82477bc33a5e08fa209df88f0631a19e686320cce2cfe9e00695b2f0e6, Total: 488, Completed: 488
	//ollama_test.go:19: Status: verifying sha256 digest, Digest: , Total: 0, Completed: 0
	//ollama_test.go:19: Status: writing manifest, Digest: , Total: 0, Completed: 0
	//ollama_test.go:19: Status: removing any unused layers, Digest: , Total: 0, Completed: 0
	//ollama_test.go:19: Status: success, Digest: , Total: 0, Completed: 0
	//_ = app.OllamaPull(request, func(response ollama.ProgressResponse) error {
	//	t.Log(fmt.Sprintf("Status: %s, Digest: %s, Total: %d, Completed: %d", response.Status, response.Digest, response.Total, response.Completed))
	//	return nil
	//})
}

func TestOllama_ModelInfoOnline(t *testing.T) {
	base, _ := url.Parse("https://ollama.com")

	client := &ollama2.Client{
		Base: base,
		Http: http.DefaultClient,
	}
	resp, err := client.ModelInfo(context.Background(), "llama3.1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("model", resp.Model)
	t.Log("tags")
	for _, tag := range resp.Tags {
		t.Log(tag)
	}
	t.Log("metas")
	for _, meta := range resp.Metas {
		t.Log(meta)
	}
	t.Log("readme", resp.Readme)
}
