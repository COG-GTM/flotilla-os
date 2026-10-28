package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetOrCreateKubeconfigRejectsInvalidClusterNames(t *testing.T) {
	base := t.TempDir()
	t.Setenv("EKS_KUBECONFIG_BASEPATH", base)
	t.Chdir(base)

	dcm := &DynamicClusterManager{awsRegion: "us-east-1"}
	for _, name := range []string{"../x", "../../etc/passwd", "a/b", "", "--profile", "/abs"} {
		if _, err := dcm.getOrCreateKubeconfig(name); err == nil {
			t.Errorf("Expected error for cluster name %q", name)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(base), "x")); !os.IsNotExist(err) {
		t.Errorf("Expected no file to be created outside the kubeconfig directory")
	}
}
