package typewriter

import "os"

type Config struct {
	Filter                func(os.FileInfo) bool
	IgnoreTypeCheckErrors bool
}

var DefaultConfig = &Config{}
