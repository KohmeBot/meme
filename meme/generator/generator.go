package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kohmebot/pkg/gopool"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
)

// MemeGenerator Meme生成器
type MemeGenerator struct {
	Url  string
	desc []CommandDesc
	cli  *http.Client
}

func NewGenerator(api string) *MemeGenerator {
	return &MemeGenerator{
		Url: api + "/" + "memes/",
		cli: &http.Client{},
	}
}

func (g *MemeGenerator) GetCommands() (map[string]CommandDesc, error) {
	keys, err := g.getKeys()
	if err != nil {
		return nil, err
	}
	return g.getInfos(keys)
}

func (g *MemeGenerator) Generate(doReq *Request) ([]byte, error) {
	if len(doReq.ArgsRaw) > 0 {
		err := doReq.encodeArgJson()
		if err != nil {
			return nil, err
		}
	}

	if len(doReq.ImageUrls) > 0 {
		err := doReq.loadImagesFromUrl(g.cli)
		if err != nil {
			return nil, err
		}
	}
	doReq.countImageHash()
	buf := newBuffer()
	defer buf.Recycle()
	wr := multipart.NewWriter(buf)
	err := doReq.writeFormTo(wr)
	if err != nil {
		return nil, err
	}
	err = wr.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s/", g.Url, doReq.Key), buf)
	// 设置 Content-Type
	req.Header.Set("Content-Type", wr.FormDataContentType())
	resp, err := g.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return io.ReadAll(resp.Body)
	}
	r := Response{}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("status %d: %v", resp.StatusCode, err)
	}
	if err = json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("status %d: %v", resp.StatusCode, err)
	}
	return nil, fmt.Errorf("status %d: %s", resp.StatusCode, r.Detail)
}

func (g *MemeGenerator) getKeys() ([]string, error) {
	url := g.Url + "keys"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := g.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var keys []string
	err = json.Unmarshal(buf, &keys)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (g *MemeGenerator) getInfo(key string) (desc CommandDesc, err error) {
	url := g.Url + key + "/" + "info"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err := g.cli.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf, &desc)
	return
}

func (g *MemeGenerator) getInfos(keys []string) (descMp map[string]CommandDesc, err error) {
	dcs := make([]CommandDesc, len(keys))
	errorSlice := make([]error, len(keys))
	wg := sync.WaitGroup{}
	for idx, key := range keys {
		wg.Add(1)
		gopool.Go(func() {
			defer wg.Done()
			dcs[idx], errorSlice[idx] = g.getInfo(key)
		})
	}
	wg.Wait()
	err = errors.Join(errorSlice...)
	if err != nil {
		return nil, err
	}
	descMp = make(map[string]CommandDesc, len(dcs))
	for idx, dc := range dcs {
		descMp[keys[idx]] = dc
	}
	return
}
