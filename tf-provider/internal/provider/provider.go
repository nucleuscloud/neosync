package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	http_client "github.com/nucleuscloud/neosync/terraform-provider/internal/http/client"
)

// Ensure NeosyncProvider satisfies various provider inferfaces
var _ provider.Provider = &NeosyncProvider{}

type NeosyncProvider struct {
	version string
}

type NeosyncProviderModel struct {
	ApiToken types.String `tfsdk:"api_token"`
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *NeosyncProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "neosync"
	resp.Version = p.version
}

func (p *NeosyncProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The URL to the backend Neosync API server",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "The account-level API token that will be used to authenticate with the API server",
				Optional:            true,
			},
		},
	}
}

type ConfigData struct {
	ConnectionClient mgmtv1alpha1connect.ConnectionServiceClient
}

func (p *NeosyncProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	apiToken := os.Getenv("NEOSYNC_API_TOKEN")
	endpoint := os.Getenv("NEOSYNC_ENDPOINT")

	var data NeosyncProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Check configuration data, which should take precedence over
	// environment variable data, if found.
	if data.ApiToken.ValueString() != "" {
		apiToken = data.ApiToken.ValueString()
	}

	if data.Endpoint.ValueString() != "" {
		endpoint = data.Endpoint.ValueString()
	}

	if apiToken == "" {
		resp.Diagnostics.AddWarning(
			"Missing API Token Configuration",
			"While configuring the provider, the API token was not found in "+
				"the NEOSYNC_API_TOKEN environment variable or provider "+
				"configuration block api_token attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Endpoint Configuration",
			"While configuring the provider, the endpoint was not found in "+
				"the NEOSYNC_ENDPOINT environment variable or provider "+
				"configuration block endpoint attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	httpclient := http.DefaultClient
	if apiToken != "" {
		httpclient = http_client.NewWithHeaders(
			map[string]string{"Authorization": fmt.Sprintf("Bearer %s", apiToken)},
		)
	}

	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		httpclient,
		endpoint,
	)

	configData := &ConfigData{
		ConnectionClient: connclient,
	}

	resp.DataSourceData = configData
	resp.ResourceData = configData
}

func (p *NeosyncProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// NewExampleResource,
	}
}

func (p *NeosyncProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NeosyncProvider{
			version: version,
		}
	}
}
