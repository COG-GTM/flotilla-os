package flotilla

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stitchfix/flotilla-os/clients/middleware"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/services"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/testutils"
	muxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
)

const testClusterAdminToken = "test-cluster-admin-token"

func setUp(t *testing.T) *muxtrace.Router {
	ep, _ := setUpEndpoints(t)
	return NewRouter(ep)
}

func setUpEndpoints(t *testing.T) (endpoints, *testutils.ImplementsAllTheThings) {
	confDir := "../conf"
	c, _ := config.NewConfig(&confDir)
	imp := testutils.ImplementsAllTheThings{
		T: t,
		Definitions: map[string]state.Definition{
			"A": {DefinitionID: "A", Alias: "aliasA"},
			"B": {DefinitionID: "B", Alias: "aliasB"},
			"C": {DefinitionID: "C", Alias: "aliasC", ExecutableResources: state.ExecutableResources{Image: "invalidimage"}},
		},
		Runs: map[string]state.Run{
			"runA": {DefinitionID: "A", ClusterName: "cluster1",
				GroupName: "A",
				RunID:     "runA", Status: state.StatusRunning},
			"runB": {DefinitionID: "B", ClusterName: "cluster2",
				GroupName: "B", RunID: "runB",
				InstanceDNSName: "cupcakedns", InstanceID: "cupcakeid"},
		},
		Qurls: map[string]string{
			"A": "a/",
			"B": "b/",
		},
		ClusterStates: []state.ClusterMetadata{
			{Name: "cluster1", Status: state.StatusActive, StatusReason: "Active and healthy"},
			{Name: "cluster2", Status: state.StatusActive, StatusReason: "Active and healthy"},
		},
		Groups: []string{"g1", "g2", "g3"},
		Tags:   []string{"t1", "t2", "t3"},
	}
	ds, _ := services.NewDefinitionService(&imp)
	es, _ := services.NewExecutionService(c, &imp, &imp, &imp, &imp)
	ls, _ := services.NewLogService(&imp, &imp)
	mwc, _ := middleware.NewClient()
	return endpoints{definitionService: ds, executionService: es, eksLogService: ls, middlewareClient: mwc,
		clusterAdminToken: testClusterAdminToken}, &imp
}

func TestEndpoints_CreateDefinition(t *testing.T) {
	router := setUp(t)

	newDef := `{"alias":"cupcake", "memory":100, "group_name":"cupcake", "image":"someimage", "command":"echo 'hi'"}`
	req := httptest.NewRequest("POST", "/api/v1/task", bytes.NewBufferString(newDef))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Definition{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if len(r.DefinitionID) == 0 {
		t.Errorf("Expected non-empty definition id")
	}
}

func TestEndpoints_UpdateDefinition(t *testing.T) {
	router := setUp(t)

	updatedDef := `{"image":"updatedImage"}`
	req := httptest.NewRequest("PUT", "/api/v1/task/A", bytes.NewBufferString(updatedDef))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Definition{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if r.Image != "updatedImage" {
		t.Errorf("Expected image [updatedImage] but was [%s]", r.Image)
	}
}

func TestEndpoints_CreateRun(t *testing.T) {
	router := setUp(t)

	newRun := `{"cluster":"cupcake", "env":[{"name":"E1","value":"V1"}]}`
	req := httptest.NewRequest("PUT", "/api/v1/task/A/execute", bytes.NewBufferString(newRun))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Run{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if len(r.RunID) == 0 {
		t.Errorf("Expected non-empty run id")
	}

	if r.Status != state.StatusQueued {
		t.Errorf("Expected new run to have status [%s] but was [%s]", state.StatusQueued, r.Status)
	}
}

func TestEndpoints_CreateRun2(t *testing.T) {
	router := setUp(t)

	newRun := `{"cluster":"cupcake", "env":[{"name":"E1","value":"V1"}], "run_tags":{"owner_email":"flotilla@github.com", "team_name":"thebest"}}`
	req := httptest.NewRequest("PUT", "/api/v2/task/A/execute", bytes.NewBufferString(newRun))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Run{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if len(r.RunID) == 0 {
		t.Errorf("Expected non-empty run id")
	}

	if r.Status != state.StatusQueued {
		t.Errorf("Expected new run to have status [%s] but was [%s]", state.StatusQueued, r.Status)
	}

	if r.User != "flotilla@github.com" {
		t.Errorf("Expected new run to have user set to run_tags.owner_email but was [%s]", r.User)
	}
}

func TestEndpoints_CreateRun4(t *testing.T) {
	router := setUp(t)

	newRun := `{"cluster":"cluster1", "env":[{"name":"E1","value":"V1"}], "run_tags":{"owner_id":"flotilla"}, "labels": {"foo": "bar"}}`
	req := httptest.NewRequest("PUT", "/api/v4/task/A/execute", bytes.NewBufferString(newRun))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v\n%s", resp.StatusCode, resp.Status)
	}

	r := state.Run{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if len(r.RunID) == 0 {
		t.Errorf("Expected non-empty run id")
	}

	if r.Status != state.StatusQueued {
		t.Errorf("Expected new run to have status [%s] but was [%s]", state.StatusQueued, r.Status)
	}

	if len(r.Labels) != 1 || r.Labels["foo"] != "bar" {
		labelRes, _ := json.Marshal(r.Labels)
		t.Error(string(labelRes))
	}

	if r.User != "flotilla" {
		t.Errorf("Expected new run to have user set to run_tags.owner_id but was [%s]", r.User)
	}
}

func TestEndpoints_CreateRunByAlias(t *testing.T) {
	router := setUp(t)

	newRun := `{"cluster":"cupcake", "env":[{"name":"E1","value":"V1"}], "run_tags":{"owner_id":"flotilla"}}`
	req := httptest.NewRequest("PUT", "/api/v1/task/alias/aliasA/execute", bytes.NewBufferString(newRun))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Run{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if len(r.RunID) == 0 {
		t.Errorf("Expected non-empty run id")
	}

	if r.Status != state.StatusQueued {
		t.Errorf("Expected new run to have status [%s] but was [%s]", state.StatusQueued, r.Status)
	}

	if r.User != "flotilla" {
		t.Errorf("Expected new run to have user set to run_tags.owner_id but was [%s]", r.User)
	}
}

func TestEndpoints_DeleteDefinition(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("DELETE", "/api/v1/task/A", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var ack map[string]bool
	err := json.NewDecoder(resp.Body).Decode(&ack)
	if err != nil {
		t.Error(err.Error())
	}
	if _, ok := ack["deleted"]; !ok {
		t.Errorf("Expected [deleted] acknowledgement")
	}
}

func TestEndpoints_GetDefinition(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/task/A", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r state.Definition
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if r.DefinitionID != "A" {
		t.Errorf("Expected definition_id [A] but was [%s]", r.DefinitionID)
	}

	if r.Env == nil {
		t.Errorf("Expected non-nil environment")
	}
}

func TestEndpoints_GetDefinitionByAlias(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/task/alias/aliasA", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r state.Definition
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if r.DefinitionID != "A" {
		t.Errorf("Expected definition_id [A] but was [%s]", r.DefinitionID)
	}

	if r.Env == nil {
		t.Errorf("Expected non-nil environment")
	}
}

func TestEndpoints_GetGroups(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/groups", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if _, ok := r["total"]; !ok {
		t.Errorf("Expected total in response")
	}

	if _, ok := r["groups"]; !ok {
		t.Errorf("Expected groups in response")
	}

	groups, _ := r["groups"]
	if _, ok := groups.([]interface{}); !ok {
		t.Errorf("Cannot cast groups to list, expected list")
	}
}

func TestEndpoints_GetLogs(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/runA/logs", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if _, ok := r["log"]; !ok {
		t.Errorf("Expected log in response")
	}
}

func TestEndpoints_GetRun(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/history/runA", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r state.Run
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if r.RunID != "runA" {
		t.Errorf("Expected run with runID [runA] but was [%s]", r.RunID)
	}
}

func TestEndpoints_GetRun2(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/history/runB", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var other map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&other)
	if err != nil {
		t.Error(err.Error())
	}

	instance, ok := other["instance"]
	if !ok {
		t.Errorf("Expected [instance] in response")
	}

	if _, ok = instance.(map[string]interface{}); !ok {
		t.Errorf("Expected [instance] in response to be a map")
	}
}

func TestEndpoints_GetTags(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/tags", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if _, ok := r["total"]; !ok {
		t.Errorf("Expected total in response")
	}

	if _, ok := r["tags"]; !ok {
		t.Errorf("Expected tags in response")
	}

	tags, _ := r["tags"]
	if _, ok := tags.([]interface{}); !ok {
		t.Errorf("Cannot cast tags to list, expected list")
	}
}

func TestEndpoints_ListDefinitions(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/task?limit=100&offset=2&sort_by=alias&order=desc&group_name=cupcake&env=E1%7CV1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if _, ok := r["total"]; !ok {
		t.Errorf("Expected total in response")
	}

	if _, ok := r["definitions"]; !ok {
		t.Errorf("Expected definitions in response")
	}

	if _, ok := r["limit"]; !ok {
		t.Errorf("Expected limit in response")
	}

	if _, ok := r["offset"]; !ok {
		t.Errorf("Expected offset in response")
	}

	if _, ok := r["sort_by"]; !ok {
		t.Errorf("Expected sort_by in response")
	}

	if _, ok := r["order"]; !ok {
		t.Errorf("Expected order in response")
	}

	if _, ok := r["group_name"]; !ok {
		t.Errorf("Expected [group_name] filter in response")
	}

	if _, ok := r["env_filters"]; !ok {
		t.Errorf("Expected env_filters in response")
	}

	definitions, _ := r["definitions"]
	if _, ok := definitions.([]interface{}); !ok {
		t.Errorf("Cannot cast definitions to list, expected list")
	}

	envFilters, _ := r["env_filters"]
	if _, ok := envFilters.(map[string]interface{}); !ok {
		t.Errorf("Cannot cast env_filters to map, expected map")
	}

	envFiltersMap := envFilters.(map[string]interface{})
	e1Filter, ok := envFiltersMap["E1"]
	if !ok {
		t.Errorf("Expected env_filters to contain key [E1]")
	}

	if e1Filter.(string) != "V1" {
		t.Errorf("Expected env_filter [E1:V1]")
	}
}

func TestEndpoints_ListRuns(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest(
		"GET",
		"/api/v1/history?status=RUNNING&status=QUEUED&limit=100&offset=2&sort_by=started_at&order=desc&cluster=cupcake&env=E1%7CV1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Error(err.Error())
	}

	if _, ok := r["total"]; !ok {
		t.Errorf("Expected total in response")
	}

	if _, ok := r["history"]; !ok {
		t.Errorf("Expected runs in response")
	}

	if _, ok := r["limit"]; !ok {
		t.Errorf("Expected limit in response")
	}

	if _, ok := r["offset"]; !ok {
		t.Errorf("Expected offset in response")
	}

	if _, ok := r["sort_by"]; !ok {
		t.Errorf("Expected sort_by in response")
	}

	if _, ok := r["order"]; !ok {
		t.Errorf("Expected order in response")
	}

	if _, ok := r["cluster"]; !ok {
		t.Errorf("Expected [cluster] filter in response")
	}

	if _, ok := r["env_filters"]; !ok {
		t.Errorf("Expected env_filters in response")
	}

	if _, ok := r["status"]; !ok {
		t.Errorf("Expected [status] filter in response")
	}

	runs, _ := r["history"]
	if _, ok := runs.([]interface{}); !ok {
		t.Errorf("Cannot cast runs to list, expected list")
	}

	statusFilters, _ := r["status"]
	if _, ok := statusFilters.([]interface{}); !ok {
		t.Errorf("Cannot cast status filters to list, expected list")
	}

	expectedStatusFilters := map[string]bool{"RUNNING": true, "QUEUED": true}
	statusFiltersList := statusFilters.([]interface{})
	if len(statusFiltersList) != 2 {
		t.Errorf("Expected 2 status filters, was %v", len(statusFiltersList))
	}
	for _, statusFilter := range statusFiltersList {
		if _, ok := expectedStatusFilters[statusFilter.(string)]; !ok {
			t.Errorf("Unexpected status filter: %s", statusFilter.(string))
		}
	}

	envFilters, _ := r["env_filters"]
	if _, ok := envFilters.(map[string]interface{}); !ok {
		t.Errorf("Cannot cast env_filters to map, expected map")
	}

	envFiltersMap := envFilters.(map[string]interface{})
	e1Filter, ok := envFiltersMap["E1"]
	if !ok {
		t.Errorf("Expected env_filters to contain key [E1]")
	}

	if e1Filter.(string) != "V1" {
		t.Errorf("Expected env_filter [E1:V1]")
	}
}

func TestEndpoints_StopRun(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("DELETE", "/api/v1/task/A/history/runA", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var ack map[string]bool
	err := json.NewDecoder(resp.Body).Decode(&ack)
	if err != nil {
		t.Error(err.Error())
	}
	if _, ok := ack["terminated"]; !ok {
		t.Errorf("Expected [terminated] acknowledgement")
	}
}

func TestEndpoints_ListClusters(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v6/clusters", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Error(err.Error())
	}

	clusters, ok := response["clusters"]
	if !ok {
		t.Errorf("Expected clusters in response")
	}

	clustersList, ok := clusters.([]interface{})
	if !ok {
		t.Errorf("Cannot cast clusters to list, expected list")
	}

	if len(clustersList) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(clustersList))
	}

	cluster, ok := clustersList[0].(map[string]interface{})
	if !ok {
		t.Errorf("Cannot cast cluster to map, expected map")
	}

	if _, ok := cluster["name"]; !ok {
		t.Errorf("Expected cluster to have name field")
	}

	if _, ok := cluster["status"]; !ok {
		t.Errorf("Expected cluster to have status field")
	}
}

func TestEndpoints_GetCluster(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v6/clusters/cluster1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var cluster map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&cluster)
	if err != nil {
		t.Error(err.Error())
	}

	if _, ok := cluster["name"]; !ok {
		t.Errorf("Expected cluster to have name field")
	}

	if _, ok := cluster["status"]; !ok {
		t.Errorf("Expected cluster to have status field")
	}
}

func TestEndpoints_UpdateCluster(t *testing.T) {
	router := setUp(t)

	updateReq := `{"name":"cluster1", "status":"ACTIVE", "status_reason":"Testing update", "allowed_tiers":["1","4"]}`
	req := httptest.NewRequest("PUT", "/api/v6/clusters/cluster1", bytes.NewBufferString(updateReq))
	req.Header.Set("Authorization", "Bearer "+testClusterAdminToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var ack map[string]bool
	err := json.NewDecoder(resp.Body).Decode(&ack)
	if err != nil {
		t.Error(err.Error())
	}

	if _, ok := ack["updated"]; !ok {
		t.Errorf("Expected [updated] acknowledgement")
	}
}

func TestEndpoints_DeleteCluster(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("DELETE", "/api/v6/clusters/cluster1", nil)
	req.Header.Set("Authorization", "Bearer "+testClusterAdminToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var ack map[string]bool
	err := json.NewDecoder(resp.Body).Decode(&ack)
	if err != nil {
		t.Error(err.Error())
	}

	if _, ok := ack["deleted"]; !ok {
		t.Errorf("Expected [deleted] acknowledgement")
	}
}

func TestEndpoints_CreateCluster(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("POST", "/api/v6/clusters", bytes.NewBufferString(`{"name":"cluster1", "status":"ACTIVE", "status_reason":"Testing create"}`))
	req.Header.Set("Authorization", "Bearer "+testClusterAdminToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var ack map[string]bool
	err := json.NewDecoder(resp.Body).Decode(&ack)
	if err != nil {
		t.Error(err.Error())
	}

	if _, ok := ack["created"]; !ok {
		t.Errorf("Expected [created] acknowledgement")
	}
}

func TestEndpoints_ClusterWritesRequireAdminToken(t *testing.T) {
	body := `{"name":"cluster3", "status":"active", "allowed_tiers":["1","2","3","4"]}`
	cases := []struct {
		name       string
		method     string
		path       string
		auth       string
		adminToken string
	}{
		{"create without token", "POST", "/api/v6/clusters", "", testClusterAdminToken},
		{"update without token", "PUT", "/api/v6/clusters/cluster1", "", testClusterAdminToken},
		{"delete without token", "DELETE", "/api/v6/clusters/cluster1", "", testClusterAdminToken},
		{"create with wrong token", "POST", "/api/v6/clusters", "Bearer wrong", testClusterAdminToken},
		{"create with non-bearer auth", "POST", "/api/v6/clusters", testClusterAdminToken, testClusterAdminToken},
		{"create when no token configured", "POST", "/api/v6/clusters", "Bearer ", ""},
		{"delete when no token configured", "DELETE", "/api/v6/clusters/cluster1", "Bearer ", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ep, imp := setUpEndpoints(t)
			ep.clusterAdminToken = tc.adminToken
			router := NewRouter(ep)

			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(body))
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != 401 {
				t.Errorf("Expected status 401, was %v", w.Code)
			}
			if w.Header().Get("WWW-Authenticate") == "" {
				t.Errorf("Expected WWW-Authenticate header on 401")
			}
			for _, call := range imp.Calls {
				if call == "UpdateClusterMetadata" || call == "DeleteClusterMetadata" {
					t.Errorf("Expected registry to be untouched, but %s was called", call)
				}
			}
		})
	}
}

func TestEndpoints_ClusterReadsStayOpen(t *testing.T) {
	router := setUp(t)
	for _, path := range []string{"/api/v6/clusters", "/api/v6/clusters/cluster1", "/api/v1/clusters"} {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Errorf("GET %s: expected status 200, was %v", path, w.Code)
		}
	}
}

func TestEndpoints_ClusterWritesRejectInvalidMetadata(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"path traversal name", "POST", "/api/v6/clusters", `{"name":"../x", "status":"active"}`},
		{"slash in name", "PUT", "/api/v6/clusters/cluster1", `{"name":"a/b", "status":"active"}`},
		{"empty name", "PUT", "/api/v6/clusters/cluster1", `{"status":"active"}`},
		{"option-like name", "POST", "/api/v6/clusters", `{"name":"--region", "status":"active"}`},
		{"unknown status", "POST", "/api/v6/clusters", `{"name":"cluster3", "status":"hijacked"}`},
		{"unknown tier", "POST", "/api/v6/clusters", `{"name":"cluster3", "status":"active", "allowed_tiers":["1","99"]}`},
		{"bad namespace", "POST", "/api/v6/clusters", `{"name":"cluster3", "status":"active", "namespace":"../kube-system"}`},
		{"bad region", "POST", "/api/v6/clusters", `{"name":"cluster3", "status":"active", "region":"us-east-1 --profile x"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ep, imp := setUpEndpoints(t)
			router := NewRouter(ep)
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			req.Header.Set("Authorization", "Bearer "+testClusterAdminToken)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != 400 {
				t.Errorf("Expected status 400, was %v (%s)", w.Code, w.Body.String())
			}
			for _, call := range imp.Calls {
				if call == "UpdateClusterMetadata" {
					t.Errorf("Expected invalid metadata not to be written")
				}
			}
		})
	}
}

func TestApp_CORSAllowsAuthorizedClusterWrites(t *testing.T) {
	const origin = "http://flotilla-ui.example"
	ep, _ := setUpEndpoints(t)
	app := App{corsAllowedOrigins: []string{origin}}
	app.configureRoutes(ep)

	preflight := httptest.NewRequest("OPTIONS", "/api/v6/clusters", nil)
	preflight.Header.Set("Origin", origin)
	preflight.Header.Set("Access-Control-Request-Method", "POST")
	preflight.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	w := httptest.NewRecorder()
	app.handler.ServeHTTP(w, preflight)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Errorf("Expected preflight to allow origin %s, got %q", origin, got)
	}
	if got := strings.ToLower(w.Header().Get("Access-Control-Allow-Headers")); !strings.Contains(got, "authorization") {
		t.Errorf("Expected preflight to allow the Authorization header, got %q", got)
	}

	req := httptest.NewRequest("POST", "/api/v6/clusters", bytes.NewBufferString(`{"name":"cluster3", "status":"active"}`))
	req.Header.Set("Origin", origin)
	req.Header.Set("Authorization", "Bearer "+testClusterAdminToken)
	w = httptest.NewRecorder()
	app.handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, was %v (%s)", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Errorf("Expected response to allow origin %s, got %q", origin, got)
	}
}
