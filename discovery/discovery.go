package discovery

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/polarismesh/polaris-go/api"
	"github.com/zhengheng7913/grpc-go-starter/naming/discovery"
	"github.com/zhengheng7913/grpc-go-starter/naming/registry"
	"google.golang.org/grpc"
	"strings"
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

func (d *Discovery) DialContext(ctx context.Context, target string, opts ...discovery.Option) (*grpc.ClientConn, error) {
	options := &discovery.Options{}
	//TODO: options转换dialOptions
	dialOptions := &DialOptions{}
	for _, o := range opts {
		o(options)
	}
	if !strings.HasPrefix(target, prefix) {
		return grpc.DialContext(ctx, target, dialOptions.gRPCDialOptions...)
	}
	str, err := json.Marshal(dialOptions)
	if err != nil {
		return nil, fmt.Errorf("marshal dialOptions error: %s", err)
	}
	endpoint := base64.URLEncoding.EncodeToString(str)
	target = target + optionsPrefix + endpoint
	return grpc.DialContext(ctx, target, dialOptions.gRPCDialOptions...)
}
