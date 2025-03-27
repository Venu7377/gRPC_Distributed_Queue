// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.1
// source: proto/queue.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	TaskCoordinator_SubmitTask_FullMethodName     = "/queue.TaskCoordinator/SubmitTask"
	TaskCoordinator_RegisterWorker_FullMethodName = "/queue.TaskCoordinator/RegisterWorker"
	TaskCoordinator_TaskCompleted_FullMethodName  = "/queue.TaskCoordinator/TaskCompleted"
)

// TaskCoordinatorClient is the client API for TaskCoordinator service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TaskCoordinatorClient interface {
	// Client submits a task to the coordinator
	SubmitTask(ctx context.Context, in *TaskRequest, opts ...grpc.CallOption) (*TaskResponse, error)
	// Workers register and receive tasks via a stream(bidi)
	RegisterWorker(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[WorkerStatus, TaskRequest], error)
	// Worker reports task completion
	TaskCompleted(ctx context.Context, in *TaskCompletion, opts ...grpc.CallOption) (*TaskResponse, error)
}

type taskCoordinatorClient struct {
	cc grpc.ClientConnInterface
}

func NewTaskCoordinatorClient(cc grpc.ClientConnInterface) TaskCoordinatorClient {
	return &taskCoordinatorClient{cc}
}

func (c *taskCoordinatorClient) SubmitTask(ctx context.Context, in *TaskRequest, opts ...grpc.CallOption) (*TaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TaskResponse)
	err := c.cc.Invoke(ctx, TaskCoordinator_SubmitTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskCoordinatorClient) RegisterWorker(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[WorkerStatus, TaskRequest], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &TaskCoordinator_ServiceDesc.Streams[0], TaskCoordinator_RegisterWorker_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[WorkerStatus, TaskRequest]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type TaskCoordinator_RegisterWorkerClient = grpc.BidiStreamingClient[WorkerStatus, TaskRequest]

func (c *taskCoordinatorClient) TaskCompleted(ctx context.Context, in *TaskCompletion, opts ...grpc.CallOption) (*TaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TaskResponse)
	err := c.cc.Invoke(ctx, TaskCoordinator_TaskCompleted_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TaskCoordinatorServer is the server API for TaskCoordinator service.
// All implementations must embed UnimplementedTaskCoordinatorServer
// for forward compatibility.
type TaskCoordinatorServer interface {
	// Client submits a task to the coordinator
	SubmitTask(context.Context, *TaskRequest) (*TaskResponse, error)
	// Workers register and receive tasks via a stream(bidi)
	RegisterWorker(grpc.BidiStreamingServer[WorkerStatus, TaskRequest]) error
	// Worker reports task completion
	TaskCompleted(context.Context, *TaskCompletion) (*TaskResponse, error)
	mustEmbedUnimplementedTaskCoordinatorServer()
}

// UnimplementedTaskCoordinatorServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTaskCoordinatorServer struct{}

func (UnimplementedTaskCoordinatorServer) SubmitTask(context.Context, *TaskRequest) (*TaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitTask not implemented")
}
func (UnimplementedTaskCoordinatorServer) RegisterWorker(grpc.BidiStreamingServer[WorkerStatus, TaskRequest]) error {
	return status.Errorf(codes.Unimplemented, "method RegisterWorker not implemented")
}
func (UnimplementedTaskCoordinatorServer) TaskCompleted(context.Context, *TaskCompletion) (*TaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TaskCompleted not implemented")
}
func (UnimplementedTaskCoordinatorServer) mustEmbedUnimplementedTaskCoordinatorServer() {}
func (UnimplementedTaskCoordinatorServer) testEmbeddedByValue()                         {}

// UnsafeTaskCoordinatorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TaskCoordinatorServer will
// result in compilation errors.
type UnsafeTaskCoordinatorServer interface {
	mustEmbedUnimplementedTaskCoordinatorServer()
}

func RegisterTaskCoordinatorServer(s grpc.ServiceRegistrar, srv TaskCoordinatorServer) {
	// If the following call pancis, it indicates UnimplementedTaskCoordinatorServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TaskCoordinator_ServiceDesc, srv)
}

func _TaskCoordinator_SubmitTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskCoordinatorServer).SubmitTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaskCoordinator_SubmitTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskCoordinatorServer).SubmitTask(ctx, req.(*TaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskCoordinator_RegisterWorker_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TaskCoordinatorServer).RegisterWorker(&grpc.GenericServerStream[WorkerStatus, TaskRequest]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type TaskCoordinator_RegisterWorkerServer = grpc.BidiStreamingServer[WorkerStatus, TaskRequest]

func _TaskCoordinator_TaskCompleted_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskCompletion)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskCoordinatorServer).TaskCompleted(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaskCoordinator_TaskCompleted_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskCoordinatorServer).TaskCompleted(ctx, req.(*TaskCompletion))
	}
	return interceptor(ctx, in, info, handler)
}

// TaskCoordinator_ServiceDesc is the grpc.ServiceDesc for TaskCoordinator service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TaskCoordinator_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "queue.TaskCoordinator",
	HandlerType: (*TaskCoordinatorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitTask",
			Handler:    _TaskCoordinator_SubmitTask_Handler,
		},
		{
			MethodName: "TaskCompleted",
			Handler:    _TaskCoordinator_TaskCompleted_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "RegisterWorker",
			Handler:       _TaskCoordinator_RegisterWorker_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proto/queue.proto",
}
