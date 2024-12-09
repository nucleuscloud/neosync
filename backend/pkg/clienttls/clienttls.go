package clienttls

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"golang.org/x/sync/errgroup"
)

type ClientTlsFileConfig struct {
	RootCert *string

	ClientCert *string
	ClientKey  *string
}

type ClientTlsFileHandler func(config *mgmtv1alpha1.ClientTlsConfig) (*ClientTlsFileConfig, error)

// Joins the client cert and key into a single file
func UpsertClientTlsFileSingleClient(config *mgmtv1alpha1.ClientTlsConfig) (*ClientTlsFileConfig, error) {
	if config == nil {
		return nil, errors.New("config was nil")
	}

	errgrp := errgroup.Group{}

	filenames := GetClientTlsFileNamesSingleClient(config)

	errgrp.Go(func() error {
		if filenames.RootCert == nil {
			return nil
		}
		_, err := os.Stat(*filenames.RootCert)
		if err != nil && !os.IsNotExist(err) {
			return err
		} else if err != nil && os.IsNotExist(err) {
			if err := os.WriteFile(*filenames.RootCert, []byte(config.GetRootCert()), 0600); err != nil {
				return err
			}
		}
		return nil
	})
	errgrp.Go(func() error {
		if filenames.ClientCert != nil && filenames.ClientKey != nil {
			_, err := os.Stat(*filenames.ClientKey)
			if err != nil && !os.IsNotExist(err) {
				return err
			} else if err != nil && os.IsNotExist(err) {
				if err := os.WriteFile(*filenames.ClientKey, []byte(fmt.Sprintf("%s\n%s", config.GetClientKey(), config.GetClientCert())), 0600); err != nil {
					return err
				}
			}
		}
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return &filenames, nil
}

// Joins the client cert and key into a single file
func GetClientTlsFileNamesSingleClient(config *mgmtv1alpha1.ClientTlsConfig) ClientTlsFileConfig {
	if config == nil {
		return ClientTlsFileConfig{}
	}

	basedir := os.TempDir()

	output := ClientTlsFileConfig{}
	if config.GetRootCert() != "" {
		content := hashContent(config.GetRootCert())
		fullpath := filepath.Join(basedir, content)
		output.RootCert = &fullpath
	}
	if config.GetClientCert() != "" && config.GetClientKey() != "" {
		certContent := hashContent(config.GetClientCert())
		keyContent := hashContent(config.GetClientKey())

		joinedContent := hashContent(fmt.Sprintf("%s%s", certContent, keyContent))
		joinedPath := filepath.Join(basedir, joinedContent)
		output.ClientCert = &joinedPath
		output.ClientKey = &joinedPath
	}
	return output
}

func hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
