package meme

type Config struct {
	// meme-generator 的api地址
	Url string `mapstructure:"url"`
	// 每个群内帮助命令在间隔多少秒内只能发送一次(因为帮助可能比较长)
	HelpDuration int64 `mapstructure:"help_duration"`
	// 获取用户头像的质量(1,2,3)三档
	AvatarSize int64 `mapstructure:"avatar_size"`
}

func (c *Config) AvatarSizeToParam() int {
	if c.AvatarSize <= 1 {
		return 100
	}
	if c.AvatarSize == 2 {
		return 140
	}
	return 640
}
