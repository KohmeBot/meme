package meme

type Config struct {
	// meme-generator 的api地址
	Url string `mapstructure:"url"`
	// 每个群内帮助命令在间隔多少秒内只能发送一次(因为帮助可能比较长)
	HelpDuration int64 `mapstructure:"help_duration"`
}
