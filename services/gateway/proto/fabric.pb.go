package proto

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// Messages
// ---------------------------------------------------------------------------

type CreateAssetRequest struct {
	AssetId        string `json:"assetId"`
	Color          string `json:"color"`
	Size           int32  `json:"size"`
	Owner          string `json:"owner"`
	AppraisedValue int32  `json:"appraisedValue"`
}

type CreateAssetResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type QueryAssetRequest struct {
	AssetId string `json:"assetId"`
}

type QueryAssetResponse struct {
	Data string `json:"data"`
}

type ListAssetsRequest struct{}

type ListAssetsResponse struct {
	AssetIds []string `json:"assetIds"`
}

// ---------------------------------------------------------------------------
// gRPC Service Interface & Registration
// ---------------------------------------------------------------------------

type FabricServiceServer interface {
	CreateAsset(context.Context, *CreateAssetRequest) (*CreateAssetResponse, error)
	QueryAsset(context.Context, *QueryAssetRequest) (*QueryAssetResponse, error)
	ListAssets(context.Context, *ListAssetsRequest) (*ListAssetsResponse, error)
}

// UnimplementedFabricServiceServer can be embedded to have forward compatible implementations.
type UnimplementedFabricServiceServer struct{}

func (UnimplementedFabricServiceServer) CreateAsset(context.Context, *CreateAssetRequest) (*CreateAssetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAsset not implemented")
}
func (UnimplementedFabricServiceServer) QueryAsset(context.Context, *QueryAssetRequest) (*QueryAssetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryAsset not implemented")
}
func (UnimplementedFabricServiceServer) ListAssets(context.Context, *ListAssetsRequest) (*ListAssetsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAssets not implemented")
}

func RegisterFabricServiceServer(s *grpc.Server, srv FabricServiceServer) {
	s.RegisterService(&_FabricService_serviceDesc, srv)
}

func _FabricService_CreateAsset_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAssetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FabricServiceServer).CreateAsset(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fabric.FabricService/CreateAsset",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FabricServiceServer).CreateAsset(ctx, req.(*CreateAssetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FabricService_QueryAsset_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAssetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FabricServiceServer).QueryAsset(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fabric.FabricService/QueryAsset",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FabricServiceServer).QueryAsset(ctx, req.(*QueryAssetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FabricService_ListAssets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAssetsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FabricServiceServer).ListAssets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fabric.FabricService/ListAssets",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FabricServiceServer).ListAssets(ctx, req.(*ListAssetsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _FabricService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "fabric.FabricService",
	HandlerType: (*FabricServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateAsset",
			Handler:    _FabricService_CreateAsset_Handler,
		},
		{
			MethodName: "QueryAsset",
			Handler:    _FabricService_QueryAsset_Handler,
		},
		{
			MethodName: "ListAssets",
			Handler:    _FabricService_ListAssets_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fabric.proto",
}
