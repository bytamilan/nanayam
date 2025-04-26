package main

import (
	"context"
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	pb "github.com/yourorg/fabric-gateway/proto"
)

// FabricHandler implements pb.FabricServiceServer
type FabricHandler struct{}

func (h *FabricHandler) CreateAsset(ctx context.Context, req *pb.CreateAssetRequest) (*pb.CreateAssetResponse, error) {
	gw, err := client.Connect( /* load CCP, identity, signer */ )
	if err != nil {
		return nil, err
	}
	network := gw.GetNetwork("mychannel")
	contract := network.GetContract("basic")
	_, err = contract.SubmitTransaction("CreateAsset", req.AssetId, req.Color, fmt.Sprint(req.Size))
	return &pb.CreateAssetResponse{Success: err == nil}, err
}

func (h *FabricHandler) QueryAsset(ctx context.Context, req *pb.QueryAssetRequest) (*pb.QueryAssetResponse, error) {
	gw, _ := client.Connect( /*...*/ )
	network := gw.GetNetwork("mychannel")
	contract := network.GetContract("basic")
	result, err := contract.EvaluateTransaction("ReadAsset", req.AssetId)
	if err != nil {
		return nil, err
	}
	return &pb.QueryAssetResponse{Data: string(result)}, nil
}

func (h *FabricHandler) ListAssets(ctx context.Context, req *pb.ListAssetsRequest) (*pb.ListAssetsResponse, error) {
	gw, _ := client.Connect( /*...*/ )
	network := gw.GetNetwork("mychannel")
	contract := network.GetContract("basic")
	result, _ := contract.EvaluateTransaction("GetAllAssets")
	// parse JSON, extract IDs
	ids := parseIDs(result)
	return &pb.ListAssetsResponse{AssetIds: ids}, nil
}
