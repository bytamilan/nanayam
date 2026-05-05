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

// ---------------------------------------------------------------------------
// Asset Methods (original)
// ---------------------------------------------------------------------------

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

func (h *FabricHandler) QueryAsset(ctx context.Context, req *pb.QueryAssetRequest) (*pb.QueryAssetResponse, error) {
	contract := h.getContract()
	result, err := contract.EvaluateTransaction("ReadAsset", req.AssetId)
	if err != nil {
		log.Printf("QueryAsset error: %v", err)
		return nil, err
	}
	return &pb.QueryAssetResponse{Data: string(result)}, nil
}

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

// ---------------------------------------------------------------------------
// Complaint Methods
// ---------------------------------------------------------------------------

func (h *FabricHandler) SubmitComplaint(ctx context.Context, req *pb.SubmitComplaintRequest) (*pb.SubmitComplaintResponse, error) {
	contract := h.getContract()
	_, err := contract.SubmitTransaction("SubmitComplaint",
		req.ComplaintID, req.Category, req.CitizenHash, req.DescriptionHash, req.AttachmentsRef,
	)
	if err != nil {
		log.Printf("SubmitComplaint error: %v", err)
		return &pb.SubmitComplaintResponse{Success: false, Error: err.Error()}, nil
	}
	return &pb.SubmitComplaintResponse{Success: true}, nil
}

func (h *FabricHandler) UpdateComplaint(ctx context.Context, req *pb.UpdateComplaintRequest) (*pb.UpdateComplaintResponse, error) {
	contract := h.getContract()
	var err error

	switch req.Action {
	case "acknowledge":
		_, err = contract.SubmitTransaction("AcknowledgeComplaint", req.ComplaintID)
	case "assign":
		_, err = contract.SubmitTransaction("AssignInvestigator", req.ComplaintID, req.Value)
	case "updateStatus":
		_, err = contract.SubmitTransaction("UpdateStatus", req.ComplaintID, req.Value)
	case "addEvidence":
		_, err = contract.SubmitTransaction("AddEvidence", req.ComplaintID, req.Value)
	case "requestClosure":
		_, err = contract.SubmitTransaction("RequestClosure", req.ComplaintID, req.Value)
	case "approveClosure":
		_, err = contract.SubmitTransaction("ApproveClosure", req.ComplaintID)
	case "reject":
		_, err = contract.SubmitTransaction("RejectComplaint", req.ComplaintID, req.Value)
	default:
		return &pb.UpdateComplaintResponse{Success: false, Error: "unknown action: " + req.Action}, nil
	}

	if err != nil {
		log.Printf("UpdateComplaint error (action=%s): %v", req.Action, err)
		return &pb.UpdateComplaintResponse{Success: false, Error: err.Error()}, nil
	}
	return &pb.UpdateComplaintResponse{Success: true}, nil
}

func (h *FabricHandler) QueryComplaint(ctx context.Context, req *pb.QueryComplaintRequest) (*pb.QueryComplaintResponse, error) {
	contract := h.getContract()
	result, err := contract.EvaluateTransaction("GetComplaint", req.ComplaintID)
	if err != nil {
		log.Printf("QueryComplaint error: %v", err)
		return nil, err
	}
	return &pb.QueryComplaintResponse{Data: string(result)}, nil
}

func (h *FabricHandler) ListComplaints(ctx context.Context, req *pb.ListComplaintsRequest) (*pb.ListComplaintsResponse, error) {
	contract := h.getContract()
	result, err := contract.EvaluateTransaction("GetAllComplaints")
	if err != nil {
		log.Printf("ListComplaints error: %v", err)
		return nil, err
	}

	ids, err := parseComplaintIDs(result)
	if err != nil {
		return nil, err
	}
	return &pb.ListComplaintsResponse{ComplaintIds: ids}, nil
}

func (h *FabricHandler) GetComplaintHistory(ctx context.Context, req *pb.GetComplaintHistoryRequest) (*pb.GetComplaintHistoryResponse, error) {
	contract := h.getContract()
	result, err := contract.EvaluateTransaction("GetComplaintHistory", req.ComplaintID)
	if err != nil {
		log.Printf("GetComplaintHistory error: %v", err)
		return nil, err
	}
	return &pb.GetComplaintHistoryResponse{Data: string(result)}, nil
}

// ---------------------------------------------------------------------------
// JSON Helpers
// ---------------------------------------------------------------------------

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

func parseComplaintIDs(data []byte) ([]string, error) {
	var complaints []map[string]interface{}
	if err := json.Unmarshal(data, &complaints); err != nil {
		return nil, fmt.Errorf("unmarshal complaints: %w", err)
	}

	ids := make([]string, 0, len(complaints))
	for _, c := range complaints {
		if id, ok := c["complaintId"].(string); ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
