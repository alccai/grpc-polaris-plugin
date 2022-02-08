package grpc_registry

import (
	"fmt"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
	"github.com/zhengheng7913/grpc-go-starter/naming/registry"
	"github.com/zhengheng7913/grpc-go-starter/plugin"
	"strings"
	"time"
)

const (
	PluginName            = "polaris"
	defaultConnectTimeout = time.Second
	defaultMessageTimeout = time.Second
	defaultProtocol       = "grpc"
)

var (
	ErrPluginDecoderEmpty = fmt.Errorf("plugin decoder is empty")
)

func init() {
	plugin.Register(PluginName, &Factory{})
}

type Factory struct {
}

func (f *Factory) Destroy() error {
	return nil
}

func (f *Factory) Type() string {
	return registry.PluginType
}

func (f *Factory) Setup(name string, dec plugin.Decoder) error {
	if dec == nil {
		return ErrPluginDecoderEmpty
	}
	conf := &FactoryConfig{}
	if err := dec.Decode(conf); err != nil {
		return err
	}
	return register(conf)
}

func newProvider(cfg *FactoryConfig) (api.ProviderAPI, error) {
	var c config.Configuration
	if len(cfg.AddressList) > 0 {
		addressList := strings.Split(cfg.AddressList, ",")
		c = config.NewDefaultConfiguration(addressList)
	} else {
		c = config.NewDefaultConfigurationWithDomain()
	}
	// 配置 cluster
	if cfg.ClusterService.Discover != "" {
		c.GetGlobal().GetSystem().GetDiscoverCluster().SetService(cfg.ClusterService.Discover)
	}
	if cfg.ClusterService.HealthCheck != "" {
		c.GetGlobal().GetSystem().GetHealthCheckCluster().SetService(cfg.ClusterService.HealthCheck)
	}
	if cfg.ClusterService.Monitor != "" {
		c.GetGlobal().GetSystem().GetMonitorCluster().SetService(cfg.ClusterService.Monitor)
	}
	if cfg.Protocol == "" {
		cfg.Protocol = defaultProtocol
	}
	c.GetGlobal().GetServerConnector().SetProtocol(cfg.Protocol)
	if cfg.ConnectTimeout != 0 {
		c.GetGlobal().GetServerConnector().SetConnectTimeout(time.Duration(cfg.ConnectTimeout) * time.Millisecond)
	} else {
		c.GetGlobal().GetServerConnector().SetConnectTimeout(defaultConnectTimeout)
	}
	// 设置消息超时时间
	messageTimeout := defaultMessageTimeout
	if cfg.MessageTimeout != nil {
		messageTimeout = *cfg.MessageTimeout
	}
	c.GetGlobal().GetServerConnector().SetMessageTimeout(messageTimeout)
	provider, err := api.NewProviderAPIByConfig(c)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func register(conf *FactoryConfig) error {
	provider, err := newProvider(conf)
	if err != nil {
		return err
	}
	for _, service := range conf.Services {
		cfg := &Config{
			Protocol:           service.Protocol,
			EnableRegister:     service.EnableRegister,
			HeartBeat:          conf.HeartbeatInterval / 1000,
			ServiceName:        service.ServiceName,
			Namespace:          service.Namespace,
			ServiceToken:       service.Token,
			InstanceID:         service.InstanceID,
			Metadata:           service.MetaData,
			BindAddress:        service.BindAddress,
			DisableHealthCheck: conf.DisableHealthCheck,
			Version:            conf.Version,
		}
		reg := newRegistry(provider, cfg)
		registry.Register(service.Name, reg)
	}
	return nil
}
