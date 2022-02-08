package discovery

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/polarismesh/polaris-go/api"
	"github.com/zhengheng7913/grpc-go-starter/naming/discovery"
	"github.com/zhengheng7913/grpc-go-starter/naming/registry"
)

func NewDiscovery(consumer api.ConsumerAPI, cfg *Config) discovery.Discovery {
	return newDiscovery(consumer, cfg)
}

func newDiscovery(consumer api.ConsumerAPI, cfg *Config) *Discovery {
	return &Discovery{
		consumer: consumer,
		cfg:      cfg,
	}
}

type Discovery struct {
	consumer api.ConsumerAPI
	cfg      *Config
}

func (d *Discovery) List(name string, opts ...discovery.Option) ([]*registry.Node, error) {
	return nil, nil
}

func (d *Discovery) Target(target string, opts ...discovery.Option) (string, error) {
	options := &discovery.Options{}
	//TODO: options转换dialOptions
	dialOptions := &DialOptions{}
	for _, o := range opts {
		o(options)
	}
	str, err := json.Marshal(dialOptions)
	if err != nil {
		return "", fmt.Errorf("marshal dialOptions error: %s", err)
	}
	endpoint := base64.URLEncoding.EncodeToString(str)
	return target + optionsPrefix + endpoint, nil
}
