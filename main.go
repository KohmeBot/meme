package main

import (
	"github.com/kohmebot/meme/meme"
	"github.com/kohmebot/plugin"
)

func NewPlugin() plugin.Plugin {
	return meme.NewPlugin()
}
