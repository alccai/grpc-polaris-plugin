package grpc_registry

import "time"

type Config struct {
	// ServiceToken 服务访问Token
	ServiceToken string
	// Protocol 服务端访问方式，支持 http grpc，默认 grpc
	Protocol string
	// HeartBeat 上报心跳时间间隔，默认为建议 为TTL/2
	HeartBeat int
	// EnableRegister 默认只上报心跳，不注册服务，为 true 则启动注册
	EnableRegister bool
	// Weight
	Weight int
	// TTL 单位s，服务端检查周期实例是否健康的周期
	TTL int
	// InstanceID 实例名
	InstanceID string
	// Namespace 命名空间
	Namespace string
	// ServiceName 服务名
	ServiceName string
	// BindAddress 指定上报地址
	BindAddress string
	// Metadata 用户自定义 metadata 信息
	Metadata map[string]string
	// DisableHealthCheck 禁用健康检查
	DisableHealthCheck bool

	Version string
}

type FactoryConfig struct {
	Protocol           string         `yaml:"protocol"`
	HeartbeatInterval  int            `yaml:"heartbeat_interval"`
	Services           []Service      `yaml:"service"`
	EnableRegister     bool           `yaml:"register_self"`
	AddressList        string         `yaml:"address_list"`
	ClusterService     ClusterService `yaml:"cluster_service"`
	ConnectTimeout     int            `yaml:"connect_timeout"`
	MessageTimeout     *time.Duration `yaml:"message_timeout"`
	DisableHealthCheck bool           `yaml:"disable_health_check"`
	Version            string         `yaml:"version"`
}

type Service struct {
	Namespace      string            `yaml:"namespace"`
	ServiceName    string            `yaml:"name"`
	Token          string            `yaml:"token"`
	InstanceID     string            `yaml:"instance_id"`
	Weight         int               `yaml:"weight"`
	BindAddress    string            `yaml:"bind_address"`
	MetaData       map[string]string `yaml:"metadata"`
	Protocol       string            `yaml:"protocol"`
	EnableRegister bool              `yaml:"register_self"`
}

// ClusterService 集群服务
type ClusterService struct {
	Discover    string `yaml:"discover"`
	HealthCheck string `yaml:"health_check"`
	Monitor     string `yaml:"monitor"`
}
