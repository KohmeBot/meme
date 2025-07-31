package meme

import (
	"fmt"
	"github.com/kohmebot/meme/meme/generator"
	"github.com/kohmebot/pkg/chain"
	"github.com/kohmebot/pkg/gopool"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/extension/shell"
	"github.com/wdvxdr1123/ZeroBot/message"
	"strconv"
	"strings"
	"time"
)

// CommandRuleTrimReply 在匹配前时修剪掉第一个reply的值，在匹配结束时重新插入
func CommandRuleTrimReply(next zero.Rule) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if len(ctx.Event.Message) == 0 {
			return false
		}
		if ctx.Event.Message[0].Type == "reply" {
			raw := ctx.Event.Message
			ctx.Event.Message = ctx.Event.Message[1:]
			defer func() {
				ctx.Event.Message = raw
			}()
		}
		return next(ctx)
	}

}

func (p *PluginMeme) SetOnCommand(engine *zero.Engine) {
	engine.OnMessage(CommandRuleTrimReply(zero.CommandRule("meme")), p.env.Groups().Rule()).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
			desc, ok := p.searchDesc(keyword)
			if !ok {
				err = fmt.Errorf(`"%s"不支持`, keyword)
				return
			}
			req := generator.Request{
				Key: desc.Key,
			}

			err = req.ParseArgs(arguments[1:], desc)
			if err != nil {
				return
			}

			// 解析消息

			parseMessageToReq(ctx, ctx.Event.Message, &req, p.conf.AvatarSizeToParam())

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
	engine.OnCommand("mhelp", p.env.Groups().Rule()).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var cmd extension.CommandModel
		err := ctx.Parse(&cmd)
		if err != nil {
			p.env.Error(ctx, err)
			return
		}
		gopool.Go(func() {
			uid := ctx.Event.UserID
			gid := ctx.Event.GroupID
			if len(cmd.Args) > 0 {
				if !p.t.AddTask(uid) {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你还有正在进行中的任务哦"))
					return
				}
				p.handleTargetHelp(ctx, cmd.Args)
				p.t.Done(uid)
			} else {
				ok, id := p.tt.AddTask(gid)
				if !ok {
					ctx.SendChain(message.Reply(id), message.At(uid), message.Text(" 之前已经说过一次了..."))
					return
				}
				p.tt.Done(gid, p.handleAllHelp(ctx))

			}
		})

	})
}

func (p *PluginMeme) handleTargetHelp(ctx *zero.Ctx, target string) {
	var err error
	defer func() {
		if err != nil {
			p.env.Error(ctx, err)
			return
		}
	}()
	target = strings.TrimSpace(target)
	desc, ok := p.searchDesc(target)
	if !ok {
		err = fmt.Errorf(`"%s"不支持`, target)
		return
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("指令详情：\n"))
	builder.WriteString("@某人会获取对方的头像\n")
	builder.WriteString(fmt.Sprintf("Key：%s\n", desc.Key))
	builder.WriteString(fmt.Sprintf("可触发指令：%s\n", strings.Join(desc.Keywords, " ")))
	builder.WriteString(fmt.Sprintf("最少需要图片数：%d\n", desc.MinImages))
	builder.WriteString(fmt.Sprintf("最多支持图片数：%d\n", desc.MaxImages))
	builder.WriteString(fmt.Sprintf("最少需要文本数：%d\n", desc.MinTexts))
	builder.WriteString(fmt.Sprintf("最多支持文本数：%d\n", desc.MaxTexts))
	if len(desc.DefaultTexts) > 0 {
		builder.WriteString(fmt.Sprintf("默认文本: %s\n", strings.Join(desc.DefaultTexts, " ")))
	}
	if desc.Args.Len() > 0 {
		builder.WriteString(fmt.Sprintf("额外参数：\n"))
		for name, arg := range desc.Args.Range {
			builder.WriteString(fmt.Sprintf("[-%s](%s)%s", name, arg.Type, arg.Description))
			if len(arg.Enum) > 0 {
				builder.WriteString(fmt.Sprintf("<%s>", strings.Join(arg.Enum, " ")))
			}
		}
		builder.WriteByte('\n')
	}
	builder.WriteString(fmt.Sprintf("生成预览："))
	var msgChain chain.MessageChain
	msgChain.Join(message.Reply(ctx.Event.MessageID))
	msgChain.Line(message.Text(builder.String()))
	img, pErr := p.g.GetPreview(desc.Key)
	if pErr != nil {
		msgChain.Join(message.Text(fmt.Sprintf("生成预览失败: %s", pErr.Error())))
	} else {
		msgChain.Join(message.ImageBytes(img))
	}
	ctx.Send(msgChain)

}

func (p *PluginMeme) handleAllHelp(ctx *zero.Ctx) message.ID {
	var builder strings.Builder
	builder.WriteString("以下是支持的制图指令\n可通过mhelp [keyword]来查看对应详情\n")

	var result [][]generator.CommandDesc
	for i := 0; i < len(p.descs); i += 30 {
		end := i + 30
		if end > len(p.descs) {
			end = len(p.descs)
		}
		result = append(result, p.descs[i:end])
	}
	// 每30个处理一次
	var mid message.ID
	for _, descs := range result {
		for _, desc := range descs {
			builder.WriteByte(' ')
			builder.WriteString(fmt.Sprintf("(%s)", desc.Keywords[0]))
		}
		mid = ctx.Send(message.Text(builder.String()))
		time.Sleep(time.Second)
		builder.Reset()
	}
	return mid
}

func (p *PluginMeme) searchDesc(k string) (desc generator.CommandDesc, ok bool) {
	key, ok := p.keywordMp[k]
	if ok {
		desc, ok = p.descMp[key]
	} else {
		desc, ok = p.descMp[k]
	}
	return
}

// parseMessageToReq 解析消息到请求
func parseMessageToReq(ctx *zero.Ctx, message message.Message, req *generator.Request, avatarSize int) {
	for _, segment := range message {
		switch segment.Type {
		case "reply":
			// TODO 需要剔除上一个消息的reply
			id := segment.Data["id"]
			msgs := ctx.GetMessage(id).Elements
			if len(msgs) > 0 && msgs[0].Type == "reply" {
				msgs = msgs[1:]
			}
			parseMessageToReq(ctx, msgs, req, avatarSize)
		case "image":
			fileName := segment.Data["file"]
			res := ctx.GetImage(fileName)
			req.Images = append(req.Images, &generator.Image{FileName: res.Map()["file"].String()})
		case "at":
			qq, _ := strconv.Atoi(segment.Data["qq"])
			url := fmt.Sprintf("https://q4.qlogo.cn/g?b=qq&nk=%d&s=%d", qq, avatarSize)
			req.Images = append(req.Images, &generator.Image{Url: url})
		}
	}
}
