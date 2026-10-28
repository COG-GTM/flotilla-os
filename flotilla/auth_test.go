package flotilla

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

const (
	aliceKey  = "alice-user-key"
	workerKey = "status-worker-key"
	opsKey    = "operator-only-key"
)

func scopedAuthenticator() *Authenticator {
	a := testAuthenticator()
	a.keys[digest(aliceKey)] = Principal{Identity: "alice@acme.example", Roles: map[Role]bool{RoleUser: true}}
	a.keys[digest(workerKey)] = Principal{Identity: "status-worker", Roles: map[Role]bool{RoleWorker: true}}
	a.keys[digest(opsKey)] = Principal{Identity: "ops", Roles: map[Role]bool{RoleOperator: true}}
	return a
}

func testPolicy() LaunchPolicy {
	return LaunchPolicy{
		serviceAccounts: map[string]bool{"flotilla-job": true, "etl-runner": true},
		imageRegistries: []string{"registry.acme.example"},
	}
}

func do(t *testing.T, h http.Handler, method, path, key, body string) *http.Response {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	}
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w.Result()
}

func TestAuth_UnauthenticatedRequestsRejected(t *testing.T) {
	router, imp := setUpWith(t, scopedAuthenticator(), testPolicy())
	defsBefore, runsBefore := len(imp.Definitions), len(imp.Runs)

	cases := []struct{ method, path, body string }{
		{"POST", "/api/v6/task", `{"alias":"x","group_name":"g","image":"attacker/miner","memory":1}`},
		{"PUT", "/api/v6/task/A/execute", `{"run_tags":{"owner_id":"mallory"},"service_account":"cluster-admin"}`},
		{"PUT", "/api/v4/task/A/execute", `{"run_tags":{"owner_id":"mallory"}}`},
		{"PUT", "/api/v1/task/A/execute", `{}`},
		{"PUT", "/api/v6/task/alias/aliasA/execute", `{"run_tags":{"owner_id":"mallory"}}`},
		{"PUT", "/api/v7/template/T/execute", `{"owner_id":"mallory"}`},
		{"GET", "/api/v6/task", ""},
		{"GET", "/api/v6/history/runA", ""},
		{"PUT", "/api/v6/runA/status", `{"status":"STOPPED","exit_code":0}`},
		{"DELETE", "/api/v6/task/A", ""},
		{"POST", "/api/v6/clusters", `{"name":"evil"}`},
		{"DELETE", "/api/v6/clusters/cluster1", ""},
		{"PUT", "/api/v5/worker", `{}`},
	}
	for _, key := range []string{"", "not-a-real-key"} {
		for _, c := range cases {
			resp := do(t, router, c.method, c.path, key, c.body)
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("%s %s key=%q: expected 401, got %d", c.method, c.path, key, resp.StatusCode)
			}
			if resp.Header.Get("WWW-Authenticate") == "" {
				t.Errorf("%s %s: expected WWW-Authenticate header", c.method, c.path)
			}
		}
	}
	if len(imp.Definitions) != defsBefore || len(imp.Runs) != runsBefore {
		t.Errorf("unauthenticated requests changed state: definitions %d->%d runs %d->%d",
			defsBefore, len(imp.Definitions), runsBefore, len(imp.Runs))
	}
	if imp.Runs["runA"].Status != state.StatusRunning {
		t.Errorf("unauthenticated status update changed runA to %s", imp.Runs["runA"].Status)
	}
}

func TestAuth_NonBearerSchemeRejected(t *testing.T) {
	router, _ := setUpWith(t, scopedAuthenticator(), testPolicy())
	req := httptest.NewRequest("GET", "/api/v6/task", nil)
	req.Header.Set("Authorization", "Basic "+aliceKey)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for Basic scheme, got %d", w.Code)
	}
}

func TestAuth_HealthCheckIsPublic(t *testing.T) {
	router, _ := setUpWith(t, scopedAuthenticator(), testPolicy())
	if resp := do(t, router, "GET", "/api/v6/health", "", ""); resp.StatusCode != http.StatusOK {
		t.Errorf("expected health check to be public, got %d", resp.StatusCode)
	}
}

func TestAuth_NilAuthenticatorFailsClosed(t *testing.T) {
	router, _ := setUpWith(t, nil, LaunchPolicy{})
	if resp := do(t, router, "GET", "/api/v6/task", testAPIKey, ""); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 with no authenticator, got %d", resp.StatusCode)
	}
}

func TestAuth_RoleEnforcement(t *testing.T) {
	router, imp := setUpWith(t, scopedAuthenticator(), testPolicy())

	forbiddenForUser := []struct{ method, path, body string }{
		{"POST", "/api/v6/clusters", `{"name":"evil","status":"ACTIVE"}`},
		{"PUT", "/api/v6/clusters/cluster1", `{"namespace":"kube-system"}`},
		{"DELETE", "/api/v6/clusters/cluster1", ""},
		{"PUT", "/api/v5/worker", `{}`},
		{"PUT", "/api/v5/worker/retry", `{}`},
		{"PUT", "/api/v6/runA/status", `{"status":"STOPPED","exit_code":0}`},
		{"PUT", "/api/v1/runA/status", `{"status":"STOPPED","exit_code":0}`},
	}
	for _, c := range forbiddenForUser {
		if resp := do(t, router, c.method, c.path, aliceKey, c.body); resp.StatusCode != http.StatusForbidden {
			t.Errorf("user %s %s: expected 403, got %d", c.method, c.path, resp.StatusCode)
		}
	}
	if imp.Runs["runA"].Status != state.StatusRunning {
		t.Errorf("user was able to change runA status to %s", imp.Runs["runA"].Status)
	}

	if resp := do(t, router, "GET", "/api/v6/task", aliceKey, ""); resp.StatusCode != http.StatusOK {
		t.Errorf("user GET /api/v6/task: expected 200, got %d", resp.StatusCode)
	}
	if resp := do(t, router, "GET", "/api/v6/task", workerKey, ""); resp.StatusCode != http.StatusForbidden {
		t.Errorf("worker GET /api/v6/task: expected 403, got %d", resp.StatusCode)
	}
	if resp := do(t, router, "PUT", "/api/v6/runA/status", workerKey, `{"status":"STOPPED","exit_code":0}`); resp.StatusCode != http.StatusOK {
		t.Errorf("worker PUT status: expected 200, got %d", resp.StatusCode)
	}
	if resp := do(t, router, "PUT", "/api/v6/runA/status", opsKey, `{"status":"STOPPED","exit_code":0}`); resp.StatusCode != http.StatusForbidden {
		t.Errorf("operator PUT status: expected 403, got %d", resp.StatusCode)
	}
	if resp := do(t, router, "GET", "/api/v6/task", opsKey, ""); resp.StatusCode != http.StatusForbidden {
		t.Errorf("operator GET /api/v6/task: expected 403, got %d", resp.StatusCode)
	}
	if resp := do(t, router, "DELETE", "/api/v6/clusters/cluster1", opsKey, ""); resp.StatusCode != http.StatusOK {
		t.Errorf("operator DELETE cluster: expected 200, got %d", resp.StatusCode)
	}
}

func TestAuth_ServiceAccountAllowlist(t *testing.T) {
	router, imp := setUpWith(t, scopedAuthenticator(), testPolicy())

	for _, path := range []string{"/api/v6/task/A/execute", "/api/v4/task/A/execute", "/api/v6/task/alias/aliasA/execute"} {
		body := `{"cluster":"cluster1","run_tags":{"owner_id":"x"},"service_account":"cluster-admin"}`
		before := len(imp.Runs)
		resp := do(t, router, "PUT", path, aliceKey, body)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s with disallowed service_account: expected 400, got %d", path, resp.StatusCode)
		}
		if len(imp.Runs) != before {
			t.Errorf("%s with disallowed service_account created a run", path)
		}
	}

	v2 := `{"cluster":"cluster1","run_tags":{"owner_email":"x","team_name":"t"},"service_account":"cluster-admin"}`
	if resp := do(t, router, "PUT", "/api/v2/task/A/execute", aliceKey, v2); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("v2 with disallowed service_account: expected 400, got %d", resp.StatusCode)
	}
	tpl := `{"owner_id":"x","service_account":"cluster-admin"}`
	if resp := do(t, router, "PUT", "/api/v7/template/T/execute", aliceKey, tpl); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("template run with disallowed service_account: expected 400, got %d", resp.StatusCode)
	}

	resp := do(t, router, "PUT", "/api/v6/task/A/execute", aliceKey,
		`{"cluster":"cluster1","run_tags":{"owner_id":"x"},"service_account":"etl-runner"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("allowlisted service_account: expected 200, got %d", resp.StatusCode)
	}
	var run state.Run
	_ = json.NewDecoder(resp.Body).Decode(&run)
	if run.ServiceAccount == nil || *run.ServiceAccount != "etl-runner" {
		t.Errorf("expected run service_account etl-runner, got %v", run.ServiceAccount)
	}
}

func TestAuth_OwnerDerivedFromVerifiedIdentity(t *testing.T) {
	router, imp := setUpWith(t, scopedAuthenticator(), testPolicy())

	for _, path := range []string{"/api/v6/task/A/execute", "/api/v6/task/alias/aliasA/execute"} {
		resp := do(t, router, "PUT", path, aliceKey, `{"cluster":"cluster1","run_tags":{"owner_id":"bob@acme.example"}}`)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, resp.StatusCode)
		}
		var run state.Run
		_ = json.NewDecoder(resp.Body).Decode(&run)
		if run.User != "alice@acme.example" {
			t.Errorf("%s: expected run user to be the verified identity, got %q", path, run.User)
		}
		if stored := imp.Runs[run.RunID]; stored.User != "alice@acme.example" {
			t.Errorf("%s: expected stored run user to be the verified identity, got %q", path, stored.User)
		}
	}

	// Reserved env vars in the request cannot override the verified owner or run id.
	resp := do(t, router, "PUT", "/api/v6/task/A/execute", aliceKey,
		`{"cluster":"cluster1","env":[{"name":"FLOTILLA_RUN_OWNER_ID","value":"bob@acme.example"},{"name":"FLOTILLA_RUN_ID","value":"fake"},{"name":"E1","value":"V1"}]}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("launch with reserved env: expected 200, got %d", resp.StatusCode)
	}
	var run state.Run
	_ = json.NewDecoder(resp.Body).Decode(&run)
	seen := map[string][]string{}
	if run.Env != nil {
		for _, e := range *run.Env {
			seen[e.Name] = append(seen[e.Name], e.Value)
		}
	}
	if v := seen["FLOTILLA_RUN_OWNER_ID"]; len(v) != 1 || v[0] != "alice@acme.example" {
		t.Errorf("expected a single verified FLOTILLA_RUN_OWNER_ID, got %v", v)
	}
	if v := seen["FLOTILLA_RUN_ID"]; len(v) != 1 || v[0] != run.RunID {
		t.Errorf("expected a single FLOTILLA_RUN_ID equal to the run id, got %v", v)
	}
	if v := seen["E1"]; len(v) != 1 || v[0] != "V1" {
		t.Errorf("expected caller env E1=V1 to be kept, got %v", v)
	}

	// owner_id may now be omitted; it comes from the credential.
	resp = do(t, router, "PUT", "/api/v6/task/A/execute", aliceKey, `{"cluster":"cluster1"}`)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("launch without owner_id: expected 200, got %d", resp.StatusCode)
	}
}

func TestAuth_ExtractUserInfoIgnoresSpoofedHeaders(t *testing.T) {
	ep := endpoints{}
	req := httptest.NewRequest("DELETE", "/api/v6/task/A/history/runA", nil)
	req.Header.Set("X-Forwarded-Email", "ceo@acme.example")
	req.Header.Set("X-User-Name", "ceo")
	req = req.WithContext(WithPrincipal(req.Context(), Principal{Identity: "alice@acme.example", Roles: map[Role]bool{RoleUser: true}}))
	info := ep.ExtractUserInfo(req)
	if info.Email != "alice@acme.example" || info.Name != "alice@acme.example" {
		t.Errorf("expected user info from verified identity, got %+v", info)
	}
}

func TestAuth_ImageRegistryAllowlist(t *testing.T) {
	router, imp := setUpWith(t, scopedAuthenticator(), testPolicy())

	for _, image := range []string{"attacker/miner", "registry.acme.example.evil.io/miner", "registry.acme.exampleX/miner"} {
		body := `{"alias":"a-` + strings.ReplaceAll(image, "/", "-") + `","group_name":"g","memory":1,"image":"` + image + `"}`
		if resp := do(t, router, "POST", "/api/v6/task", aliceKey, body); resp.StatusCode != http.StatusBadRequest {
			t.Errorf("create definition with image %s: expected 400, got %d", image, resp.StatusCode)
		}
	}
	if resp := do(t, router, "PUT", "/api/v6/task/A", aliceKey, `{"image":"attacker/miner"}`); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("update definition to disallowed image: expected 400, got %d", resp.StatusCode)
	}
	if imp.Definitions["A"].Image == "attacker/miner" {
		t.Errorf("definition image was updated to a disallowed registry")
	}
	if resp := do(t, router, "POST", "/api/v7/template", aliceKey, `{"template_name":"t","image":"attacker/miner"}`); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("create template with disallowed image: expected 400, got %d", resp.StatusCode)
	}

	ok := `{"alias":"ok","group_name":"g","memory":1,"image":"registry.acme.example/team/job:1"}`
	if resp := do(t, router, "POST", "/api/v6/task", aliceKey, ok); resp.StatusCode != http.StatusOK {
		t.Errorf("create definition with allowed image: expected 200, got %d", resp.StatusCode)
	}
}

func TestNewAuthenticator_Config(t *testing.T) {
	t.Setenv("HTTP_SERVER_AUTH_DISABLED", "")
	t.Setenv("HTTP_SERVER_API_KEYS", "")
	c, _ := config.NewConfig(nil)
	if _, err := NewAuthenticator(c); err == nil {
		t.Errorf("expected an error when no API keys are configured")
	}

	t.Setenv("HTTP_SERVER_API_KEYS", "alice@acme.example:user:"+digest(aliceKey)+", status-worker:worker+operator:"+digest(workerKey))
	c, _ = config.NewConfig(nil)
	a, err := NewAuthenticator(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+workerKey)
	p, err := a.authenticate(req)
	if err != nil || p.Identity != "status-worker" || !p.Roles[RoleWorker] || !p.Roles[RoleOperator] || p.Roles[RoleUser] {
		t.Errorf("unexpected principal %+v err=%v", p, err)
	}

	for _, bad := range []string{
		"alice:user",
		"alice:superuser:" + digest(aliceKey),
		"alice:user:not-hex",
		":user:" + digest(aliceKey),
		"alice:user:" + digest(aliceKey) + ",bob:user:" + digest(aliceKey),
	} {
		t.Setenv("HTTP_SERVER_API_KEYS", bad)
		c, _ = config.NewConfig(nil)
		if _, err := NewAuthenticator(c); err == nil {
			t.Errorf("expected error for entry %q", bad)
		}
	}

	t.Setenv("HTTP_SERVER_API_KEYS", "")
	t.Setenv("HTTP_SERVER_AUTH_DISABLED", "true")
	c, _ = config.NewConfig(nil)
	a, err = NewAuthenticator(c)
	if err != nil || !a.Disabled() {
		t.Errorf("expected explicit opt-out to disable auth, err=%v", err)
	}
}

func TestNewLaunchPolicy_Config(t *testing.T) {
	t.Setenv("EKS_DEFAULT_SERVICE_ACCOUNT", "flotilla-job")
	t.Setenv("EMR_DEFAULT_SERVICE_ACCOUNT", "flotilla-emr")
	t.Setenv("ALLOWED_SERVICE_ACCOUNTS", "etl-runner, ml-trainer")
	t.Setenv("ALLOWED_IMAGE_REGISTRIES", "registry.acme.example/")
	c, _ := config.NewConfig(nil)
	p := NewLaunchPolicy(c)

	for _, sa := range []string{"flotilla-job", "flotilla-emr", "etl-runner", "ml-trainer"} {
		s := sa
		if err := p.CheckServiceAccount(&state.ExecutionRequestCommon{ServiceAccount: &s}); err != nil {
			t.Errorf("expected %s to be allowed: %v", sa, err)
		}
	}
	bad := "cluster-admin"
	if err := p.CheckServiceAccount(&state.ExecutionRequestCommon{ServiceAccount: &bad}); err == nil {
		t.Errorf("expected cluster-admin to be rejected")
	}
	empty := ""
	common := &state.ExecutionRequestCommon{ServiceAccount: &empty}
	if err := p.CheckServiceAccount(common); err != nil || common.ServiceAccount != nil {
		t.Errorf("expected empty service account to be normalised to the engine default")
	}
	if err := p.CheckImage("registry.acme.example/a/b:1"); err != nil {
		t.Errorf("expected registry image to be allowed: %v", err)
	}
	if err := p.CheckImage("docker.io/library/alpine"); err == nil {
		t.Errorf("expected docker.io image to be rejected")
	}
}
