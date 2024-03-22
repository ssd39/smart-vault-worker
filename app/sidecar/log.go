package sidecar

import (
	"github.com/charmbracelet/log"

	"github.com/ssd39/smart-vault-sgx-app/app/utils"
)

var logger *log.Logger

func init() {
	logger = utils.Logger
	logger.SetPrefix("sidecar")
}
