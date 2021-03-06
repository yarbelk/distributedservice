// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ClusterManagmentClient is the client API for ClusterManagment service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClusterManagmentClient interface {
	MembershipChanges(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (ClusterManagment_MembershipChangesClient, error)
	MembershipList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*Membership, error)
}

type clusterManagmentClient struct {
	cc grpc.ClientConnInterface
}

func NewClusterManagmentClient(cc grpc.ClientConnInterface) ClusterManagmentClient {
	return &clusterManagmentClient{cc}
}

func (c *clusterManagmentClient) MembershipChanges(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (ClusterManagment_MembershipChangesClient, error) {
	stream, err := c.cc.NewStream(ctx, &ClusterManagment_ServiceDesc.Streams[0], "/proto.ClusterManagment/MembershipChanges", opts...)
	if err != nil {
		return nil, err
	}
	x := &clusterManagmentMembershipChangesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ClusterManagment_MembershipChangesClient interface {
	Recv() (*MembershipChange, error)
	grpc.ClientStream
}

type clusterManagmentMembershipChangesClient struct {
	grpc.ClientStream
}

func (x *clusterManagmentMembershipChangesClient) Recv() (*MembershipChange, error) {
	m := new(MembershipChange)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *clusterManagmentClient) MembershipList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*Membership, error) {
	out := new(Membership)
	err := c.cc.Invoke(ctx, "/proto.ClusterManagment/MembershipList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClusterManagmentServer is the server API for ClusterManagment service.
// All implementations must embed UnimplementedClusterManagmentServer
// for forward compatibility
type ClusterManagmentServer interface {
	MembershipChanges(*emptypb.Empty, ClusterManagment_MembershipChangesServer) error
	MembershipList(context.Context, *emptypb.Empty) (*Membership, error)
	mustEmbedUnimplementedClusterManagmentServer()
}

// UnimplementedClusterManagmentServer must be embedded to have forward compatible implementations.
type UnimplementedClusterManagmentServer struct {
}

func (UnimplementedClusterManagmentServer) MembershipChanges(*emptypb.Empty, ClusterManagment_MembershipChangesServer) error {
	return status.Errorf(codes.Unimplemented, "method MembershipChanges not implemented")
}
func (UnimplementedClusterManagmentServer) MembershipList(context.Context, *emptypb.Empty) (*Membership, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MembershipList not implemented")
}
func (UnimplementedClusterManagmentServer) mustEmbedUnimplementedClusterManagmentServer() {}

// UnsafeClusterManagmentServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClusterManagmentServer will
// result in compilation errors.
type UnsafeClusterManagmentServer interface {
	mustEmbedUnimplementedClusterManagmentServer()
}

func RegisterClusterManagmentServer(s grpc.ServiceRegistrar, srv ClusterManagmentServer) {
	s.RegisterService(&ClusterManagment_ServiceDesc, srv)
}

func _ClusterManagment_MembershipChanges_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(emptypb.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ClusterManagmentServer).MembershipChanges(m, &clusterManagmentMembershipChangesServer{stream})
}

type ClusterManagment_MembershipChangesServer interface {
	Send(*MembershipChange) error
	grpc.ServerStream
}

type clusterManagmentMembershipChangesServer struct {
	grpc.ServerStream
}

func (x *clusterManagmentMembershipChangesServer) Send(m *MembershipChange) error {
	return x.ServerStream.SendMsg(m)
}

func _ClusterManagment_MembershipList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterManagmentServer).MembershipList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ClusterManagment/MembershipList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterManagmentServer).MembershipList(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// ClusterManagment_ServiceDesc is the grpc.ServiceDesc for ClusterManagment service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClusterManagment_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ClusterManagment",
	HandlerType: (*ClusterManagmentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MembershipList",
			Handler:    _ClusterManagment_MembershipList_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "MembershipChanges",
			Handler:       _ClusterManagment_MembershipChanges_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "managment.proto",
}
