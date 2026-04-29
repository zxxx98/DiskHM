package power

import (
	"context"
	"os/exec"
)

type Executor struct{}

func (Executor) Standby(ctx context.Context, devicePath string) error {
	cmd := exec.CommandContext(ctx, "hdparm", "-y", devicePath)
	return cmd.Run()
}

func (Executor) CheckState(ctx context.Context, devicePath string) error {
	cmd := exec.CommandContext(ctx, "hdparm", "-C", devicePath)
	return cmd.Run()
}
