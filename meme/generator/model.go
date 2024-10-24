package generator

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kohmebot/pkg/gopool"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

// CommandDesc 生成的命令描述
type CommandDesc struct {
	Key      string   `json:"key"`
	Keywords []string `json:"keywords"`
	Params   `json:"params_type"`
}

// KeywordsMappingKeyTo 映射关键词
func (d *CommandDesc) KeywordsMappingKeyTo(to map[string]string) {
	for _, keyword := range d.Keywords {
		to[keyword] = d.Key
	}
}

func (d *CommandDesc) ParseArgs(k, v string) (val any, found bool, err error) {
	for _, arg := range d.Args {
		if arg.Name == k {
			val, err = arg.ParseValue(v)
			found = true
			return
		}
	}
	return nil, false, nil
}

type Params struct {
	// 最小需要的图片数量
	MinImages int `json:"min_images"`
	// 最大允许的图片数量
	MaxImages int `json:"max_images"`
	// 最小需要的文本数量
	MinTexts int `json:"min_texts"`
	// 最大允许的文本数量
	MaxTexts int `json:"max_texts"`
	// 默认的文本
	DefaultTexts []string `json:"default_texts"`
	// 额外参数
	Args []Arg `json:"args_type"`
}

type Arg struct {
	// 参数名称
	Name string `json:"name"`
	// 参数类型
	Type string `json:"type"`
	// 参数描述
	Description string `json:"description"`
	// 默认参数
	Default any `json:"default"`
	// 参数的可选值
	Enum []string `json:"enum"`
}

// ParseValue 解析参数值
func (a *Arg) ParseValue(s string) (any, error) {
	switch a.Type {
	case "string":
		return s, nil
	case "boolean":
		return strconv.ParseBool(s)
	case "integer":
		return strconv.ParseInt(s, 10, 64)
	}

	return nil, fmt.Errorf("unknown type %s", a.Type)
}

type Request struct {
	Key       string
	Images    [][]byte
	ImageUrls []string
	imagesMd5 []string
	Texts     []string
	ArgsRaw   map[string]any
	// json 字符串 由 ArgsRaw 生成
	Args string
}

type Response struct {
	Detail string `json:"detail"`
}

func (r *Request) ParseArgs(args []string, desc CommandDesc) error {
	for _, arg := range args {
		k, ok := strings.CutPrefix(arg, "-")
		if ok {
			split := strings.SplitN(k, "=", 1)
			if len(split) < 2 {
				continue
			}
			k = split[0]
			v := split[1]
			val, found, err := desc.ParseArgs(k, v)
			if found {
				if err != nil {
					return err
				}
				r.ArgsRaw[k] = val
			}
		} else {
			r.Texts = append(r.Texts, arg)
		}
	}
	return nil
}

// Validate 验证req和desc是否一致
func (r *Request) Validate(desc CommandDesc) error {
	if len(r.ImageUrls) < desc.MinImages {
		return fmt.Errorf("最少需要%d张图片 (%d/%d)", desc.MinImages, len(r.ImageUrls), desc.MinImages)
	}
	if len(r.ImageUrls) > desc.MaxImages {
		return fmt.Errorf("最多支持%d张图片", desc.MaxImages)
	}
	if len(r.Texts) < desc.MinTexts {
		return fmt.Errorf("最少需要%d条文本 (%d/%d)", desc.MinTexts, len(r.Texts), desc.MinTexts)
	}
	if len(r.Texts) > desc.MaxTexts {
		return fmt.Errorf("最多支持%d条文本", desc.MaxTexts)
	}
	return nil

}

func (r *Request) writeFormTo(w *multipart.Writer) (err error) {
	// 写入普通文本字段
	for _, text := range r.Texts {
		err = w.WriteField("texts", text)
		if err != nil {
			return fmt.Errorf("error writing texts field: %v", err)
		}
	}
	// 写入 Args 字段
	err = w.WriteField("args", r.Args)
	if err != nil {
		return fmt.Errorf("error writing args field: %v", err)
	}

	// 写入图片字段
	for idx, image := range r.Images {
		m := r.imagesMd5[idx]
		part, err := w.CreateFormFile("images", m)
		if err != nil {
			return fmt.Errorf("error creating form file: %v", err)
		}
		_, err = part.Write(image)
		if err != nil {
			return fmt.Errorf("error writing image data: %v", err)
		}
	}
	return

}

func (r *Request) loadImagesFromUrl(cli *http.Client) error {
	wg := sync.WaitGroup{}
	if cap(r.Images) < len(r.ImageUrls) {
		r.Images = make([][]byte, 0, len(r.ImageUrls))
	}
	r.Images = r.Images[:len(r.ImageUrls)]
	errorSlice := make([]error, len(r.ImageUrls))
	for idx, url := range r.ImageUrls {
		wg.Add(1)
		gopool.Go(func() {
			defer wg.Done()
			r.Images[idx], errorSlice[idx] = DownloadImage(cli, url)
		})
	}
	wg.Wait()
	return errors.Join(errorSlice...)
}

func (r *Request) countImageHash() {
	for _, image := range r.Images {
		hash := md5.Sum(image)
		// 将哈希值转换为字符串表示
		md5String := hex.EncodeToString(hash[:])
		r.imagesMd5 = append(r.imagesMd5, md5String)
	}
}

func (r *Request) encodeArgJson() error {
	jsonData, err := json.Marshal(r.ArgsRaw)
	if err != nil {
		return err
	}
	r.Args = unsafe.String(unsafe.SliceData(jsonData), len(jsonData))
	return nil
}

type Form struct {
	Images [][]byte
	Texts  []string
	Args   string
}

func DownloadImage(cli *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
