package discovery

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
	"sync"
)

const (
	scheme         = "polaris"
	keyDialOptions = "options"
)

type DialOptions struct {
	gRPCDialOptions []grpc.DialOption
	Namespace       string            `json:"Namespace"`
	DstMetadata     map[string]string `json:"dst_metadata"`
	SrcMetadata     map[string]string `json:"src_metadata"`
	SrcService      string            `json:"src_service"`
	// 可选，规则路由Meta匹配前缀，用于过滤作为路由规则的gRPC Header
	HeaderPrefix []string `json:"header_prefix"`
}

func NewPolarisResolverBuilder(consumer api.ConsumerAPI) *PolarisResolverBuilder {
	return &PolarisResolverBuilder{
		consumer: consumer,
	}
}

type PolarisResolverBuilder struct {
	consumer api.ConsumerAPI
}

func resolveTarget(target resolver.Target) (*DialOptions, error) {
	options := &DialOptions{}
	if len(target.Endpoint) > 0 {
		endpoint := target.Endpoint
		if len(endpoint) > 0 {
			value, err := base64.URLEncoding.DecodeString(endpoint)
			if nil != err {
				return nil, fmt.Errorf(
					"fail to decode endpoint %s %v", endpoint, err)
			}
			if err = json.Unmarshal(value, options); nil != err {
				return nil, fmt.Errorf("fail to unmarshal options %s: %v", string(value), err)
			}
		}
	}
	return options, nil
}

func (p *PolarisResolverBuilder) Build(
	target resolver.Target,
	cc resolver.ClientConn,
	opts resolver.BuildOptions,
) (resolver.Resolver, error) {
	options, err := resolveTarget(target)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	d := &PolarisResolver{
		ctx:      ctx,
		cancel:   cancel,
		cc:       cc,
		rn:       make(chan struct{}, 1),
		consumer: p.consumer,
		options:  options,
		target:   target,
	}
	d.wg.Add(1)
	go d.watcher()
	d.ResolveNow(resolver.ResolveNowOptions{})
	return d, nil
}

func (p *PolarisResolverBuilder) Scheme() string {
	return scheme
}

type PolarisResolver struct {
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	consumer api.ConsumerAPI
	rn       chan struct{}
	cc       resolver.ClientConn
	options  *DialOptions
	target   resolver.Target
}

func (pr *PolarisResolver) ResolveNow(options resolver.ResolveNowOptions) {
	select {
	case pr.rn <- struct{}{}:
	default:

	}
}

func (pr *PolarisResolver) Close() {
	pr.cancel()
}

func (pr *PolarisResolver) watcher() {
	defer pr.wg.Done()
	var eventChan <-chan model.SubScribeEvent
	for {
		select {
		case <-pr.ctx.Done():
			return
		case <-pr.rn:
		case <-eventChan:
		}
		var state *resolver.State
		var err error
		state, err = pr.lookup()
		if err != nil {
			pr.cc.ReportError(err)
		} else {
			err := pr.cc.UpdateState(*state)
			if err != nil {
				grpclog.Errorf("fail to do update state for client conn: %v", err)
			}
			var svcKey model.ServiceKey
			svcKey, eventChan, err = pr.doWatch()
			if nil != err {
				grpclog.Errorf("fail to do watch for service %s: %v", svcKey, err)
			}
		}
	}
}

func (pr *PolarisResolver) lookup() (*resolver.State, error) {
	instancesRequest := &api.GetInstancesRequest{}
	instancesRequest.Namespace = pr.options.Namespace
	instancesRequest.Service = pr.target.Authority
	if len(pr.options.DstMetadata) > 0 {
		instancesRequest.Metadata = pr.options.DstMetadata
	}
	sourceService := &model.ServiceInfo{
		Service:   pr.options.SrcService,
		Metadata:  pr.options.DstMetadata,
		Namespace: pr.options.Namespace,
	}
	if sourceService != nil {
		// 如果在Conf中配置了SourceService，则优先使用配置
		instancesRequest.SourceService = sourceService
	}
	resp, err := pr.consumer.GetInstances(instancesRequest)
	if nil != err {
		return nil, err
	}
	state := &resolver.State{}
	for _, instance := range resp.Instances {
		state.Addresses = append(state.Addresses, resolver.Address{
			Addr:       fmt.Sprintf("%s:%d", instance.GetHost(), instance.GetPort()),
			Attributes: attributes.New(keyDialOptions, pr.options),
		})
	}
	return state, nil
}

func (pr *PolarisResolver) doWatch() (model.ServiceKey, <-chan model.SubScribeEvent, error) {
	watchRequest := &api.WatchServiceRequest{}
	watchRequest.Key = model.ServiceKey{
		Namespace: pr.options.Namespace,
		Service:   pr.target.Authority,
	}
	resp, err := pr.consumer.WatchService(watchRequest)
	if nil != err {
		return watchRequest.Key, nil, err
	}
	return watchRequest.Key, resp.EventChannel, nil
}
