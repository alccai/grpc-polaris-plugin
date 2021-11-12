package grpc_registry

import (
	"fmt"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/model"
	"github.com/zhengheng7913/grpc-go-starter/naming/registry"
	"google.golang.org/grpc/grpclog"
	"net"
	"strconv"
	"time"
)

const (
	DefaultHeartBeat = 5
	DefaultWeight    = 100
	DefaultTTL       = 5
)

type Registry struct {
	provider api.ProviderAPI
	cfg      *Config
	host     string
	port     int
}

func NewRegistry(provider api.ProviderAPI, cfg *Config) (*Registry, error) {
	return newRegistry(provider, cfg)
}

func newRegistry(provider api.ProviderAPI, cfg *Config) (*Registry, error) {
	if cfg.HeartBeat == 0 {
		cfg.HeartBeat = DefaultHeartBeat
	}
	if cfg.Weight == 0 {
		cfg.Weight = DefaultWeight
	}
	if cfg.TTL == 0 {
		cfg.TTL = DefaultTTL
	}
	return &Registry{
		provider: provider,
		cfg:      cfg,
	}, nil
}

func (r *Registry) Register(_ string, opt ...registry.Option) error {
	opts := &registry.Options{}
	for _, o := range opt {
		o(opts)
	}
	address := opts.Address
	if r.cfg.BindAddress != "" {
		address = r.cfg.BindAddress
	}
	host, portRaw, _ := net.SplitHostPort(address)
	port, _ := strconv.ParseInt(portRaw, 10, 64)
	r.host = host
	r.port = int(port)
	if r.cfg.EnableRegister {
		if err := r.register(); err != nil {
			return err
		}
	}
	go r.heartBeat()
	return nil
}

func (r *Registry) register() error {
	req := &api.InstanceRegisterRequest{
		InstanceRegisterRequest: model.InstanceRegisterRequest{
			Namespace:    r.cfg.Namespace,
			Service:      r.cfg.ServiceName,
			Host:         r.host,
			Port:         r.port,
			ServiceToken: r.cfg.ServiceToken,
			Weight:       &r.cfg.Weight,
			Metadata:     r.cfg.Metadata,
			Protocol:     &r.cfg.Protocol,
			Version:      &r.cfg.Version,
		},
	}
	if !r.cfg.DisableHealthCheck {
		req.SetTTL(r.cfg.TTL)
	}
	resp, err := r.provider.Register(req)
	if err != nil {
		return fmt.Errorf("fail to Register instance, err is %v", err)
	}
	grpclog.Info("success to register instance1, id is %s\n", resp.InstanceID)
	r.cfg.InstanceID = resp.InstanceID
	return nil
}

func (r *Registry) heartBeat() {
	tick := time.Second * time.Duration(r.cfg.HeartBeat)
	go func() {
		for {
			req := &api.InstanceHeartbeatRequest{
				InstanceHeartbeatRequest: model.InstanceHeartbeatRequest{
					Service:      r.cfg.ServiceName,
					ServiceToken: r.cfg.ServiceToken,
					Namespace:    r.cfg.Namespace,
					InstanceID:   r.cfg.InstanceID,
					Host:         r.host,
					Port:         r.port,
				},
			}
			if err := r.provider.Heartbeat(req); err != nil {
				grpclog.Error("heartbeat report err: %v\n", err)
			} else {
				grpclog.Info("heart beat success")
			}
			time.Sleep(tick)
		}
	}()
}

// Deregister 反注册
func (r *Registry) Deregister(_ string) error {
	if !r.cfg.EnableRegister {
		return nil
	}
	req := &api.InstanceDeRegisterRequest{
		InstanceDeRegisterRequest: model.InstanceDeRegisterRequest{
			Service:      r.cfg.ServiceName,
			Namespace:    r.cfg.Namespace,
			InstanceID:   r.cfg.InstanceID,
			ServiceToken: r.cfg.ServiceToken,
			Host:         r.host,
			Port:         r.port,
		},
	}
	if err := r.provider.Deregister(req); err != nil {
		return fmt.Errorf("deregister error: %s", err.Error())
	}
	return nil
}
