package test

import (
	"testing"

	"github.com/ssd39/smart-vault-sgx-app/app/chainhelper"
)

func TestEvents(t *testing.T) {
	err := chainhelper.ListenEvents()
	if err != nil {
		t.Error(err)
	}
}
