package main

import (
	"encoding/json"
	"net/http"

	pb "github.com/bytamilan/nanayam/services/gateway/proto"
)

// registerRESTHandlers mounts simple HTTP handlers that wrap the gRPC service.
func registerRESTHandlers(mux *http.ServeMux, handler *FabricHandler) {
	mux.HandleFunc("/v1/CreateAsset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req pb.CreateAssetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := handler.CreateAsset(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v1/QueryAsset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		assetID := r.URL.Query().Get("assetId")
		if assetID == "" {
			http.Error(w, "missing assetId", http.StatusBadRequest)
			return
		}
		resp, err := handler.QueryAsset(r.Context(), &pb.QueryAssetRequest{AssetId: assetID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v1/ListAssets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp, err := handler.ListAssets(r.Context(), &pb.ListAssetsRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}
