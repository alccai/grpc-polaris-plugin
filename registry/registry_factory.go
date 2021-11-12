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
	PluginType            = "registry"
	defaultConnectTimeout = time.Second
	defaultMessageTimeout = time.Second
	defaultProtocol       = "grpc"
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
	return PluginType
}

func (f *Factory) Setup(name string, dec plugin.Decoder) error {
	if dec == nil {
		return fmt.Errorf("grpc-registry config decoder emtpy")
	}
	conf := &FactoryConfig{}
	if err := dec.Decode(conf); err != nil {
		return err
	}
	return register(conf)
}

func newProvider(cfg *FactoryConfig) (api.ProviderAPI, error) {
	var c *config.ConfigurationImpl
	if len(cfg.AddressList) > 0 {
		addressList := strings.Split(cfg.AddressList, ",")
		c = config.NewDefaultConfiguration(addressList)
	} else {
		c = config.NewDefaultConfigurationWithDomain()
	}
	// 配置 cluster
	if cfg.ClusterService.Discover != "" {
		c.Global.GetSystem().GetDiscoverCluster().SetService(cfg.ClusterService.Discover)
	}
	if cfg.ClusterService.HealthCheck != "" {
		c.Global.GetSystem().GetHealthCheckCluster().SetService(cfg.ClusterService.HealthCheck)
	}
	if cfg.ClusterService.Monitor != "" {
		c.Global.GetSystem().GetMonitorCluster().SetService(cfg.ClusterService.Monitor)
	}
	if cfg.Protocol == "" {
		cfg.Protocol = defaultProtocol
	}
	c.Global.ServerConnector.Protocol = cfg.Protocol
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
		reg, err := newRegistry(provider, cfg)
		if err != nil {
			return err
		}
		registry.Register(service.ServiceName, reg)
	}
	return nil
}
