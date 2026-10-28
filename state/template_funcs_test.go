package state

import (
	"os"
	"strings"
	"testing"

	"github.com/Masterminds/sprig"
)

func templateWithCommand(command string) Template {
	memory := int64(512)
	return Template{
		TemplateName:    "tpl",
		Schema:          TemplateJSONSchema{"type": "object"},
		CommandTemplate: command,
		Defaults:        TemplatePayload{},
		ExecutableResources: ExecutableResources{
			Image:  "image:latest",
			Memory: &memory,
		},
	}
}

func renderTemplate(tpl Template, payload TemplatePayload) (string, error) {
	return tpl.GetExecutableCommand(TemplateExecutionRequest{TemplatePayload: payload})
}

func TestTemplateFuncMap_AllowlistResolves(t *testing.T) {
	all := sprig.TxtFuncMap()
	for _, name := range templateFuncAllowlist {
		if _, ok := all[name]; !ok {
			t.Errorf("allowlisted function [%s] is not provided by sprig", name)
		}
	}
	if got, want := len(TemplateFuncMap()), len(templateFuncAllowlist); got != want {
		t.Errorf("expected %d functions, got %d", want, got)
	}
}

func TestTemplateFuncMap_ExcludesDangerousFunctions(t *testing.T) {
	funcs := TemplateFuncMap()
	for _, name := range []string{
		"env", "expandenv", "getHostByName", "genPrivateKey", "genCA",
		"genSelfSignedCert", "genSignedCert", "derivePassword", "bcrypt",
		"htpasswd", "repeat", "until", "untilStep",
	} {
		if _, ok := funcs[name]; ok {
			t.Errorf("function [%s] must not be available to templates", name)
		}
	}
}

func TestTemplate_IsValid_RejectsDisallowedFunctions(t *testing.T) {
	for _, command := range []string{
		`echo {{ env "HOME" }}`,
		`echo {{ expandenv "$HOME" }}`,
		`echo {{ getHostByName "localhost" }}`,
		`echo {{ "x" | env }}`,
	} {
		tpl := templateWithCommand(command)
		valid, reasons := tpl.IsValid()
		if valid {
			t.Errorf("expected command %q to be invalid", command)
			continue
		}
		joined := strings.Join(reasons, "\n")
		if !strings.Contains(joined, "[command_template] is invalid") || !strings.Contains(joined, "not defined") {
			t.Errorf("expected undefined function reason for %q, got %q", command, joined)
		}
	}
}

func TestTemplate_IsValid_AcceptsAllowedFunctions(t *testing.T) {
	tpl := templateWithCommand(`python run.py --name {{ .name | upper | quote }} --n {{ default 1 .n }}`)
	if valid, reasons := tpl.IsValid(); !valid {
		t.Errorf("expected template to be valid, got %v", reasons)
	}
}

func TestTemplate_GetExecutableCommand_DoesNotLeakEnvironment(t *testing.T) {
	os.Setenv("FLOTILLA_TEST_SECRET", "s3cr3t-value")
	defer os.Unsetenv("FLOTILLA_TEST_SECRET")

	// A template stored before validation existed must not render server env.
	tpl := templateWithCommand(`echo {{ env "FLOTILLA_TEST_SECRET" }}`)
	cmd, err := renderTemplate(tpl, TemplatePayload{})
	if err == nil {
		t.Fatalf("expected render to fail, got command %q", cmd)
	}
	if strings.Contains(cmd, "s3cr3t-value") || strings.Contains(err.Error(), "s3cr3t-value") {
		t.Errorf("server environment leaked into command or error")
	}
	if !strings.Contains(err.Error(), `function "env" not defined`) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTemplate_GetExecutableCommand_RendersAllowedFunctions(t *testing.T) {
	tpl := templateWithCommand(`echo {{ .name | upper }} {{ join "," .items }} {{ default "none" .missing }}`)
	cmd, err := renderTemplate(tpl, TemplatePayload{"name": "flotilla", "items": []interface{}{"a", "b"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != "echo FLOTILLA a,b none" {
		t.Errorf("unexpected command %q", cmd)
	}
}
