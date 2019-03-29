package plugin

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"

	go_plugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/vagrant/ext/go-plugin/vagrant"
	"github.com/hashicorp/vagrant/ext/go-plugin/vagrant/plugin/proto/vagrant_caps"
	"github.com/hashicorp/vagrant/ext/go-plugin/vagrant/plugin/proto/vagrant_common"

	"github.com/LK4D4/joincontext"
)

type GuestCapabilities interface {
	vagrant.GuestCapabilities
	Meta
}

type GuestCapabilitiesPlugin struct {
	go_plugin.NetRPCUnsupportedPlugin
	Impl GuestCapabilities
}

func (g *GuestCapabilitiesPlugin) GRPCServer(broker *go_plugin.GRPCBroker, s *grpc.Server) error {
	g.Impl.Init()
	vagrant_caps.RegisterGuestCapabilitiesServer(s, &GRPCGuestCapabilitiesServer{
		Impl: g.Impl,
		GRPCIOServer: GRPCIOServer{
			Impl: g.Impl}})
	return nil
}

func (g *GuestCapabilitiesPlugin) GRPCClient(ctx context.Context, broker *go_plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	client := vagrant_caps.NewGuestCapabilitiesClient(c)
	return &GRPCGuestCapabilitiesClient{
		client:  client,
		doneCtx: ctx,
		GRPCIOClient: GRPCIOClient{
			client:  client,
			doneCtx: ctx}}, nil
}

type GRPCGuestCapabilitiesServer struct {
	GRPCIOServer
	Impl GuestCapabilities
}

func (s *GRPCGuestCapabilitiesServer) GuestCapabilities(ctx context.Context, req *vagrant_common.NullRequest) (resp *vagrant_caps.CapabilitiesResponse, err error) {
	resp = &vagrant_caps.CapabilitiesResponse{}
	var r []vagrant.SystemCapability
	n := make(chan struct{}, 1)
	go func() {
		r, err = s.Impl.GuestCapabilities()
		n <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return
	case <-n:
	}
	if err != nil {
		return
	}
	for _, cap := range r {
		rcap := &vagrant_caps.Capability{Name: cap.Name, Platform: cap.Platform}
		resp.Capabilities = append(resp.Capabilities, rcap)
	}
	return
}

func (s *GRPCGuestCapabilitiesServer) GuestCapability(ctx context.Context, req *vagrant_caps.GuestCapabilityRequest) (resp *vagrant_caps.GuestCapabilityResponse, err error) {
	resp = &vagrant_caps.GuestCapabilityResponse{}
	var args, r interface{}
	if err = json.Unmarshal([]byte(req.Arguments), &args); err != nil {
		return
	}
	machine, err := vagrant.LoadMachine(req.Machine, s.Impl)
	if err != nil {
		return
	}
	cap := &vagrant.SystemCapability{
		Name:     req.Capability.Name,
		Platform: req.Capability.Platform}
	n := make(chan struct{}, 1)
	go func() {
		r, err = s.Impl.GuestCapability(ctx, cap, args, machine)
		n <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return
	case <-n:
	}

	if err != nil {
		return
	}
	result, err := json.Marshal(r)
	if err != nil {
		return
	}
	resp.Result = string(result)
	return
}

type GRPCGuestCapabilitiesClient struct {
	GRPCCoreClient
	GRPCIOClient
	client  vagrant_caps.GuestCapabilitiesClient
	doneCtx context.Context
}

func (c *GRPCGuestCapabilitiesClient) GuestCapabilities() (caps []vagrant.SystemCapability, err error) {
	ctx := context.Background()
	jctx, _ := joincontext.Join(ctx, c.doneCtx)
	resp, err := c.client.GuestCapabilities(jctx, &vagrant_common.NullRequest{})
	if err != nil {
		return nil, handleGrpcError(err, c.doneCtx, ctx)
	}
	caps = make([]vagrant.SystemCapability, len(resp.Capabilities))
	for i := 0; i < len(resp.Capabilities); i++ {
		cap := vagrant.SystemCapability{
			Name:     resp.Capabilities[i].Name,
			Platform: resp.Capabilities[i].Platform}
		caps[i] = cap
	}
	return
}

func (c *GRPCGuestCapabilitiesClient) GuestCapability(ctx context.Context, cap *vagrant.SystemCapability, args interface{}, machine *vagrant.Machine) (result interface{}, err error) {
	a, err := json.Marshal(args)
	if err != nil {
		return
	}
	m, err := vagrant.DumpMachine(machine)
	if err != nil {
		return
	}
	jctx, _ := joincontext.Join(ctx, c.doneCtx)
	resp, err := c.client.GuestCapability(jctx, &vagrant_caps.GuestCapabilityRequest{
		Capability: &vagrant_caps.Capability{Name: cap.Name, Platform: cap.Platform},
		Machine:    m,
		Arguments:  string(a)})
	if err != nil {
		return nil, handleGrpcError(err, c.doneCtx, ctx)
	}
	err = json.Unmarshal([]byte(resp.Result), &result)
	return
}

type HostCapabilities interface {
	vagrant.HostCapabilities
	Meta
}

type HostCapabilitiesPlugin struct {
	go_plugin.NetRPCUnsupportedPlugin
	Impl HostCapabilities
}

func (h *HostCapabilitiesPlugin) GRPCServer(broker *go_plugin.GRPCBroker, s *grpc.Server) error {
	h.Impl.Init()
	vagrant_caps.RegisterHostCapabilitiesServer(s, &GRPCHostCapabilitiesServer{
		Impl: h.Impl,
		GRPCIOServer: GRPCIOServer{
			Impl: h.Impl}})
	return nil
}

func (h *HostCapabilitiesPlugin) GRPCClient(ctx context.Context, broker *go_plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	client := vagrant_caps.NewHostCapabilitiesClient(c)
	return &GRPCHostCapabilitiesClient{
		client:  client,
		doneCtx: ctx,
		GRPCIOClient: GRPCIOClient{
			client:  client,
			doneCtx: ctx}}, nil
}

type GRPCHostCapabilitiesServer struct {
	GRPCIOServer
	Impl HostCapabilities
}

func (s *GRPCHostCapabilitiesServer) HostCapabilities(ctx context.Context, req *vagrant_common.NullRequest) (resp *vagrant_caps.CapabilitiesResponse, err error) {
	resp = &vagrant_caps.CapabilitiesResponse{}
	var r []vagrant.SystemCapability
	n := make(chan struct{}, 1)
	go func() {
		r, err = s.Impl.HostCapabilities()
		n <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return
	case <-n:
	}
	if err != nil {
		return
	}
	for _, cap := range r {
		rcap := &vagrant_caps.Capability{Name: cap.Name, Platform: cap.Platform}
		resp.Capabilities = append(resp.Capabilities, rcap)
	}
	return
}

func (s *GRPCHostCapabilitiesServer) HostCapability(ctx context.Context, req *vagrant_caps.HostCapabilityRequest) (resp *vagrant_caps.HostCapabilityResponse, err error) {
	resp = &vagrant_caps.HostCapabilityResponse{}
	var args, r interface{}
	if err = json.Unmarshal([]byte(req.Arguments), &args); err != nil {
		return
	}
	env, err := vagrant.LoadEnvironment(req.Environment, s.Impl)
	if err != nil {
		return
	}
	cap := &vagrant.SystemCapability{
		Name:     req.Capability.Name,
		Platform: req.Capability.Platform}
	n := make(chan struct{}, 1)
	go func() {
		r, err = s.Impl.HostCapability(ctx, cap, args, env)
		n <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return
	case <-n:
	}
	if err != nil {
		return
	}
	result, err := json.Marshal(r)
	if err != nil {
		return
	}
	resp.Result = string(result)
	return
}

type GRPCHostCapabilitiesClient struct {
	GRPCCoreClient
	GRPCIOClient
	client  vagrant_caps.HostCapabilitiesClient
	doneCtx context.Context
}

func (c *GRPCHostCapabilitiesClient) HostCapabilities() (caps []vagrant.SystemCapability, err error) {
	ctx := context.Background()
	jctx, _ := joincontext.Join(ctx, c.doneCtx)
	resp, err := c.client.HostCapabilities(jctx, &vagrant_common.NullRequest{})
	if err != nil {
		return nil, handleGrpcError(err, c.doneCtx, ctx)
	}
	caps = make([]vagrant.SystemCapability, len(resp.Capabilities))
	for i := 0; i < len(resp.Capabilities); i++ {
		cap := vagrant.SystemCapability{
			Name:     resp.Capabilities[i].Name,
			Platform: resp.Capabilities[i].Platform}
		caps[i] = cap
	}
	return
}

func (c *GRPCHostCapabilitiesClient) HostCapability(ctx context.Context, cap *vagrant.SystemCapability, args interface{}, env *vagrant.Environment) (result interface{}, err error) {
	a, err := json.Marshal(args)
	if err != nil {
		return
	}
	e, err := vagrant.DumpEnvironment(env)
	if err != nil {
		return
	}
	jctx, _ := joincontext.Join(ctx, c.doneCtx)
	resp, err := c.client.HostCapability(jctx, &vagrant_caps.HostCapabilityRequest{
		Capability: &vagrant_caps.Capability{
			Name:     cap.Name,
			Platform: cap.Platform},
		Environment: e,
		Arguments:   string(a)})
	if err != nil {
		return nil, handleGrpcError(err, c.doneCtx, ctx)
	}
	err = json.Unmarshal([]byte(resp.Result), &result)
	return
}

type ProviderCapabilities interface {
	vagrant.ProviderCapabilities
	Meta
}

type ProviderCapabilitiesPlugin struct {
	go_plugin.NetRPCUnsupportedPlugin
	Impl ProviderCapabilities
}

func (p *ProviderCapabilitiesPlugin) GRPCServer(broker *go_plugin.GRPCBroker, s *grpc.Server) error {
	p.Impl.Init()
	vagrant_caps.RegisterProviderCapabilitiesServer(s, &GRPCProviderCapabilitiesServer{
		Impl: p.Impl,
		GRPCIOServer: GRPCIOServer{
			Impl: p.Impl}})
	return nil
}

func (p *ProviderCapabilitiesPlugin) GRPCClient(ctx context.Context, broker *go_plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	client := vagrant_caps.NewProviderCapabilitiesClient(c)
	return &GRPCProviderCapabilitiesClient{
		client:  client,
		doneCtx: ctx,
		GRPCIOClient: GRPCIOClient{
			client:  client,
			doneCtx: ctx}}, nil
}

type GRPCProviderCapabilitiesServer struct {
	GRPCIOServer
	Impl ProviderCapabilities
}

func (s *GRPCProviderCapabilitiesServer) ProviderCapabilities(ctx context.Context, req *vagrant_common.NullRequest) (resp *vagrant_caps.ProviderCapabilitiesResponse, err error) {
	resp = &vagrant_caps.ProviderCapabilitiesResponse{}
	var r []vagrant.ProviderCapability
	n := make(chan struct{}, 1)
	go func() {
		r, err = s.Impl.ProviderCapabilities()
		n <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return
	case <-n:
	}
	if err != nil {
		return
	}
	for _, cap := range r {
		rcap := &vagrant_caps.ProviderCapability{Name: cap.Name, Provider: cap.Provider}
		resp.Capabilities = append(resp.Capabilities, rcap)
	}
	return
}

func (s *GRPCProviderCapabilitiesServer) ProviderCapability(ctx context.Context, req *vagrant_caps.ProviderCapabilityRequest) (resp *vagrant_caps.ProviderCapabilityResponse, err error) {
	resp = &vagrant_caps.ProviderCapabilityResponse{}
	var args, r interface{}
	err = json.Unmarshal([]byte(req.Arguments), &args)
	if err != nil {
		return
	}
	m, err := vagrant.LoadMachine(req.Machine, s.Impl)
	if err != nil {
		return
	}
	cap := &vagrant.ProviderCapability{
		Name:     req.Capability.Name,
		Provider: req.Capability.Provider}
	n := make(chan struct{}, 1)
	go func() {
		r, err = s.Impl.ProviderCapability(ctx, cap, args, m)
		n <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return
	case <-n:
	}
	if err != nil {
		return
	}
	result, err := json.Marshal(r)
	if err != nil {
		return
	}
	resp.Result = string(result)
	return
}

type GRPCProviderCapabilitiesClient struct {
	GRPCCoreClient
	GRPCIOClient
	client  vagrant_caps.ProviderCapabilitiesClient
	doneCtx context.Context
}

func (c *GRPCProviderCapabilitiesClient) ProviderCapabilities() (caps []vagrant.ProviderCapability, err error) {
	ctx := context.Background()
	jctx, _ := joincontext.Join(ctx, c.doneCtx)
	resp, err := c.client.ProviderCapabilities(jctx, &vagrant_common.NullRequest{})
	if err != nil {
		return nil, handleGrpcError(err, c.doneCtx, ctx)
	}
	caps = make([]vagrant.ProviderCapability, len(resp.Capabilities))
	for i := 0; i < len(resp.Capabilities); i++ {
		cap := vagrant.ProviderCapability{
			Name:     resp.Capabilities[i].Name,
			Provider: resp.Capabilities[i].Provider}
		caps[i] = cap
	}
	return
}

func (c *GRPCProviderCapabilitiesClient) ProviderCapability(ctx context.Context, cap *vagrant.ProviderCapability, args interface{}, machine *vagrant.Machine) (result interface{}, err error) {
	a, err := json.Marshal(args)
	if err != nil {
		return
	}
	m, err := vagrant.DumpMachine(machine)
	if err != nil {
		return
	}
	jctx, _ := joincontext.Join(ctx, c.doneCtx)
	resp, err := c.client.ProviderCapability(jctx, &vagrant_caps.ProviderCapabilityRequest{
		Capability: &vagrant_caps.ProviderCapability{
			Name:     cap.Name,
			Provider: cap.Provider},
		Machine:   m,
		Arguments: string(a)})
	if err != nil {
		return nil, handleGrpcError(err, c.doneCtx, ctx)
	}
	err = json.Unmarshal([]byte(resp.Result), &result)
	return
}