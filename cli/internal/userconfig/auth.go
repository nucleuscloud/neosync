package userconfig

import (
	"os"
	"path/filepath"
)

const (
	accessTokenFileName  = "access_token"
	refreshTokenfileName = "refresh_token"
)

func GetAccessToken() (string, error) {
	dirpath, err := GetOrCreateNeosyncFolder()
	if err != nil {
		return "", err
	}

	bits, err := os.ReadFile(getFullAccessTokenFileName(dirpath))
	if err != nil {
		return "", err
	}
	return string(bits), nil
}

func SetAccessToken(token string) error {
	dirpath, err := GetOrCreateNeosyncFolder()
	if err != nil {
		return err
	}
	return os.WriteFile(getFullAccessTokenFileName(dirpath), []byte(token), 0600)
}

func RemoveAccessToken() error {
	dirpath, err := GetOrCreateNeosyncFolder()
	if err != nil {
		return err
	}
	return os.Remove(getFullAccessTokenFileName(dirpath))
}

func GetRefreshToken() (string, error) {
	dirpath, err := GetOrCreateNeosyncFolder()
	if err != nil {
		return "", err
	}

	bits, err := os.ReadFile(getFullRefreshTokenFileName(dirpath))
	if err != nil {
		return "", err
	}
	return string(bits), nil
}

func SetRefreshToken(token string) error {
	dirpath, err := GetOrCreateNeosyncFolder()
	if err != nil {
		return err
	}
	return os.WriteFile(getFullRefreshTokenFileName(dirpath), []byte(token), 0600)
}

func RemoveRefreshToken() error {
	dirpath, err := GetOrCreateNeosyncFolder()
	if err != nil {
		return err
	}
	return os.Remove(getFullRefreshTokenFileName(dirpath))
}

func getFullAccessTokenFileName(dirpath string) string {
	return filepath.Join(dirpath, accessTokenFileName)
}

func getFullRefreshTokenFileName(dirpath string) string {
	return filepath.Join(dirpath, refreshTokenfileName)
}
