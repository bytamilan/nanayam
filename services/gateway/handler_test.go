package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/bytamilan/nanayam/services/gateway/proto"
)

// mockHandler implements pb.FabricServiceServer for unit testing.
type mockHandler struct{}

func (m *mockHandler) CreateAsset(ctx context.Context, req *pb.CreateAssetRequest) (*pb.CreateAssetResponse, error) {
	return &pb.CreateAssetResponse{Success: true}, nil
}
func (m *mockHandler) QueryAsset(ctx context.Context, req *pb.QueryAssetRequest) (*pb.QueryAssetResponse, error) {
	return &pb.QueryAssetResponse{Data: `{"ID":"asset1"}`}, nil
}
func (m *mockHandler) ListAssets(ctx context.Context, req *pb.ListAssetsRequest) (*pb.ListAssetsResponse, error) {
	return &pb.ListAssetsResponse{AssetIds: []string{"asset1", "asset2"}}, nil
}
func (m *mockHandler) SubmitComplaint(ctx context.Context, req *pb.SubmitComplaintRequest) (*pb.SubmitComplaintResponse, error) {
	return &pb.SubmitComplaintResponse{Success: true}, nil
}
func (m *mockHandler) UpdateComplaint(ctx context.Context, req *pb.UpdateComplaintRequest) (*pb.UpdateComplaintResponse, error) {
	return &pb.UpdateComplaintResponse{Success: true}, nil
}
func (m *mockHandler) QueryComplaint(ctx context.Context, req *pb.QueryComplaintRequest) (*pb.QueryComplaintResponse, error) {
	return &pb.QueryComplaintResponse{Data: `{"complaintId":"COMP-001","status":"Submitted"}`}, nil
}
func (m *mockHandler) ListComplaints(ctx context.Context, req *pb.ListComplaintsRequest) (*pb.ListComplaintsResponse, error) {
	return &pb.ListComplaintsResponse{ComplaintIds: []string{"COMP-001", "COMP-002"}}, nil
}
func (m *mockHandler) GetComplaintHistory(ctx context.Context, req *pb.GetComplaintHistoryRequest) (*pb.GetComplaintHistoryResponse, error) {
	return &pb.GetComplaintHistoryResponse{Data: `[{"txId":"tx1"}]`}, nil
}

func TestHealthEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	registerRESTHandlers(mux, &mockHandler{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", body)
	}
}

func TestCreateAssetEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	registerRESTHandlers(mux, &mockHandler{})

	payload := map[string]interface{}{
		"assetId": "asset1", "color": "blue", "size": 10,
		"owner": "Alice", "appraisedValue": 100,
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/CreateAsset", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var body pb.CreateAssetResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success {
		t.Fatalf("expected success, got %v", body)
	}
}

func TestSubmitComplaintEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	registerRESTHandlers(mux, &mockHandler{})

	payload := map[string]interface{}{
		"complaintId":     "COMP-001",
		"category":        "bribery",
		"citizenHash":     "sha256:abc",
		"descriptionHash": "sha256:def",
		"attachmentsRef":  "ipfs:QmTest",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/SubmitComplaint", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var body pb.SubmitComplaintResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success {
		t.Fatalf("expected success, got %v", body)
	}
}

func TestUpdateComplaintEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	registerRESTHandlers(mux, &mockHandler{})

	payload := map[string]interface{}{
		"complaintId": "COMP-001",
		"action":      "acknowledge",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/UpdateComplaint", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var body pb.UpdateComplaintResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success {
		t.Fatalf("expected success, got %v", body)
	}
}

func TestQueryComplaintEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	registerRESTHandlers(mux, &mockHandler{})

	req := httptest.NewRequest(http.MethodGet, "/v1/QueryComplaint?complaintId=COMP-001", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var body pb.QueryComplaintResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data == "" {
		t.Fatalf("expected data, got empty")
	}
}

func TestListComplaintsEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	registerRESTHandlers(mux, &mockHandler{})

	req := httptest.NewRequest(http.MethodGet, "/v1/ListComplaints", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var body pb.ListComplaintsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.ComplaintIds) == 0 {
		t.Fatalf("expected complaint ids, got none")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	registerRESTHandlers(mux, &mockHandler{})

	req := httptest.NewRequest(http.MethodDelete, "/v1/SubmitComplaint", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
