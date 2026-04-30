package install

import (
	"strings"
	"testing"
)

func TestRenderServiceUnitIncludesExecStart(t *testing.T) {
	t.Parallel()

	unit, err := RenderServiceUnit("/usr/local/bin/diskhm", "/etc/diskhm/config.yaml")
	if err != nil {
		t.Fatalf("RenderServiceUnit returned error: %v", err)
	}

	if !strings.Contains(unit, "ExecStart=/usr/local/bin/diskhm") {
		t.Fatalf("service unit missing ExecStart: %q", unit)
	}
}
