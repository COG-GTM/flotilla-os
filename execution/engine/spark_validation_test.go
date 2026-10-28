package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
)

type oomManager struct {
	state.Manager
	executorOOM bool
	driverOOM   bool
}

func (m oomManager) ExecutorOOM(_ context.Context, _ string, _ string) (bool, error) {
	return m.executorOOM, nil
}

func (m oomManager) DriverOOM(_ context.Context, _ string, _ string) (bool, error) {
	return m.driverOOM, nil
}

func sparkRun(confs ...state.Conf) state.Run {
	return state.Run{
		RunID:       "eks-spark-run",
		CommandHash: aws.String("hash"),
		SparkExtension: &state.SparkExtension{
			SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
				EntryPoint:      aws.String("s3://bucket/job.py"),
				SparkSubmitConf: confs,
			},
		},
	}
}

func TestValidateSparkExtension(t *testing.T) {
	tests := []struct {
		name    string
		se      *state.SparkExtension
		wantErr string
	}{
		{"nil", nil, "spark_extension is required"},
		{"empty", &state.SparkExtension{}, "spark_submit_job_driver is required"},
		{"nil conf name", sparkRun(state.Conf{Value: aws.String("1")}).SparkExtension, "spark_submit_conf[0] must have a name and a value"},
		{"nil conf value", sparkRun(state.Conf{Name: aws.String("spark.executor.cores")}).SparkExtension, "spark_submit_conf[0] must have a name and a value"},
		{"bad executor memory", sparkRun(state.Conf{Name: aws.String("spark.executor.memory"), Value: aws.String("abc")}).SparkExtension, "invalid memory value"},
		{"bad driver memory", sparkRun(state.Conf{Name: aws.String("spark.driver.memory"), Value: aws.String("4gb")}).SparkExtension, "invalid memory value"},
		{"nil application conf name", &state.SparkExtension{
			SparkSubmitJobDriver: &state.SparkSubmitJobDriver{},
			ApplicationConf:      []state.Conf{{Value: aws.String("x")}},
		}, "application_conf[0] must have a name and a value"},
		{"valid", sparkRun(
			state.Conf{Name: aws.String("spark.executor.memory"), Value: aws.String("4g")},
			state.Conf{Name: aws.String("spark.driver.memory"), Value: aws.String("2048m")},
			state.Conf{Name: aws.String("spark.executor.cores"), Value: aws.String("4")},
		).SparkExtension, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSparkExtension(tt.se)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if _, ok := err.(exceptions.MalformedInput); !ok {
				t.Errorf("expected exceptions.MalformedInput, got %T", err)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestEMRExecuteRejectsMalformedSparkExtension(t *testing.T) {
	emr := newTestEMREngine()
	runs := map[string]state.Run{
		"empty spark_extension": {RunID: "r1", CommandHash: aws.String("h"), SparkExtension: &state.SparkExtension{}},
		"nil spark_extension":   {RunID: "r2"},
		"unparseable memory":    sparkRun(state.Conf{Name: aws.String("spark.executor.memory"), Value: aws.String("abc")}),
		"nil conf name":         sparkRun(state.Conf{Value: aws.String("abc")}),
	}
	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			out, retryable, err := emr.Execute(context.Background(), state.Definition{}, run, oomManager{})
			if err == nil {
				t.Fatal("expected an error")
			}
			if retryable {
				t.Error("expected a non-retryable error")
			}
			if out.Status != state.StatusStopped {
				t.Errorf("expected status %s, got %s", state.StatusStopped, out.Status)
			}
			if out.ExitReason == nil || *out.ExitReason != err.Error() {
				t.Errorf("expected exit reason %q, got %v", err.Error(), out.ExitReason)
			}
		})
	}
}

func TestEstimateMemoryResourcesToleratesMalformedConf(t *testing.T) {
	emr := newTestEMREngine()
	for _, m := range []oomManager{{}, {executorOOM: true, driverOOM: true}} {
		run := sparkRun(
			state.Conf{Value: aws.String("orphan")},
			state.Conf{Name: aws.String("spark.executor.memory"), Value: aws.String("abc")},
			state.Conf{Name: aws.String("spark.driver.memory"), Value: aws.String("xyz")},
			state.Conf{Name: aws.String("spark.executor.cores")},
		)
		out := emr.estimateMemoryResources(context.Background(), run, m)
		confs := out.SparkExtension.SparkSubmitJobDriver.SparkSubmitConf
		if len(confs) != 4 {
			t.Fatalf("expected 4 confs to be preserved, got %d", len(confs))
		}
		if *confs[1].Value != "abc" || *confs[2].Value != "xyz" {
			t.Errorf("expected unparseable values to be left unchanged, got %q and %q", *confs[1].Value, *confs[2].Value)
		}
	}

	noDriver := state.Run{RunID: "r", CommandHash: aws.String("h"), SparkExtension: &state.SparkExtension{}}
	emr.estimateMemoryResources(context.Background(), noDriver, oomManager{})
}

func TestEstimateMemoryResourcesStillAdjustsValidMemory(t *testing.T) {
	emr := newTestEMREngine()
	run := sparkRun(
		state.Conf{Name: aws.String("spark.executor.memory"), Value: aws.String("4g")},
		state.Conf{Name: aws.String("spark.driver.memory"), Value: aws.String("1g")},
	)
	out := emr.estimateMemoryResources(context.Background(), run, oomManager{executorOOM: true, driverOOM: true})
	confs := out.SparkExtension.SparkSubmitJobDriver.SparkSubmitConf
	if *confs[0].Value != "5g" {
		t.Errorf("expected executor memory 5g after 1.25x, got %q", *confs[0].Value)
	}
	if *confs[1].Value != "3g" {
		t.Errorf("expected driver memory 3g after 3x, got %q", *confs[1].Value)
	}
}

func TestSparkSubmitParamsSkipsIncompleteConf(t *testing.T) {
	emr := newTestEMREngine()
	run := sparkRun(
		state.Conf{Value: aws.String("orphan")},
		state.Conf{Name: aws.String("spark.executor.cores")},
		state.Conf{Name: aws.String("spark.executor.instances"), Value: aws.String("2")},
	)
	params := *emr.sparkSubmitParams(run)
	if !strings.Contains(params, "--conf spark.executor.instances=2") {
		t.Errorf("expected valid conf in params, got %q", params)
	}
	if strings.Contains(params, "orphan") || strings.Contains(params, "spark.executor.cores") {
		t.Errorf("expected incomplete confs to be skipped, got %q", params)
	}

	emr.sparkSubmitParams(state.Run{RunID: "r", SparkExtension: &state.SparkExtension{}})
	emr.sparkSubmitParams(state.Run{RunID: "r"})
}
