package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"google.golang.org/protobuf/proto"

	pb "github.com/bytamilan/nanayam/services/gateway/proto"
)

// RESTServer holds the gRPC handler, auth store, and gateway config for HTTP serving.
type RESTServer struct {
	handler   pb.FabricServiceServer
	authStore *AuthStore
	cfg       *GatewayConfig
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func (s *RESTServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		_, err := s.authStore.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ---------------------------------------------------------------------------
// Registration / Login / Me
// ---------------------------------------------------------------------------

func (s *RESTServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.authStore.IsSignupEnabled() {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "registration is disabled"})
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Org      string `json:"org"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	user, err := s.authStore.Register(req.Username, req.Password, req.Org, "user")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (s *RESTServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	token, err := s.authStore.Login(req.Username, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *RESTServer) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	authHeader := r.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	claims, err := s.authStore.ValidateToken(parts[1])
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	username, _ := claims["usr"].(string)
	user, ok := s.authStore.GetUserByUsername(username)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// ---------------------------------------------------------------------------
// Channel Info
// ---------------------------------------------------------------------------

func (s *RESTServer) handleChannelInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	info := map[string]interface{}{
		"channel":   s.cfg.ChannelName,
		"chaincode": s.cfg.ChaincodeName,
		"mspId":     s.cfg.MSP_ID,
		"organizations": []map[string]interface{}{
			{"mspId": "ACBMSP", "name": "ACB", "peers": []string{"peer0.acb.nanayam.com:7051"}, "role": "Acknowledge, assign, investigate, request closure"},
			{"mspId": "DeptMSP", "name": "Department", "peers": []string{"peer0.dept.nanayam.com:9051"}, "role": "Update status, add evidence"},
			{"mspId": "OversightMSP", "name": "Oversight", "peers": []string{"peer0.oversight.nanayam.com:10051"}, "role": "Co-endorse closure"},
			{"mspId": "JudiciaryMSP", "name": "Judiciary", "peers": []string{"peer0.judiciary.nanayam.com:11051"}, "role": "Escalation"},
		},
		"orderers": []string{"orderer.nanayam.com:7050"},
	}
	writeJSON(w, http.StatusOK, info)
}

// ---------------------------------------------------------------------------
// Ledger Explorer
// ---------------------------------------------------------------------------

func (s *RESTServer) handleLedgerBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" {
		startStr = "0"
	}
	if endStr == "" {
		endStr = "10"
	}
	start, err := strconv.ParseUint(startStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid start"})
		return
	}
	end, err := strconv.ParseUint(endStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid end"})
		return
	}

	network := s.cfg.gateway.GetNetwork(s.cfg.ChannelName)
	contract := network.GetContract("qscc")

	type blockSummary struct {
		Number        uint64 `json:"number"`
		Hash          string `json:"hash"`
		PrevHash      string `json:"prevHash"`
		TxCount       int    `json:"txCount"`
		Timestamp     string `json:"timestamp"`
		DataHash      string `json:"dataHash"`
	}

	var blocks []blockSummary
	for i := start; i <= end; i++ {
		blockNumHex := strconv.FormatUint(i, 16)
		result, err := contract.EvaluateTransaction("GetBlockByNumber", s.cfg.ChannelName, blockNumHex)
		if err != nil {
			// If we hit a block that doesn't exist yet, stop
			break
		}

		var blk common.Block
		if err := proto.Unmarshal(result, &blk); err != nil {
			continue
		}

		headerBytes, _ := proto.Marshal(blk.Header)
		hash := sha256.Sum256(headerBytes)
		blocks = append(blocks, blockSummary{
			Number:    i,
			Hash:      hex.EncodeToString(hash[:]),
			PrevHash:  hex.EncodeToString(blk.Header.PreviousHash),
			TxCount:   len(blk.Data.Data),
			Timestamp: time.Now().UTC().Format(time.RFC3339), // placeholder; real timestamp requires tx decode
			DataHash:  hex.EncodeToString(blk.Header.DataHash),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"blocks": blocks})
}

func (s *RESTServer) handleLedgerActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	network := s.cfg.gateway.GetNetwork(s.cfg.ChannelName)
	contract := network.GetContract("qscc")

	// Get chain info for height
	chainInfo, err := contract.EvaluateTransaction("GetChainInfo", s.cfg.ChannelName)
	var height uint64
	if err == nil {
		var ci common.BlockchainInfo
		if proto.Unmarshal(chainInfo, &ci) == nil {
			height = ci.Height
		}
	}

	// Also get complaint stats
	fh := s.handler.(*FabricHandler)
	listResp, _ := fh.ListComplaints(r.Context(), &pb.ListComplaintsRequest{})
	complaintCount := 0
	if listResp != nil {
		complaintCount = len(listResp.ComplaintIds)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"height":          height,
		"complaintCount":  complaintCount,
		"channel":         s.cfg.ChannelName,
		"chaincode":       s.cfg.ChaincodeName,
	})
}

// ---------------------------------------------------------------------------
// Registration of all handlers
// ---------------------------------------------------------------------------

func (s *RESTServer) register(mux *http.ServeMux) {
	// Auth (no auth required)
	mux.HandleFunc("/v1/Register", corsMiddleware(s.handleRegister))
	mux.HandleFunc("/v1/Login", corsMiddleware(s.handleLogin))
	mux.HandleFunc("/v1/Me", corsMiddleware(s.authMiddleware(s.handleMe)))

	// Channel & Ledger (auth required)
	mux.HandleFunc("/v1/ChannelInfo", corsMiddleware(s.authMiddleware(s.handleChannelInfo)))
	mux.HandleFunc("/v1/LedgerBlocks", corsMiddleware(s.authMiddleware(s.handleLedgerBlocks)))
	mux.HandleFunc("/v1/LedgerActivity", corsMiddleware(s.authMiddleware(s.handleLedgerActivity)))

	// -------------------------------------------------------------------------
	// Asset endpoints (original)
	// -------------------------------------------------------------------------
	mux.HandleFunc("/v1/CreateAsset", corsMiddleware(s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req pb.CreateAssetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := s.handler.CreateAsset(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})))

	mux.HandleFunc("/v1/QueryAsset", corsMiddleware(s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		assetID := r.URL.Query().Get("assetId")
		if assetID == "" {
			http.Error(w, "missing assetId", http.StatusBadRequest)
			return
		}
		resp, err := s.handler.QueryAsset(r.Context(), &pb.QueryAssetRequest{AssetId: assetID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})))

	mux.HandleFunc("/v1/ListAssets", corsMiddleware(s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp, err := s.handler.ListAssets(r.Context(), &pb.ListAssetsRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})))

	// -------------------------------------------------------------------------
	// Complaint endpoints
	// -------------------------------------------------------------------------
	mux.HandleFunc("/v1/SubmitComplaint", corsMiddleware(s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req pb.SubmitComplaintRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := s.handler.SubmitComplaint(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})))

	mux.HandleFunc("/v1/UpdateComplaint", corsMiddleware(s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req pb.UpdateComplaintRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := s.handler.UpdateComplaint(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})))

	mux.HandleFunc("/v1/QueryComplaint", corsMiddleware(s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Query().Get("complaintId")
		if id == "" {
			http.Error(w, "missing complaintId", http.StatusBadRequest)
			return
		}
		resp, err := s.handler.QueryComplaint(r.Context(), &pb.QueryComplaintRequest{ComplaintID: id})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})))

	mux.HandleFunc("/v1/ListComplaints", corsMiddleware(s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp, err := s.handler.ListComplaints(r.Context(), &pb.ListComplaintsRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})))

	mux.HandleFunc("/v1/GetComplaintHistory", corsMiddleware(s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Query().Get("complaintId")
		if id == "" {
			http.Error(w, "missing complaintId", http.StatusBadRequest)
			return
		}
		resp, err := s.handler.GetComplaintHistory(r.Context(), &pb.GetComplaintHistoryRequest{ComplaintID: id})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})))

	// Health check (no auth)
	mux.HandleFunc("/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}))
}
