package generator

import (
	"flag"
	"testing"
)

func TestGetCommands(t *testing.T) {
	g := NewGenerator("http://127.0.0.1:2233")
	mp, err := g.GetCommands()
	if err != nil {
		t.Fatal(err)
	}

	for key, desc := range mp {
		t.Log(key, desc)
	}

}

func TestGenerate(t *testing.T) {
	g := NewGenerator("http://127.0.0.1:2233")
	req := &Request{
		Key:       "anya_suki",
		ImageUrls: []string{"https://pica.zhimg.com/70/v2-da40e549a90d4d5e95af35e532a7f608_1440w.avis?source=172ae18b&biz_tag=Post"},
	}
	img, err := g.Generate(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(img)

}

func TestParse(t *testing.T) {
	fSet := flag.FlagSet{}

	err := fSet.Parse([]string{"摸", "摸摸你", "-file=true"})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(fSet.Args())

}
