package services

import (
	"context"
	"strings"
	"testing"

	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/testutils"
)

func setUpTemplateServiceTest(t *testing.T) (TemplateService, *testutils.ImplementsAllTheThings) {
	memory := int64(512)
	imp := testutils.ImplementsAllTheThings{
		T: t,
		Templates: map[string]state.Template{
			"tpl-existing": {
				TemplateID:      "tpl-existing",
				TemplateName:    "existing",
				Version:         1,
				Schema:          state.TemplateJSONSchema{"type": "object"},
				CommandTemplate: "echo hi",
				ExecutableResources: state.ExecutableResources{
					Image:  "image:latest",
					Memory: &memory,
				},
			},
		},
	}
	ts, _ := NewTemplateService(nil, &imp)
	return ts, &imp
}

func TestTemplateService_Create_RejectsEnvFunction(t *testing.T) {
	ts, imp := setUpTemplateServiceTest(t)
	memory := int64(512)
	req := state.CreateTemplateRequest{
		TemplateName:    "existing",
		Schema:          state.TemplateJSONSchema{"type": "object"},
		CommandTemplate: `echo {{ env "AWS_SECRET_ACCESS_KEY" }}`,
		ExecutableResources: state.ExecutableResources{
			Image:  "image:latest",
			Memory: &memory,
		},
	}

	_, err := ts.Create(context.Background(), &req)
	if err == nil {
		t.Fatalf("expected template with env function to be rejected")
	}
	if _, ok := err.(exceptions.MalformedInput); !ok {
		t.Errorf("expected MalformedInput, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), `function "env" not defined`) {
		t.Errorf("unexpected error: %v", err)
	}
	for _, call := range imp.Calls {
		if call == "CreateTemplate" {
			t.Errorf("rejected template must not be stored")
		}
	}
}

func TestTemplateService_Create_DefaultsAdaptiveResourceAllocation(t *testing.T) {
	ts, imp := setUpTemplateServiceTest(t)
	memory := int64(512)
	req := state.CreateTemplateRequest{
		TemplateName:    "existing",
		Schema:          state.TemplateJSONSchema{"type": "object"},
		CommandTemplate: `echo {{ .name | quote }}`,
		ExecutableResources: state.ExecutableResources{
			Image:  "image:latest",
			Memory: &memory,
		},
	}

	res, err := ts.Create(context.Background(), &req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.DidCreate || res.Template.Version != 2 {
		t.Errorf("expected new version 2 to be created, got %+v", res)
	}
	ara := res.Template.AdaptiveResourceAllocation
	if ara == nil || *ara != true {
		t.Errorf("expected adaptive_resource_allocation to default to true, got %v", ara)
	}
	if _, ok := imp.Templates[res.Template.TemplateID]; !ok {
		t.Errorf("expected template to be stored")
	}
}
