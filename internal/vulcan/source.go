package vulcan

import (
	"embed"
	"github.com/hashicorp/go-version"
	"os"
	"strings"
)

type SourceInfo struct {
	Version *version.Version
	Uid     string
}

type Source interface {
	Scan() ([]*SourceInfo, error)
	Read(string) ([]byte, error)
}

type EmbedFSSource struct {
	Fs    embed.FS
	Paths []string
}

func (s *EmbedFSSource) Scan() ([]*SourceInfo, error) {
	var infos []*SourceInfo
	for _, path := range s.Paths {
		files, err := s.Fs.ReadDir(path)
		if err != nil {
			// 未找到，则表明不需要升级
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".xml") {
				continue
			}
			v, err := version.NewVersion(strings.TrimSuffix(f.Name(), ".xml"))
			if err != nil {
				return nil, err
			}
			infos = append(infos, &SourceInfo{
				Version: v,
				Uid:     path + "/" + f.Name(),
			})
		}
	}
	return infos, nil
}

func (s *EmbedFSSource) Read(uid string) ([]byte, error) {
	return s.Fs.ReadFile(uid)
}
