package flotilla

import (
	muxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
)

// NewRouter creates and returns a Mux Router. Every route except the health
// check requires an authenticated caller holding the listed role(s).
func NewRouter(ep endpoints) *muxtrace.Router {
	auth := ep.auth
	r := muxtrace.NewRouter()
	v1 := r.PathPrefix("/api/v1").Subrouter()

	v1.HandleFunc("/task", auth.Require(ep.ListDefinitions, RoleUser)).Methods("GET")
	v1.HandleFunc("/task", auth.Require(ep.CreateDefinition, RoleUser)).Methods("POST")
	v1.HandleFunc("/task/{definition_id}", auth.Require(ep.GetDefinition, RoleUser)).Methods("GET")
	v1.HandleFunc("/task/{definition_id}", auth.Require(ep.UpdateDefinition, RoleUser)).Methods("PUT")
	v1.HandleFunc("/task/{definition_id}", auth.Require(ep.DeleteDefinition, RoleUser)).Methods("DELETE")
	v1.HandleFunc("/task/{definition_id}/execute", auth.Require(ep.CreateRun, RoleUser)).Methods("PUT")
	v1.HandleFunc("/task/alias/{alias}", auth.Require(ep.GetDefinitionByAlias, RoleUser)).Methods("GET")
	v1.HandleFunc("/task/alias/{alias}/execute", auth.Require(ep.CreateRunByAlias, RoleUser)).Methods("PUT")

	v1.HandleFunc("/history", auth.Require(ep.ListRuns, RoleUser)).Methods("GET")
	v1.HandleFunc("/history/{run_id}", auth.Require(ep.GetRun, RoleUser)).Methods("GET")
	v1.HandleFunc("/task/history/{run_id}", auth.Require(ep.GetRun, RoleUser)).Methods("GET")
	v1.HandleFunc("/task/{definition_id}/history", auth.Require(ep.ListDefinitionRuns, RoleUser)).Methods("GET")
	v1.HandleFunc("/task/{definition_id}/history/{run_id}", auth.Require(ep.GetRun, RoleUser)).Methods("GET")
	v1.HandleFunc("/task/{definition_id}/history/{run_id}", auth.Require(ep.StopRun, RoleUser)).Methods("DELETE")

	v1.HandleFunc("/{run_id}/status", auth.Require(ep.UpdateRun, RoleWorker, RoleOperator)).Methods("PUT")
	v1.HandleFunc("/{run_id}/logs", auth.Require(ep.GetLogs, RoleUser)).Methods("GET")
	v1.HandleFunc("/{run_id}/events", auth.Require(ep.GetEvents, RoleUser)).Methods("GET")
	v1.HandleFunc("/groups", auth.Require(ep.GetGroups, RoleUser)).Methods("GET")
	v1.HandleFunc("/tags", auth.Require(ep.GetTags, RoleUser)).Methods("GET")
	v1.HandleFunc("/clusters", auth.Require(ep.ListClusters, RoleUser)).Methods("GET")

	v2 := r.PathPrefix("/api/v2").Subrouter()
	v2.HandleFunc("/task/{definition_id}/execute", auth.Require(ep.CreateRunV2, RoleUser)).Methods("PUT")

	v4 := r.PathPrefix("/api/v4").Subrouter()
	v4.HandleFunc("/task/{definition_id}/execute", auth.Require(ep.CreateRunV4, RoleUser)).Methods("PUT")

	v5 := r.PathPrefix("/api/v5").Subrouter()
	v5.HandleFunc("/worker", auth.Require(ep.ListWorkers, RoleUser)).Methods("GET")
	v5.HandleFunc("/worker", auth.Require(ep.BatchUpdateWorkers, RoleOperator)).Methods("PUT")
	v5.HandleFunc("/worker/{worker_type}", auth.Require(ep.GetWorker, RoleUser)).Methods("GET")
	v5.HandleFunc("/worker/{worker_type}", auth.Require(ep.UpdateWorker, RoleOperator)).Methods("PUT")

	v6 := r.PathPrefix("/api/v6").Subrouter()
	v6.HandleFunc("/clusters", auth.Require(ep.ListClusters, RoleUser)).Methods("GET")
	v6.HandleFunc("/clusters", auth.Require(ep.CreateCluster, RoleOperator)).Methods("POST")
	v6.HandleFunc("/clusters/{cluster_id}", auth.Require(ep.GetCluster, RoleUser)).Methods("GET")
	v6.HandleFunc("/clusters/{cluster_id}", auth.Require(ep.UpdateCluster, RoleOperator)).Methods("PUT")
	v6.HandleFunc("/clusters/{cluster_id}", auth.Require(ep.DeleteCluster, RoleOperator)).Methods("DELETE")
	v6.HandleFunc("/{run_id}/events", auth.Require(ep.GetEvents, RoleUser)).Methods("GET")
	v6.HandleFunc("/groups", auth.Require(ep.GetGroups, RoleUser)).Methods("GET")
	v6.HandleFunc("/health", ep.HealthCheck).Methods("GET")
	v6.HandleFunc("/history", auth.Require(ep.ListRuns, RoleUser)).Methods("GET")
	v6.HandleFunc("/history/{run_id}", auth.Require(ep.GetRun, RoleUser)).Methods("GET")
	v6.HandleFunc("/tags", auth.Require(ep.GetTags, RoleUser)).Methods("GET")
	v6.HandleFunc("/task", auth.Require(ep.ListDefinitions, RoleUser)).Methods("GET")
	v6.HandleFunc("/task", auth.Require(ep.CreateDefinition, RoleUser)).Methods("POST")
	v6.HandleFunc("/task/alias/{alias}", auth.Require(ep.GetDefinitionByAlias, RoleUser)).Methods("GET")
	v6.HandleFunc("/task/alias/{alias}/execute", auth.Require(ep.CreateRunByAlias, RoleUser)).Methods("PUT")
	v6.HandleFunc("/task/{definition_id}", auth.Require(ep.GetDefinition, RoleUser)).Methods("GET")
	v6.HandleFunc("/task/{definition_id}", auth.Require(ep.UpdateDefinition, RoleUser)).Methods("PUT")
	v6.HandleFunc("/task/{definition_id}", auth.Require(ep.DeleteDefinition, RoleUser)).Methods("DELETE")
	v6.HandleFunc("/task/{definition_id}/execute", auth.Require(ep.CreateRunV4, RoleUser)).Methods("PUT")
	v6.HandleFunc("/task/{definition_id}/history", auth.Require(ep.ListDefinitionRuns, RoleUser)).Methods("GET")
	v6.HandleFunc("/task/{definition_id}/history/{run_id}", auth.Require(ep.GetRun, RoleUser)).Methods("GET")
	v6.HandleFunc("/task/{definition_id}/history/{run_id}", auth.Require(ep.StopRun, RoleUser)).Methods("DELETE")
	v6.HandleFunc("/task/history/{run_id}", auth.Require(ep.GetRun, RoleUser)).Methods("GET")
	v6.HandleFunc("/{run_id}/status", auth.Require(ep.UpdateRun, RoleWorker, RoleOperator)).Methods("PUT")
	v6.HandleFunc("/{run_id}/status", auth.Require(ep.GetRunStatus, RoleUser, RoleWorker, RoleOperator)).Methods("GET")
	v6.HandleFunc("/{run_id}/logs", auth.Require(ep.GetLogs, RoleUser)).Methods("GET")

	v7 := r.PathPrefix("/api/v7").Subrouter()
	v7.HandleFunc("/template/{template_id}/execute", auth.Require(ep.CreateTemplateRun, RoleUser)).Methods("PUT")
	v7.HandleFunc("/template/name/{template_name}/version/{template_version}/execute", auth.Require(ep.CreateTemplateRunByName, RoleUser)).Methods("PUT")
	v7.HandleFunc("/template", auth.Require(ep.ListTemplates, RoleUser)).Methods("GET")
	v7.HandleFunc("/template", auth.Require(ep.CreateTemplate, RoleUser)).Methods("POST")
	v7.HandleFunc("/template/{template_id}", auth.Require(ep.GetTemplate, RoleUser)).Methods("GET")
	v7.HandleFunc("/template/history/{run_id}", auth.Require(ep.GetRun, RoleUser)).Methods("GET")
	v7.HandleFunc("/template/{template_id}/history", auth.Require(ep.ListTemplateRuns, RoleUser)).Methods("GET")
	v7.HandleFunc("/template/{template_id}/history/{run_id}", auth.Require(ep.GetRun, RoleUser)).Methods("GET")
	v7.HandleFunc("/template/{template_id}/history/{run_id}", auth.Require(ep.StopRun, RoleUser)).Methods("DELETE")

	return r
}
