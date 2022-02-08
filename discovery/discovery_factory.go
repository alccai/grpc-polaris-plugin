package discovery

import (
	"fmt"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
	"github.com/zhengheng7913/grpc-go-starter/naming/discovery"
	"github.com/zhengheng7913/grpc-go-starter/plugin"
	"google.golang.org/grpc/resolver"
	"strings"
)

const (
	PluginName = "polaris"
)

var (
	ErrPluginDecoderEmpty = fmt.Errorf("plugin decoder is empty")
)

func init() {
	plugin.Register(PluginName, &Factory{})
}

func newConfiguration(cfg *FactoryConfig) config.Configuration {
	var c config.Configuration
	if len(cfg.AddressList) > 0 {
		addressList := strings.Split(cfg.AddressList, ",")
		c = config.NewDefaultConfiguration(addressList)
	} else {
		c = config.NewDefaultConfigurationWithDomain()
	}
	return c
}

func newConsumer(cfg *FactoryConfig) (api.ConsumerAPI, error) {
	sdkCtx, err := api.InitContextByConfig(newConfiguration(cfg))
	if err != nil {
		return nil, err
	}
	consumerAPI := api.NewConsumerAPIByContext(sdkCtx)
	return consumerAPI, nil
}

type Factory struct {
}

func (f Factory) Type() string {
	return discovery.PluginType
}

func (f Factory) Setup(name string, dec plugin.Decoder) error {
	if dec == nil {
		return ErrPluginDecoderEmpty
	}
	cfg := &FactoryConfig{}
	if err := dec.Decode(cfg); err != nil {
		return err
	}
	consumer, err := newConsumer(cfg)
	if err != nil {
		return err
	}
	for _, client := range cfg.Clients {
		cfg := &Config{
			Name: client.Name,
		}
		d := newDiscovery(consumer, cfg)
		discovery.Register(cfg.Name, d)
	}
	resolver.Register(NewPolarisResolverBuilder(consumer))
	return nil
}

func (f Factory) Destroy() error {
	return nil
}
