package userconfig

import (
	"os"
	"path/filepath"
)

const (
	accountIdFileName = "account_id"
)

func GetAccountId() (string, error) {
	dirpath, err := GetOrCreateNeosyncFolder()
	if err != nil {
		return "", err
	}

	bits, err := os.ReadFile(getFullAccountIdFileName(dirpath))
	if err != nil {
		return "", err
	}
	return string(bits), nil
}

func SetAccountId(id string) error {
	dirpath, err := GetOrCreateNeosyncFolder()
	if err != nil {
		return err
	}
	return os.WriteFile(getFullAccountIdFileName(dirpath), []byte(id), 0600)
}

func getFullAccountIdFileName(dirpath string) string {
	return filepath.Join(dirpath, accountIdFileName)
}
