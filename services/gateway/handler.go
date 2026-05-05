package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	pb "github.com/bytamilan/nanayam/services/gateway/proto"
)

// FabricHandler implements the gRPC FabricServiceServer.
type FabricHandler struct {
	pb.UnimplementedFabricServiceServer
	gateway   *client.Gateway
	channel   string
	chaincode string
}

// NewFabricHandler creates a new handler with an open Fabric gateway connection.
func NewFabricHandler(gw *client.Gateway, channel, chaincode string) *FabricHandler {
	return &FabricHandler{
		gateway:   gw,
		channel:   channel,
		chaincode: chaincode,
	}
}

func (h *FabricHandler) getContract() *client.Contract {
	network := h.gateway.GetNetwork(h.channel)
	return network.GetContract(h.chaincode)
}

// CreateAsset creates a new asset on the ledger.
func (h *FabricHandler) CreateAsset(ctx context.Context, req *pb.CreateAssetRequest) (*pb.CreateAssetResponse, error) {
	contract := h.getContract()
	_, err := contract.SubmitTransaction("CreateAsset",
		req.AssetId, req.Color, fmt.Sprint(req.Size), req.Owner, fmt.Sprint(req.AppraisedValue),
	)
	if err != nil {
		log.Printf("CreateAsset error: %v", err)
		return &pb.CreateAssetResponse{Success: false, Error: err.Error()}, nil
	}
	return &pb.CreateAssetResponse{Success: true}, nil
}

// QueryAsset reads a single asset from the ledger.
func (h *FabricHandler) QueryAsset(ctx context.Context, req *pb.QueryAssetRequest) (*pb.QueryAssetResponse, error) {
	contract := h.getContract()
	result, err := contract.EvaluateTransaction("ReadAsset", req.AssetId)
	if err != nil {
		log.Printf("QueryAsset error: %v", err)
		return nil, err
	}
	return &pb.QueryAssetResponse{Data: string(result)}, nil
}

// ListAssets returns all asset IDs on the ledger.
func (h *FabricHandler) ListAssets(ctx context.Context, req *pb.ListAssetsRequest) (*pb.ListAssetsResponse, error) {
	contract := h.getContract()
	result, err := contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		log.Printf("ListAssets error: %v", err)
		return nil, err
	}

	ids, err := parseAssetIDs(result)
	if err != nil {
		return nil, err
	}
	return &pb.ListAssetsResponse{AssetIds: ids}, nil
}

// parseAssetIDs extracts asset IDs from the JSON array returned by GetAllAssets.
func parseAssetIDs(data []byte) ([]string, error) {
	var assets []map[string]interface{}
	if err := json.Unmarshal(data, &assets); err != nil {
		return nil, fmt.Errorf("unmarshal assets: %w", err)
	}

	ids := make([]string, 0, len(assets))
	for _, asset := range assets {
		if id, ok := asset["ID"].(string); ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
