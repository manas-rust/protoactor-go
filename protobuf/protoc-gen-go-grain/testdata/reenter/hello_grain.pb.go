// Code generated by protoc-gen-grain. DO NOT EDIT.
// versions:
//  protoc-gen-grain v0.5.0
//  protoc           v4.25.0
// source: testdata/reenter/hello.proto

package hello

import (
	fmt "fmt"
	actor "github.com/asynkron/protoactor-go/actor"
	cluster "github.com/asynkron/protoactor-go/cluster"
	proto "google.golang.org/protobuf/proto"
	slog "log/slog"
	time "time"
)

var xHelloFactory func() Hello

// HelloFactory produces a Hello
func HelloFactory(factory func() Hello) {
	xHelloFactory = factory
}

// GetHelloGrainClient instantiates a new HelloGrainClient with given Identity
func GetHelloGrainClient(c *cluster.Cluster, id string) *HelloGrainClient {
	if c == nil {
		panic(fmt.Errorf("nil cluster instance"))
	}
	if id == "" {
		panic(fmt.Errorf("empty id"))
	}
	return &HelloGrainClient{Identity: id, cluster: c}
}

// GetHelloKind instantiates a new cluster.Kind for Hello
func GetHelloKind(opts ...actor.PropsOption) *cluster.Kind {
	props := actor.PropsFromProducer(func() actor.Actor {
		return &HelloActor{
			Timeout: 60 * time.Second,
		}
	}, opts...)
	kind := cluster.NewKind("Hello", props)
	return kind
}

// GetHelloKind instantiates a new cluster.Kind for Hello
func NewHelloKind(factory func() Hello, timeout time.Duration, opts ...actor.PropsOption) *cluster.Kind {
	xHelloFactory = factory
	props := actor.PropsFromProducer(func() actor.Actor {
		return &HelloActor{
			Timeout: timeout,
		}
	}, opts...)
	kind := cluster.NewKind("Hello", props)
	return kind
}

// Hello interfaces the services available to the Hello
type Hello interface {
	Init(ctx cluster.GrainContext)
	Terminate(ctx cluster.GrainContext)
	ReceiveDefault(ctx cluster.GrainContext)
	SayHello(req *SayHelloRequest, respond func(*SayHelloResponse), onError func(error), ctx cluster.GrainContext) error
	Dowork(req *DoworkRequest, ctx cluster.GrainContext) (*DoworkResponse, error)
}

// HelloGrainClient holds the base data for the HelloGrain
type HelloGrainClient struct {
	Identity string
	cluster  *cluster.Cluster
}

// SayHello requests the execution on to the cluster with CallOptions
func (g *HelloGrainClient) SayHello(r *SayHelloRequest, opts ...cluster.GrainCallOption) (*SayHelloResponse, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 0, MessageData: bytes}
	resp, err := g.cluster.Request(g.Identity, "Hello", reqMsg, opts...)
	if err != nil {
		return nil, fmt.Errorf("error request: %w", err)
	}
	switch msg := resp.(type) {
	case *SayHelloResponse:
		return msg, nil
	case *cluster.GrainErrorResponse:
		if msg == nil {
			return nil, nil
		}
		return nil, msg
	default:
		return nil, fmt.Errorf("unknown response type %T", resp)
	}
}

// DoworkFuture return a future for the execution of Dowork on the cluster
func (g *HelloGrainClient) DoworkFuture(r *DoworkRequest, opts ...cluster.GrainCallOption) (*actor.Future, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}

	reqMsg := &cluster.GrainRequest{MethodIndex: 1, MessageData: bytes}
	f, err := g.cluster.RequestFuture(g.Identity, "Hello", reqMsg, opts...)
	if err != nil {
		return nil, fmt.Errorf("error request future: %w", err)
	}

	return f, nil
}

// Dowork requests the execution on to the cluster with CallOptions
func (g *HelloGrainClient) Dowork(r *DoworkRequest, opts ...cluster.GrainCallOption) (*DoworkResponse, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 1, MessageData: bytes}
	resp, err := g.cluster.Request(g.Identity, "Hello", reqMsg, opts...)
	if err != nil {
		return nil, fmt.Errorf("error request: %w", err)
	}
	switch msg := resp.(type) {
	case *DoworkResponse:
		return msg, nil
	case *cluster.GrainErrorResponse:
		if msg == nil {
			return nil, nil
		}
		return nil, msg
	default:
		return nil, fmt.Errorf("unknown response type %T", resp)
	}
}

// HelloActor represents the actor structure
type HelloActor struct {
	ctx     cluster.GrainContext
	inner   Hello
	Timeout time.Duration
}

// Receive ensures the lifecycle of the actor for the received message
func (a *HelloActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started: //pass
	case *cluster.ClusterInit:
		a.ctx = cluster.NewGrainContext(ctx, msg.Identity, msg.Cluster)
		a.inner = xHelloFactory()
		a.inner.Init(a.ctx)

		if a.Timeout > 0 {
			ctx.SetReceiveTimeout(a.Timeout)
		}
	case *actor.ReceiveTimeout:
		ctx.Poison(ctx.Self())
	case *actor.Stopped:
		a.inner.Terminate(a.ctx)
	case actor.AutoReceiveMessage: // pass
	case actor.SystemMessage: // pass

	case *cluster.GrainRequest:
		switch msg.MethodIndex {
		case 0:
			req := &SayHelloRequest{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				ctx.Logger().Error("[Grain] SayHello(SayHelloRequest) proto.Unmarshal failed.", slog.Any("error", err))
				resp := cluster.NewGrainErrorResponse(cluster.ErrorReason_INVALID_ARGUMENT, err.Error()).
					WithMetadata(map[string]string{
						"argument": req.String(),
					})
				ctx.Respond(resp)
				return
			}
			err = a.inner.SayHello(req, respond[*SayHelloResponse](a.ctx), a.onError, a.ctx)
			if err != nil {
				resp := cluster.FromError(err)
				ctx.Respond(resp)
				return
			}
		case 1:
			req := &DoworkRequest{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				ctx.Logger().Error("[Grain] Dowork(DoworkRequest) proto.Unmarshal failed.", slog.Any("error", err))
				resp := cluster.NewGrainErrorResponse(cluster.ErrorReason_INVALID_ARGUMENT, err.Error()).
					WithMetadata(map[string]string{
						"argument": req.String(),
					})
				ctx.Respond(resp)
				return
			}

			r0, err := a.inner.Dowork(req, a.ctx)
			if err != nil {
				resp := cluster.FromError(err)
				ctx.Respond(resp)
				return
			}
			ctx.Respond(r0)
		}
	default:
		a.inner.ReceiveDefault(a.ctx)
	}
}

// onError should be used in ctx.ReenterAfter
// you can just return error in reenterable method for other errors
func (a *HelloActor) onError(err error) {
	resp := cluster.FromError(err)
	a.ctx.Respond(resp)
}

func respond[T proto.Message](ctx cluster.GrainContext) func(T) {
	return func(resp T) {
		ctx.Respond(resp)
	}
}
