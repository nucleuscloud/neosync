//go:build tools
// +build tools

package main

import (
	_ "github.com/golang-migrate/migrate/v4/cmd/migrate"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
