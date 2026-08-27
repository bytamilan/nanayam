package proto

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// Asset Messages (original)
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
// Complaint Messages
// ---------------------------------------------------------------------------

type SubmitComplaintRequest struct {
	ComplaintID     string `json:"complaintId"`
	Category        string `json:"category"`
	CitizenHash     string `json:"citizenHash"`
	DescriptionHash string `json:"descriptionHash"`
	AttachmentsRef  string `json:"attachmentsRef"`
}

type SubmitComplaintResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type UpdateComplaintRequest struct {
	ComplaintID string `json:"complaintId"`
	Action      string `json:"action"` // acknowledge, assign, updateStatus, addEvidence, requestClosure, approveClosure, reject
	Value       string `json:"value"`  // new status / dept / reason / attachment
}

type UpdateComplaintResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type QueryComplaintRequest struct {
	ComplaintID string `json:"complaintId"`
}

type QueryComplaintResponse struct {
	Data string `json:"data"`
}

type ListComplaintsRequest struct{}

type ListComplaintsResponse struct {
	ComplaintIds []string `json:"complaintIds"`
}

type GetComplaintHistoryRequest struct {
	ComplaintID string `json:"complaintId"`
}

type GetComplaintHistoryResponse struct {
	Data string `json:"data"`
}

// ---------------------------------------------------------------------------
// gRPC Service Interface & Registration
// ---------------------------------------------------------------------------

type FabricServiceServer interface {
	// Asset methods
	CreateAsset(context.Context, *CreateAssetRequest) (*CreateAssetResponse, error)
	QueryAsset(context.Context, *QueryAssetRequest) (*QueryAssetResponse, error)
	ListAssets(context.Context, *ListAssetsRequest) (*ListAssetsResponse, error)

	// Complaint methods
	SubmitComplaint(context.Context, *SubmitComplaintRequest) (*SubmitComplaintResponse, error)
	UpdateComplaint(context.Context, *UpdateComplaintRequest) (*UpdateComplaintResponse, error)
	QueryComplaint(context.Context, *QueryComplaintRequest) (*QueryComplaintResponse, error)
	ListComplaints(context.Context, *ListComplaintsRequest) (*ListComplaintsResponse, error)
	GetComplaintHistory(context.Context, *GetComplaintHistoryRequest) (*GetComplaintHistoryResponse, error)
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
func (UnimplementedFabricServiceServer) SubmitComplaint(context.Context, *SubmitComplaintRequest) (*SubmitComplaintResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitComplaint not implemented")
}
func (UnimplementedFabricServiceServer) UpdateComplaint(context.Context, *UpdateComplaintRequest) (*UpdateComplaintResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateComplaint not implemented")
}
func (UnimplementedFabricServiceServer) QueryComplaint(context.Context, *QueryComplaintRequest) (*QueryComplaintResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryComplaint not implemented")
}
func (UnimplementedFabricServiceServer) ListComplaints(context.Context, *ListComplaintsRequest) (*ListComplaintsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListComplaints not implemented")
}
func (UnimplementedFabricServiceServer) GetComplaintHistory(context.Context, *GetComplaintHistoryRequest) (*GetComplaintHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetComplaintHistory not implemented")
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

func _FabricService_SubmitComplaint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitComplaintRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FabricServiceServer).SubmitComplaint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fabric.FabricService/SubmitComplaint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FabricServiceServer).SubmitComplaint(ctx, req.(*SubmitComplaintRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FabricService_UpdateComplaint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateComplaintRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FabricServiceServer).UpdateComplaint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fabric.FabricService/UpdateComplaint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FabricServiceServer).UpdateComplaint(ctx, req.(*UpdateComplaintRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FabricService_QueryComplaint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryComplaintRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FabricServiceServer).QueryComplaint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fabric.FabricService/QueryComplaint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FabricServiceServer).QueryComplaint(ctx, req.(*QueryComplaintRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FabricService_ListComplaints_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListComplaintsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FabricServiceServer).ListComplaints(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fabric.FabricService/ListComplaints",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FabricServiceServer).ListComplaints(ctx, req.(*ListComplaintsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FabricService_GetComplaintHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetComplaintHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FabricServiceServer).GetComplaintHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fabric.FabricService/GetComplaintHistory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FabricServiceServer).GetComplaintHistory(ctx, req.(*GetComplaintHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _FabricService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "fabric.FabricService",
	HandlerType: (*FabricServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateAsset", Handler: _FabricService_CreateAsset_Handler},
		{MethodName: "QueryAsset", Handler: _FabricService_QueryAsset_Handler},
		{MethodName: "ListAssets", Handler: _FabricService_ListAssets_Handler},
		{MethodName: "SubmitComplaint", Handler: _FabricService_SubmitComplaint_Handler},
		{MethodName: "UpdateComplaint", Handler: _FabricService_UpdateComplaint_Handler},
		{MethodName: "QueryComplaint", Handler: _FabricService_QueryComplaint_Handler},
		{MethodName: "ListComplaints", Handler: _FabricService_ListComplaints_Handler},
		{MethodName: "GetComplaintHistory", Handler: _FabricService_GetComplaintHistory_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fabric.proto",
}
