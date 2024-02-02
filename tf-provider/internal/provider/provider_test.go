package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"neosync": providerserver.NewProtocol6WithError(New("test")()),
}

// nolint
func testAccPreCheck(t *testing.T) {
	mustHaveEnv(t, "NEOSYNC_ENDPOINT")
	if os.Getenv("NEOSYNC_API_TOKEN") == "" {
		mustHaveEnv(t, "NEOSYNC_ACCOUNT_ID")
	} else {
		mustHaveEnv(t, "NEOSYNC_API_TOKEN")
	}
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

func mustHaveEnv(t *testing.T, name string) {
	if os.Getenv(name) == "" {
		t.Fatalf("%s environment variable must be set for acceptance tests", name)
	}
}
