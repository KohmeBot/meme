package meme

import (
	"github.com/kohmebot/meme/meme/generator"
	"github.com/kohmebot/plugin/v2"
	"github.com/wdvxdr1123/ZeroBot"
	"slices"
	"time"
)

type PluginMeme struct {
	env       plugin.Env
	conf      Config
	g         *generator.MemeGenerator
	descMp    map[string]generator.CommandDesc
	descs     []generator.CommandDesc
	keywordMp map[string]string
	t         *HelpTasks
	tt        *TaskDuration
}

func NewPlugin() plugin.Plugin {
	return &PluginMeme{keywordMp: make(map[string]string), t: NewTasks()}
}

func (p *PluginMeme) OnInit(engine plugin.Engine, env plugin.Env) error {

	p.env = env
	err := p.env.GetConf(&p.conf)
	if err != nil {
		return err
	}
	p.g = generator.NewGenerator(p.conf.Url)
	p.tt = NewTaskDuration(time.Duration(p.conf.HelpDuration) * time.Second)
	p.descMp, err = p.g.GetCommandsWithRetry(time.Minute)
	if err != nil {
		return err
	}

	for _, desc := range p.descMp {
		desc.SortKeywords()
		desc.KeywordsMappingKeyTo(p.keywordMp)
		p.descs = append(p.descs, desc)
	}

	slices.SortFunc(p.descs, func(i, j generator.CommandDesc) int {
		return len(i.Keywords[0]) - len(j.Keywords[0])
	})

	p.SetOnCommand(engine)

	return nil
}

func (p *PluginMeme) OnHelp(ctx *zero.Ctx) {
	p.onHelp(ctx)
}

func (p *PluginMeme) Name() string {
	return "meme"
}

func (p *PluginMeme) Version() string {
	return "v1.0.1"
}

func (p *PluginMeme) OnBoot() {

}
