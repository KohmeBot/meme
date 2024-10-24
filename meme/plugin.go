package meme

import (
	"fmt"
	"github.com/kohmebot/meme/meme/generator"
	"github.com/kohmebot/pkg/command"
	"github.com/kohmebot/pkg/version"
	"github.com/kohmebot/plugin"
	"github.com/wdvxdr1123/ZeroBot"
)

type PluginMeme struct {
	env       plugin.Env
	conf      Config
	g         *generator.MemeGenerator
	dcsMp     map[string]generator.CommandDesc
	keywordMp map[string]string
	t         *Tasks
}

func NewPlugin() plugin.Plugin {
	return &PluginMeme{keywordMp: make(map[string]string), t: NewTasks()}
}

func (p *PluginMeme) Init(engine *zero.Engine, env plugin.Env) error {
	p.env = env
	err := p.env.GetConf(&p.conf)
	if err != nil {
		return err
	}
	p.g = generator.NewGenerator(p.conf.Url)
	p.dcsMp, err = p.g.GetCommands()
	if err != nil {
		return nil
	}

	for _, desc := range p.dcsMp {
		desc.KeywordsMappingKeyTo(p.keywordMp)
	}

	p.SetOnCommand(engine)
	return nil
}

func (p *PluginMeme) Name() string {
	return "meme"
}

func (p *PluginMeme) Description() string {
	return "表情包生成"
}

func (p *PluginMeme) Commands() fmt.Stringer {
	return command.NewCommands(
		command.NewCommand("查看生成帮助", "help meme"),
		command.NewCommand("生成图片", "meme"),
	)
}

func (p *PluginMeme) Version() uint64 {
	return uint64(version.NewVersion(0, 0, 10))
}

func (p *PluginMeme) OnBoot() {

}
