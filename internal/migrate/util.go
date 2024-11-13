package neomigrate

import (
	"fmt"
	"path/filepath"
	"strings"
)

func getMigrationSourceUrl(migrationDirectory string) (string, error) {
	var absSchemaDir string
	if filepath.IsAbs(migrationDirectory) {
		absSchemaDir = migrationDirectory
	} else {
		a, err := filepath.Abs(migrationDirectory)
		if err != nil {
			return "", err
		}
		absSchemaDir = a
	}

	return fmt.Sprintf("file://%s", strings.TrimPrefix(absSchemaDir, "file://")), nil
}
