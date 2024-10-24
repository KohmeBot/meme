package meme

import (
	"fmt"
	"github.com/kohmebot/meme/meme/generator"
	"github.com/kohmebot/pkg/gopool"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/shell"
	"github.com/wdvxdr1123/ZeroBot/message"
	"strconv"
	"strings"
)

func (p *PluginMeme) SetOnCommand(engine *zero.Engine) {
	engine.OnCommand("meme", p.env.Groups().Rule()).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		if !p.t.AddTask(uid) {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你还有正在进行中的任务哦"))
			return
		}
		gopool.Go(func() {
			var err error
			defer func() {
				if err != nil {
					p.env.Error(ctx, err)
				}
				p.t.Done(uid)
			}()

			arguments := shell.Parse(ctx.State["args"].(string))
			if len(arguments) <= 0 {
				err = fmt.Errorf("参数错误")
				return
			}
			keyword := arguments[0]
			key, ok := p.keywordMp[keyword]
			if !ok {
				err = fmt.Errorf(`"%s"不支持`, keyword)
				return
			}
			desc := p.descMp[key]
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
				case "at":
					qq, _ := strconv.Atoi(segment.Data["qq"])
					url := fmt.Sprintf("https://q4.qlogo.cn/g?b=qq&nk=%d&s=%d", qq, p.conf.AvatarSizeToParam())
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

func (p *PluginMeme) SetOnHelp(engine *zero.Engine) {
	engine.OnCommand("mhelp", p.env.Groups().Rule()).Handle(func(ctx *zero.Ctx) {
		if !p.tt.AddTask(ctx.Event.GroupID) {
			ctx.Send(message.Text("在不久前好像问过一次了..."))
			return
		}
		var builder strings.Builder
		builder.WriteString("以下是支持的制图指令:\n")
		for _, desc := range p.descs {
			builder.WriteString(fmt.Sprintf("(%s)", strings.Join(desc.Keywords, " ")))
			if len(desc.Args) > 0 {
				builder.WriteString("|")
			}
			for _, arg := range desc.Args {
				builder.WriteString(fmt.Sprintf("[-%s]%s", arg.Name, arg.Description))
			}
			builder.WriteByte('\n')
		}
		ctx.Send(message.Text(builder.String()))
	})
}
