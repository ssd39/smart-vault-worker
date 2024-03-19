package entrypoint

import (
	"os"

	"github.com/edgelesssys/ego/ecrypto"
)

func SealKeyToFile(seed []byte, filePath string) error {
	out, err := ecrypto.SealWithUniqueKey(seed, nil)
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, out, 0600)
	if err != nil {
		return err
	}
	return nil
}

func UnSealKey(filePath string) ([]byte, error) {
	encSeed, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	seed, err := ecrypto.Unseal(encSeed, nil)
	if err != nil {
		return nil, err
	}
	return seed, nil
}
