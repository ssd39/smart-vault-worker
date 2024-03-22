//go:build sidecar

package main

import "github.com/ssd39/smart-vault-sgx-app/cmd"

func main() {
	cmd.StartSidecar()
}
