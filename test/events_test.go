package test

import (
	"testing"

	"github.com/ssd39/smart-vault-sgx-app/app/chainhelper"
)

func TestEvents(t *testing.T) {
	eventChan := make(chan chainhelper.Instruction)
	err := chainhelper.ListenEvents(eventChan)
	if err != nil {
		t.Error(err)
	}
	for event := range eventChan {
		t.Log(event)
	}
}
