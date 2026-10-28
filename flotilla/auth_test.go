package flotilla

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stitchfix/flotilla-os/state"
	muxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
)

const (
	userToken     = "user-secret-token"
	operatorToken = "operator-secret-token"
)

func setUpAuthRouter(t *testing.T) *muxtrace.Router {
	t.Setenv("EKS_ALLOWED_SERVICE_ACCOUNTS", "approved-sa")
	auth, err := newAuthenticatorFromEntries([]string{
		"alice:user:" + userToken,
		"scheduler:operator:" + operatorToken,
	})
	if err != nil {
		t.Fatal(err)
	}
	return setUpWithAuth(t, auth)
}

func doRequest(router http.Handler, method, path, body, token string) *http.Response {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Result()
}

func TestAuth_RejectsUnauthenticatedRequests(t *testing.T) {
	router := setUpAuthRouter(t)
	run := `{"run_tags":{"owner_id":"mallory"}, "service_account":"approved-sa"}`
	cases := []struct{ method, path, body, token string }{
		{"POST", "/api/v6/task", `{"alias":"x","image":"evil/image","group_name":"g","memory":100}`, ""},
		{"PUT", "/api/v6/task/A/execute", run, ""},
		{"PUT", "/api/v4/task/A/execute", run, ""},
		{"GET", "/api/v1/task", "", ""},
		{"GET", "/api/v6/history/runA", "", ""},
		{"PUT", "/api/v6/runA/status", `{"status":"STOPPED"}`, ""},
		{"POST", "/api/v6/clusters", `{"name":"x"}`, ""},
		{"PUT", "/api/v6/task/A/execute", run, "wrong-token"},
	}
	for _, c := range cases {
		resp := doRequest(router, c.method, c.path, c.body, c.token)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s (token=%q): expected 401, got %d", c.method, c.path, c.token, resp.StatusCode)
		}
	}
}

func TestAuth_HealthCheckIsPublic(t *testing.T) {
	router := setUpAuthRouter(t)
	resp := doRequest(router, "GET", "/api/v6/health", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for health check, got %d", resp.StatusCode)
	}
}

func TestAuth_OwnerIDDerivedFromVerifiedIdentity(t *testing.T) {
	router := setUpAuthRouter(t)
	body := `{"cluster":"cluster1", "run_tags":{"owner_id":"someone-else"}}`
	resp := doRequest(router, "PUT", "/api/v6/task/A/execute", body, userToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var r state.Run
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatal(err)
	}
	if r.User != "alice" {
		t.Errorf("expected run user to be verified identity [alice], got [%s]", r.User)
	}

	// owner_id is optional for authenticated users.
	resp = doRequest(router, "PUT", "/api/v6/task/A/execute", `{"cluster":"cluster1"}`, userToken)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 without owner_id, got %d", resp.StatusCode)
	}
}

func TestAuth_OperatorMayLaunchOnBehalfOfOwner(t *testing.T) {
	router := setUpAuthRouter(t)
	body := `{"cluster":"cluster1", "run_tags":{"owner_id":"bob"}}`
	resp := doRequest(router, "PUT", "/api/v6/task/A/execute", body, operatorToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var r state.Run
	_ = json.NewDecoder(resp.Body).Decode(&r)
	if r.User != "bob" {
		t.Errorf("expected run user [bob], got [%s]", r.User)
	}
}

func TestAuth_ServiceAccountAllowlist(t *testing.T) {
	router := setUpAuthRouter(t)
	bad := `{"cluster":"cluster1", "service_account":"cluster-admin"}`
	resp := doRequest(router, "PUT", "/api/v6/task/A/execute", bad, userToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for unlisted service account, got %d", resp.StatusCode)
	}
	good := `{"cluster":"cluster1", "service_account":"approved-sa"}`
	resp = doRequest(router, "PUT", "/api/v6/task/A/execute", good, userToken)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for allowlisted service account, got %d", resp.StatusCode)
	}
}

func TestAuth_OperatorOnlyRoutes(t *testing.T) {
	router := setUpAuthRouter(t)
	cases := []struct{ method, path, body string }{
		{"PUT", "/api/v1/runA/status", `{"status":"STOPPED"}`},
		{"PUT", "/api/v6/runA/status", `{"status":"STOPPED"}`},
		{"POST", "/api/v6/clusters", `{"name":"x"}`},
		{"PUT", "/api/v6/clusters/c1", `{"name":"x"}`},
		{"DELETE", "/api/v6/clusters/c1", ""},
		{"PUT", "/api/v5/worker", `[]`},
		{"PUT", "/api/v5/worker/retry", `{}`},
	}
	for _, c := range cases {
		resp := doRequest(router, c.method, c.path, c.body, userToken)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s %s as user: expected 403, got %d", c.method, c.path, resp.StatusCode)
		}
	}
	resp := doRequest(router, "PUT", "/api/v6/runA/status", `{"status":"STOPPED"}`, operatorToken)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		t.Errorf("operator should be allowed to update run status, got %d", resp.StatusCode)
	}
}

func TestAuth_NilAuthenticatorFailsClosed(t *testing.T) {
	router := setUpWithAuth(t, nil)
	resp := doRequest(router, "GET", "/api/v6/task", "", "anything")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 with no authenticator configured, got %d", resp.StatusCode)
	}
}

func TestNewAuthenticatorFromEntries(t *testing.T) {
	if _, err := newAuthenticatorFromEntries(nil); err == nil {
		t.Error("expected error when no keys are configured")
	}
	if _, err := newAuthenticatorFromEntries([]string{"alice:admin:tok"}); err == nil {
		t.Error("expected error for unknown role")
	}
	if _, err := newAuthenticatorFromEntries([]string{"alice:user:"}); err == nil {
		t.Error("expected error for empty token")
	}
	a, err := newAuthenticatorFromEntries([]string{" alice:user:a:b:c ", ""})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "bearer a:b:c")
	if c, ok := a.authenticate(req); !ok || c.Identity != "alice" || c.Role != RoleUser {
		t.Errorf("expected token containing colons to authenticate as alice, got %+v %v", c, ok)
	}
}
