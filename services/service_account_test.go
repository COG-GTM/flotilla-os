package services

import (
	"context"
	"testing"

	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
)

func launchRequest(engine string, sa *string, labels *state.Labels) *state.DefinitionExecutionRequest {
	cmd := "_test_cmd_"
	extension := &state.SparkExtension{}
	if engine != state.EKSSparkEngine {
		extension = nil
	}
	description := "spark job"
	return &state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			ClusterName:    "clusta",
			OwnerID:        "mallory",
			Command:        &cmd,
			Engine:         &engine,
			SparkExtension: extension,
			Description:    &description,
			Labels:         labels,
			ServiceAccount: sa,
		},
	}
}

func strPtr(s string) *string { return &s }

func TestExecutionService_ServiceAccountAllowlist(t *testing.T) {
	dataPlatform := state.Labels{"team": "data-platform"}
	otherTeam := state.Labels{"team": "growth"}

	cases := []struct {
		name      string
		engine    string
		sa        *string
		labels    *state.Labels
		wantErr   bool
		wantSA    *string
		forbidden string
	}{
		{name: "eks unset uses engine default", engine: state.EKSEngine},
		{name: "eks blank is treated as unset", engine: state.EKSEngine, sa: strPtr("  ")},
		{name: "eks default allowed", engine: state.EKSEngine, sa: strPtr("flotilla-job"), wantSA: strPtr("flotilla-job")},
		{name: "eks global allowlist", engine: state.EKSEngine, sa: strPtr("flotilla-shared-job"), wantSA: strPtr("flotilla-shared-job")},
		{name: "eks team allowlist", engine: state.EKSEngine, sa: strPtr("data-platform-reader"), labels: &dataPlatform, wantSA: strPtr("data-platform-reader")},
		{name: "eks team allowlist case-insensitive team", engine: state.EKSEngine, sa: strPtr("data-platform-reader"), labels: &state.Labels{"team": "Data-Platform"}, wantSA: strPtr("data-platform-reader")},
		{name: "eks privileged account rejected", engine: state.EKSEngine, sa: strPtr("data-admin"), wantErr: true},
		{name: "eks other team's account rejected", engine: state.EKSEngine, sa: strPtr("data-platform-reader"), labels: &otherTeam, wantErr: true},
		{name: "eks team account rejected without team label", engine: state.EKSEngine, sa: strPtr("data-platform-reader"), wantErr: true},
		{name: "eks emr default is not an eks default", engine: state.EKSEngine, sa: strPtr("flotilla-spark-job"), wantErr: true},
		{name: "spark unset uses emr default", engine: state.EKSSparkEngine},
		{name: "spark team account with role", engine: state.EKSSparkEngine, sa: strPtr("data-platform-etl"), labels: &dataPlatform, wantSA: strPtr("data-platform-etl")},
		{name: "spark mapped role outside allowlist rejected", engine: state.EKSSparkEngine, sa: strPtr("data-platform-etl"), labels: &otherTeam, wantErr: true},
		{name: "spark allowlisted account without role rejected", engine: state.EKSSparkEngine, sa: strPtr("flotilla-shared-job"), wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			es, imp := setUp(t)
			run, err := es.CreateDefinitionRunByDefinitionID(context.Background(), "B", launchRequest(tc.engine, tc.sa, tc.labels))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected service_account %q to be rejected", *tc.sa)
				}
				if _, ok := err.(exceptions.MalformedInput); !ok {
					t.Fatalf("expected MalformedInput (HTTP 400), got %T: %v", err, err)
				}
				for _, call := range imp.Calls {
					if call == "CreateRun" || call == "Enqueue" {
						t.Fatalf("rejected request must not be persisted or queued, saw %s", call)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantSA == nil {
				if run.ServiceAccount != nil {
					t.Fatalf("expected nil service account so the engine default applies, got %q", *run.ServiceAccount)
				}
				return
			}
			if run.ServiceAccount == nil || *run.ServiceAccount != *tc.wantSA {
				t.Fatalf("expected service account %q, got %v", *tc.wantSA, run.ServiceAccount)
			}
		})
	}
}
