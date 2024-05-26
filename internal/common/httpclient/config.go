package httpclient

import "time"

type Config struct {
	Timeout time.Duration `mapstructure:"timeout"`
}
