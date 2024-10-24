package meme

import (
	"fmt"
	"github.com/kohmebot/meme/meme/generator"
	"github.com/kohmebot/pkg/gopool"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/shell"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func (p *PluginMeme) SetOnCommand(engine *zero.Engine) {
	engine.OnCommand("meme", p.env.Groups().Rule()).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		if !p.t.AddTask(uid) {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你还有正在进行中的任务哦"))
			return
		}
		gopool.Go(func() {
			// TODO 为每个人限流
			var err error
			defer func() {
				if err != nil {
					p.env.Error(ctx, err)
				}
				p.t.Done(uid)
			}()

			arguments := shell.Parse(ctx.State["args"].(string))
			if len(arguments) < 0 {
				err = fmt.Errorf("参数错误")
				return
			}
			keyword := arguments[0]
			key, ok := p.keywordMp[keyword]
			if !ok {
				err = fmt.Errorf(`"%s"不支持`, keyword)
				return
			}
			desc := p.dcsMp[key]
			req := generator.Request{
				Key: key,
			}

			err = req.ParseArgs(arguments[1:], desc)
			if err != nil {
				return
			}

			for _, segment := range ctx.Event.Message {
				switch segment.Type {
				case "image":
					url := segment.Data["url"]
					req.ImageUrls = append(req.ImageUrls, url)
				}
			}

			err = req.Validate(desc)
			if err != nil {
				return
			}
			img, err := p.g.Generate(&req)
			if err != nil {
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.ImageBytes(img))
		})

	})
}
