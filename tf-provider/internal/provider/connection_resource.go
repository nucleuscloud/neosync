package provider

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
)

var _ resource.Resource = &ConnectionResource{}
var _ resource.ResourceWithImportState = &ConnectionResource{}

func NewConnectionResource() resource.Resource {
	return &ConnectionResource{}
}

type ConnectionResource struct {
	client    mgmtv1alpha1connect.ConnectionServiceClient
	accountId *string
}

type ConnectionResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	AccountId types.String `tfsdk:"accountId"`

	Postgres *Postgres `tfsdk:"postgres"`
}

type Postgres struct {
	Url types.String `tfsdk:"url"`

	Host    types.String `tfsdk:"host"`
	Port    types.Int64  `tfsdk:"post"`
	Name    types.String `tfsdk:"name"`
	User    types.String `tfsdk:"user"`
	Pass    types.String `tfsdk:"pass"`
	SslMode types.String `tfsdk:"sslMode"`
}

func (r *ConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}

func (r *ConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The unique friendly name of the connection",
				Required:            true,
			},
			"accountId": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the account. Can be pulled from the API Key if present, or must be specified if using a user access token",
				Optional:            true,
			},
			"postgres": schema.SingleNestedAttribute{
				Description: "The postgres database that will be associated with this connection",
				Optional:    true,
				// PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "Standard postgres url connection string. Must be uri compliant",
						Required:    false,
					},

					"host": schema.StringAttribute{
						Description: "The host name of the postgres server",
						Required:    false,
					},
					"port": schema.Int64Attribute{
						Description: "The post of the postgres server",
						Required:    false,
						Default:     int64default.StaticInt64(5432),
					},
					"name": schema.StringAttribute{
						Description: "The name of the database that will be connected to",
						Required:    false,
					},
					"user": schema.StringAttribute{
						Description: "The name of the user that will be authenticated with",
						Required:    false,
					},
					"pass": schema.StringAttribute{
						Description: "The password that will be authenticated with",
						Required:    false,
						Sensitive:   true,
					},
					"sslMode": schema.StringAttribute{
						Description: "The SSL mode for the postgres server",
						Required:    false,
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the connection",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(mgmtv1alpha1connect.ConnectionServiceClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected mgmtv1alpha1connect.ConnectionServiceClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func hydrateResourceModelFromConnectionConfig(cc *mgmtv1alpha1.ConnectionConfig, data *ConnectionResourceModel) error {
	switch config := cc.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		switch pgcc := config.PgConfig.ConnectionConfig.(type) {
		case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
			data.Postgres = &Postgres{
				Host:    types.StringValue(pgcc.Connection.Host),
				Port:    types.Int64Value(int64(pgcc.Connection.Port)),
				Name:    types.StringValue(pgcc.Connection.Name),
				User:    types.StringValue(pgcc.Connection.User),
				Pass:    types.StringValue(pgcc.Connection.Pass),
				SslMode: types.StringPointerValue(pgcc.Connection.SslMode),
			}
			return nil
		case *mgmtv1alpha1.PostgresConnectionConfig_Url:
			data.Postgres = &Postgres{
				Url: types.StringValue(pgcc.Url),
			}
			return nil
		default:
			return errors.New("unable to find a config to hydrate connection resource model")

		}
	default:
		return errors.New("unable to find a config to hydrate connection resource model")
	}
}

func getConnectionConfigFromResourceModel(data *ConnectionResourceModel) (*mgmtv1alpha1.ConnectionConfig, error) {
	if data.Postgres != nil {
		if !data.Postgres.Url.IsNull() {
			return &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: data.Postgres.Url.String(),
						},
						Tunnel: nil,
					},
				},
			}, nil
		} else {
			pg := data.Postgres
			if pg.Host.IsNull() || pg.Port.IsNull() || pg.Name.IsNull() || pg.User.IsNull() || pg.Pass.IsNull() {
				return nil, fmt.Errorf("invalid postgres config")
			}
			return &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
							Connection: &mgmtv1alpha1.PostgresConnection{
								Host:    pg.Host.String(),
								Port:    int32(pg.Port.ValueInt64()),
								Name:    pg.Name.String(),
								User:    pg.User.String(),
								Pass:    pg.Pass.String(),
								SslMode: pg.SslMode.ValueStringPointer(),
							},
						},
						Tunnel: nil,
					},
				},
			}, nil
		}
	}
	return nil, errors.New("invalid connection config")
}

// nolint
func (r *ConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConnectionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var accountId string
	if data.AccountId.IsUnknown() || data.AccountId.IsNull() {
		if r.accountId != nil {
			accountId = *r.accountId
		}
	} else {
		accountId = data.AccountId.String()
	}
	if accountId == "" {
		resp.Diagnostics.AddError("no account id", "must provide account id either on the resource or provide through environment configuration")
		return
	}

	cc, err := getConnectionConfigFromResourceModel(&data)
	if err != nil {
		resp.Diagnostics.AddError("connection config error", err.Error())
		return
	}
	connResp, err := r.client.CreateConnection(ctx, connect.NewRequest(&mgmtv1alpha1.CreateConnectionRequest{
		Name:             data.Name.String(),
		AccountId:        accountId, // compute account id
		ConnectionConfig: cc,
	}))
	if err != nil {
		resp.Diagnostics.AddError("create connection error", err.Error())
		return
	}

	connection := connResp.Msg.Connection

	data.Id = types.StringValue(connection.Id)
	data.Name = types.StringValue(connection.Name)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created connection resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// nolint
func (r *ConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConnectionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	connResp, err := r.client.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: data.Id.String(),
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to get connection", err.Error())
		return
	}

	connection := connResp.Msg.Connection

	data.Id = types.StringValue(connection.Id)
	data.Name = types.StringValue(connection.Name)
	data.AccountId = types.StringValue(connection.AccountId)
	err = hydrateResourceModelFromConnectionConfig(connection.ConnectionConfig, &data)
	if err != nil {
		resp.Diagnostics.AddError("connection config hydration error", err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// nolint
func (r *ConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConnectionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cc, err := getConnectionConfigFromResourceModel(&data)
	if err != nil {
		resp.Diagnostics.AddError("connection config error", err.Error())
		return
	}

	connResp, err := r.client.UpdateConnection(ctx, connect.NewRequest(&mgmtv1alpha1.UpdateConnectionRequest{
		Id:               data.Id.String(),
		Name:             data.Name.String(),
		ConnectionConfig: cc,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update connection", err.Error())
		return
	}

	connection := connResp.Msg.Connection

	data.Id = types.StringValue(connection.Id)
	data.Name = types.StringValue(connection.Name)
	data.AccountId = types.StringValue(connection.AccountId)
	err = hydrateResourceModelFromConnectionConfig(connection.ConnectionConfig, &data)
	if err != nil {
		resp.Diagnostics.AddError("connection config hydration error", err.Error())
		return
	}

	tflog.Trace(ctx, "updated connection")
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// nolint
func (r *ConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConnectionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteConnection(ctx, connect.NewRequest(&mgmtv1alpha1.DeleteConnectionRequest{
		Id: data.Id.String(),
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete connection", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted connection")
}

func (r *ConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	// todo
}
