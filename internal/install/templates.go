package install

import "fmt"

func RenderServiceUnit(binaryPath, configPath string) (string, error) {
	return fmt.Sprintf(`[Unit]
Description=DiskHM daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s --config %s
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`, binaryPath, configPath), nil
}
