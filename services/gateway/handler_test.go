package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/bytamilan/nanayam/services/gateway/proto"
)

// mockHandler implements pb.FabricServiceServer for unit testing.
// Setting err makes every method fail, which exercises the error paths.
type mockHandler struct {
	err error
}

func (m *mockHandler) CreateAsset(ctx context.Context, req *pb.CreateAssetRequest) (*pb.CreateAssetResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pb.CreateAssetResponse{Success: true}, nil
}
func (m *mockHandler) QueryAsset(ctx context.Context, req *pb.QueryAssetRequest) (*pb.QueryAssetResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pb.QueryAssetResponse{Data: fmt.Sprintf(`{"ID":%q}`, req.AssetId)}, nil
}
func (m *mockHandler) ListAssets(ctx context.Context, req *pb.ListAssetsRequest) (*pb.ListAssetsResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pb.ListAssetsResponse{AssetIds: []string{"asset1", "asset2"}}, nil
}
func (m *mockHandler) SubmitComplaint(ctx context.Context, req *pb.SubmitComplaintRequest) (*pb.SubmitComplaintResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pb.SubmitComplaintResponse{Success: true}, nil
}
func (m *mockHandler) UpdateComplaint(ctx context.Context, req *pb.UpdateComplaintRequest) (*pb.UpdateComplaintResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pb.UpdateComplaintResponse{Success: true}, nil
}
func (m *mockHandler) QueryComplaint(ctx context.Context, req *pb.QueryComplaintRequest) (*pb.QueryComplaintResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pb.QueryComplaintResponse{Data: fmt.Sprintf(`{"complaintId":%q,"status":"Submitted"}`, req.ComplaintID)}, nil
}
func (m *mockHandler) ListComplaints(ctx context.Context, req *pb.ListComplaintsRequest) (*pb.ListComplaintsResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pb.ListComplaintsResponse{ComplaintIds: []string{"COMP-001", "COMP-002"}}, nil
}
func (m *mockHandler) GetComplaintHistory(ctx context.Context, req *pb.GetComplaintHistoryRequest) (*pb.GetComplaintHistoryResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pb.GetComplaintHistoryResponse{Data: `[{"txId":"tx1"}]`}, nil
}

// newTestServer wires a RESTServer around a mock Fabric handler and returns the
// mux plus a bearer token for the seeded admin user.
func newTestServer(t *testing.T, handler pb.FabricServiceServer) (*http.ServeMux, string) {
	t.Helper()

	store := NewAuthStore()
	store.SeedAdmin()

	token, err := store.Login("admin", "admin")
	if err != nil {
		t.Fatalf("login seeded admin: %v", err)
	}

	rest := &RESTServer{
		handler:   handler,
		authStore: store,
		cfg: &GatewayConfig{
			MSP_ID:        "Org1MSP",
			ChannelName:   "mychannel",
			ChaincodeName: "basic",
		},
	}

	mux := http.NewServeMux()
	rest.register(mux)
	return mux, token
}

// do issues a request against the mux, attaching the bearer token when non-empty.
func do(t *testing.T, mux *http.ServeMux, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(encoded)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func decode[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()

	var out T
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response %q: %v", rr.Body.String(), err)
	}
	return out
}

// ---------------------------------------------------------------------------
// Unauthenticated endpoints
// ---------------------------------------------------------------------------

func TestHealthEndpoint(t *testing.T) {
	mux, _ := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodGet, "/health", "", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body := decode[map[string]string](t, rr)
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", body)
	}
}

func TestConfigEndpointReportsChannelAndSignupState(t *testing.T) {
	mux, _ := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodGet, "/v1/Config", "", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body := decode[map[string]any](t, rr)
	if body["channel"] != "mychannel" {
		t.Fatalf("expected channel mychannel, got %v", body["channel"])
	}
	if body["chaincode"] != "basic" {
		t.Fatalf("expected chaincode basic, got %v", body["chaincode"])
	}
	// Signup defaults to disabled so a fresh deployment is not open to the world.
	if body["signupEnabled"] != false {
		t.Fatalf("expected signupEnabled false by default, got %v", body["signupEnabled"])
	}
}

func TestRegisterRejectedWhenSignupDisabled(t *testing.T) {
	mux, _ := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodPost, "/v1/Register", "", map[string]string{
		"username": "mallory",
		"password": "hunter2",
		"org":      "Org1MSP",
	})
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when signup is disabled, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRegisterAndLoginWhenSignupEnabled(t *testing.T) {
	t.Setenv("AUTH_SIGNUP_ENABLED", "true")
	mux, _ := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodPost, "/v1/Register", "", map[string]string{
		"username": "alice",
		"password": "alice-pw",
		"org":      "Org1MSP",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	user := decode[map[string]any](t, rr)
	if user["username"] != "alice" {
		t.Fatalf("expected username alice, got %v", user["username"])
	}
	// The password hash must never cross the wire.
	if _, leaked := user["PasswordHash"]; leaked {
		t.Fatalf("register response leaked the password hash: %v", user)
	}

	rr = do(t, mux, http.MethodPost, "/v1/Login", "", map[string]string{
		"username": "alice",
		"password": "alice-pw",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 on login, got %d: %s", rr.Code, rr.Body.String())
	}
	if decode[map[string]string](t, rr)["token"] == "" {
		t.Fatal("expected a token in the login response")
	}
}

func TestLoginWithWrongPasswordIsRejected(t *testing.T) {
	mux, _ := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodPost, "/v1/Login", "", map[string]string{
		"username": "admin",
		"password": "not-the-password",
	})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Authentication enforcement
// ---------------------------------------------------------------------------

func TestProtectedEndpointsRequireAToken(t *testing.T) {
	mux, _ := newTestServer(t, &mockHandler{})

	protected := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/v1/Me"},
		{http.MethodGet, "/v1/ChannelInfo"},
		{http.MethodGet, "/v1/ListAssets"},
		{http.MethodGet, "/v1/QueryAsset?assetId=asset1"},
		{http.MethodPost, "/v1/CreateAsset"},
		{http.MethodGet, "/v1/ListComplaints"},
		{http.MethodGet, "/v1/QueryComplaint?complaintId=COMP-001"},
		{http.MethodPost, "/v1/SubmitComplaint"},
		{http.MethodPost, "/v1/UpdateComplaint"},
		{http.MethodGet, "/v1/GetComplaintHistory?complaintId=COMP-001"},
		{http.MethodGet, "/v1/LedgerBlocks"},
		{http.MethodGet, "/v1/LedgerActivity"},
	}

	for _, tc := range protected {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			if rr := do(t, mux, tc.method, tc.path, "", nil); rr.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401 without a token, got %d", rr.Code)
			}
		})
	}
}

func TestMalformedAuthorizationHeadersAreRejected(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	headers := map[string]string{
		"missing scheme":  token,
		"wrong scheme":    "Basic " + token,
		"garbage token":   "Bearer not-a-jwt",
		"empty bearer":    "Bearer ",
		"only the scheme": "Bearer",
	}

	for name, header := range headers {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/ListAssets", nil)
			req.Header.Set("Authorization", header)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rr.Code)
			}
		})
	}
}

func TestMeReturnsTheAuthenticatedUser(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodGet, "/v1/Me", token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	user := decode[map[string]any](t, rr)
	if user["username"] != "admin" {
		t.Fatalf("expected admin, got %v", user["username"])
	}
	if user["role"] != "admin" {
		t.Fatalf("expected role admin, got %v", user["role"])
	}
}

// A token signed with a different secret must not unlock another deployment.
func TestTokenFromAnotherDeploymentIsRejected(t *testing.T) {
	t.Setenv("AUTH_JWT_SECRET", "secret-of-deployment-a")
	otherStore := NewAuthStore()
	otherStore.SeedAdmin()
	foreignToken, err := otherStore.Login("admin", "admin")
	if err != nil {
		t.Fatalf("login on foreign store: %v", err)
	}

	t.Setenv("AUTH_JWT_SECRET", "secret-of-deployment-b")
	mux, _ := newTestServer(t, &mockHandler{})

	if rr := do(t, mux, http.MethodGet, "/v1/ListAssets", foreignToken, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for a foreign token, got %d", rr.Code)
	}
}

func TestCORSPreflightSucceedsWithoutAuth(t *testing.T) {
	mux, _ := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodOptions, "/v1/ListAssets", "", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for preflight, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected permissive CORS origin, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("expected Access-Control-Allow-Headers on the preflight response")
	}
}

// ---------------------------------------------------------------------------
// Asset endpoints
// ---------------------------------------------------------------------------

func TestCreateAssetEndpoint(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodPost, "/v1/CreateAsset", token, map[string]any{
		"assetId": "asset1", "color": "blue", "size": 10,
		"owner": "Alice", "appraisedValue": 100,
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !decode[pb.CreateAssetResponse](t, rr).Success {
		t.Fatal("expected success")
	}
}

func TestCreateAssetRejectsMalformedJSON(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	req := httptest.NewRequest(http.MethodPost, "/v1/CreateAsset", bytes.NewReader([]byte("{not json")))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestQueryAssetRequiresAssetID(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	if rr := do(t, mux, http.MethodGet, "/v1/QueryAsset", token, nil); rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without assetId, got %d", rr.Code)
	}

	rr := do(t, mux, http.MethodGet, "/v1/QueryAsset?assetId=asset1", token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if data := decode[pb.QueryAssetResponse](t, rr).Data; data != `{"ID":"asset1"}` {
		t.Fatalf("assetId was not forwarded to the handler, got %q", data)
	}
}

func TestListAssetsEndpoint(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodGet, "/v1/ListAssets", token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if got := decode[pb.ListAssetsResponse](t, rr).AssetIds; len(got) != 2 {
		t.Fatalf("expected 2 asset ids, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Complaint endpoints
// ---------------------------------------------------------------------------

func TestSubmitComplaintEndpoint(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodPost, "/v1/SubmitComplaint", token, map[string]any{
		"complaintId":     "COMP-001",
		"category":        "bribery",
		"citizenHash":     "sha256:abc",
		"descriptionHash": "sha256:def",
		"attachmentsRef":  "ipfs:QmTest",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !decode[pb.SubmitComplaintResponse](t, rr).Success {
		t.Fatal("expected success")
	}
}

func TestUpdateComplaintEndpoint(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodPost, "/v1/UpdateComplaint", token, map[string]any{
		"complaintId": "COMP-001",
		"action":      "acknowledge",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !decode[pb.UpdateComplaintResponse](t, rr).Success {
		t.Fatal("expected success")
	}
}

func TestQueryComplaintEndpoint(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	if rr := do(t, mux, http.MethodGet, "/v1/QueryComplaint", token, nil); rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without complaintId, got %d", rr.Code)
	}

	rr := do(t, mux, http.MethodGet, "/v1/QueryComplaint?complaintId=COMP-001", token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if decode[pb.QueryComplaintResponse](t, rr).Data == "" {
		t.Fatal("expected complaint data")
	}
}

func TestListComplaintsEndpoint(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodGet, "/v1/ListComplaints", token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if got := decode[pb.ListComplaintsResponse](t, rr).ComplaintIds; len(got) == 0 {
		t.Fatal("expected complaint ids")
	}
}

func TestGetComplaintHistoryEndpoint(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	if rr := do(t, mux, http.MethodGet, "/v1/GetComplaintHistory", token, nil); rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without complaintId, got %d", rr.Code)
	}

	rr := do(t, mux, http.MethodGet, "/v1/GetComplaintHistory?complaintId=COMP-001", token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if decode[pb.GetComplaintHistoryResponse](t, rr).Data == "" {
		t.Fatal("expected history data")
	}
}

func TestChannelInfoEndpoint(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	rr := do(t, mux, http.MethodGet, "/v1/ChannelInfo", token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	info := decode[map[string]any](t, rr)
	if info["channel"] != "mychannel" {
		t.Fatalf("expected channel mychannel, got %v", info["channel"])
	}
	orgs, ok := info["organizations"].([]any)
	if !ok || len(orgs) == 0 {
		t.Fatalf("expected organizations in the response, got %v", info["organizations"])
	}
}

// ---------------------------------------------------------------------------
// Error propagation and method routing
// ---------------------------------------------------------------------------

func TestHandlerErrorsSurfaceAs500(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{err: fmt.Errorf("chaincode unavailable")})

	cases := []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodPost, "/v1/CreateAsset", map[string]any{"assetId": "asset1"}},
		{http.MethodGet, "/v1/QueryAsset?assetId=asset1", nil},
		{http.MethodGet, "/v1/ListAssets", nil},
		{http.MethodPost, "/v1/SubmitComplaint", map[string]any{"complaintId": "COMP-001"}},
		{http.MethodPost, "/v1/UpdateComplaint", map[string]any{"complaintId": "COMP-001"}},
		{http.MethodGet, "/v1/QueryComplaint?complaintId=COMP-001", nil},
		{http.MethodGet, "/v1/ListComplaints", nil},
		{http.MethodGet, "/v1/GetComplaintHistory?complaintId=COMP-001", nil},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			if rr := do(t, mux, tc.method, tc.path, token, tc.body); rr.Code != http.StatusInternalServerError {
				t.Fatalf("expected 500, got %d: %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	mux, token := newTestServer(t, &mockHandler{})

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodDelete, "/v1/SubmitComplaint"},
		{http.MethodGet, "/v1/CreateAsset"},
		{http.MethodPost, "/v1/ListAssets"},
		{http.MethodPost, "/v1/QueryAsset"},
		{http.MethodDelete, "/v1/ListComplaints"},
		{http.MethodPost, "/v1/Me"},
		{http.MethodPost, "/v1/ChannelInfo"},
		{http.MethodGet, "/v1/Login"},
		{http.MethodGet, "/v1/Register"},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			if rr := do(t, mux, tc.method, tc.path, token, nil); rr.Code != http.StatusMethodNotAllowed {
				t.Fatalf("expected 405, got %d", rr.Code)
			}
		})
	}
}
