package temporal

type Config struct {
	Host           string `mapstructure:"host"`
	Namespace      string `mapstructure:"namespace"`
	TaskQueue      string `mapstructure:"task_queue"`
	EnableTLS      bool   `mapstructure:"enable_tls"`
	EnableEncoder  bool   `mapstructure:"enable_encoder"`
	EnableCompress bool   `mapstructure:"enable_compress"`
}
