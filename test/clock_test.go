package test

import (
	"testing"

	"github.com/ssd39/smart-vault-sgx-app/app/chainhelper"
)

func TestClock(t *testing.T) {
	chainhelper.GetCurTime()
}
