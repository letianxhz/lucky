package config

// Config 网关服务配置
type Config struct {
	// TCP监听地址
	TCPAddr string `yaml:"tcp_addr"`

	// WebSocket监听地址
	WSAddr string `yaml:"ws_addr"`

	// 心跳超时时间（秒）
	HeartbeatTimeout int32 `yaml:"heartbeat_timeout"`

	// 消息积压数量
	WriteBacklog int32 `yaml:"write_backlog"`
}

// Get 获取配置实例（单例）
var instance *Config

// Get 获取配置
func Get() *Config {
	if instance == nil {
		instance = &Config{
			TCPAddr:          ":10011",
			WSAddr:           "",
			HeartbeatTimeout: 60,
			WriteBacklog:     64,
		}
	}
	return instance
}

// MustInitialize 初始化配置（必须成功）
func MustInitialize(configPath string) {
	// TODO: 从配置文件加载配置
	instance = &Config{
		TCPAddr:          ":10011",
		WSAddr:           "",
		HeartbeatTimeout: 60,
		WriteBacklog:     64,
	}
}
