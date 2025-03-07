package piidetect_table_activities

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/connectiondata"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	"github.com/openai/openai-go"
	"go.temporal.io/sdk/activity"
)

type Activities struct {
	connclient            mgmtv1alpha1connect.ConnectionServiceClient
	openaiclient          *openai.Client
	connectiondatabuilder connectiondata.ConnectionDataBuilder
}

func New(
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	openaiclient *openai.Client,
	connectiondatabuilder connectiondata.ConnectionDataBuilder,
) *Activities {
	return &Activities{
		connclient:            connclient,
		openaiclient:          openaiclient,
		connectiondatabuilder: connectiondatabuilder,
	}
}

type GetColumnDataRequest struct {
	ConnectionId string
	TableSchema  string
	TableName    string
}

type GetColumnDataResponse struct {
	ColumnData []*ColumnData
}

type ColumnData struct {
	Column     string
	DataType   string
	IsNullable bool
	Comment    *string
}

func (a *Activities) GetColumnData(ctx context.Context, req *GetColumnDataRequest) (*GetColumnDataResponse, error) {
	logger := activity.GetLogger(ctx)
	slogger := temporallogger.NewSlogger(logger)

	databaseColumns, err := a.getTableDetailsFromConnection(
		ctx,
		req.ConnectionId,
		req.TableSchema,
		req.TableName,
		slogger,
	)
	if err != nil {
		return nil, err
	}

	columnData := make([]*ColumnData, len(databaseColumns))
	for i, column := range databaseColumns {
		columnData[i] = &ColumnData{
			Column:     column.GetColumn(),
			DataType:   column.GetDataType(),
			IsNullable: column.GetIsNullable() == "YES",
			Comment:    nil, // todo
		}
	}

	return &GetColumnDataResponse{
		ColumnData: columnData,
	}, nil
}

func (a *Activities) getTableDetailsFromConnection(
	ctx context.Context,
	connectionId string,
	tableSchema string,
	tableName string,
	logger *slog.Logger,
) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	connResp, err := a.connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: connectionId,
	}))
	if err != nil {
		return nil, err
	}
	connection := connResp.Msg.GetConnection()

	connectionData, err := a.connectiondatabuilder.NewDataConnection(logger, connection)
	if err != nil {
		return nil, err
	}

	tableDetails, err := connectionData.GetTableSchema(ctx, tableSchema, tableName)
	if err != nil {
		return nil, err
	}

	return getSchemasByTable(tableDetails, tableSchema, tableName), nil
}

type PiiCategory string

const (
	PiiCategoryNationalId PiiCategory = "national_id"
	PiiCategoryContact    PiiCategory = "contact"
	PiiCategoryFinancial  PiiCategory = "financial"
	PiiCategoryPersonal   PiiCategory = "personal"
	PiiCategoryLocation   PiiCategory = "location"
	PiiCategoryAuth       PiiCategory = "authentication"
)

// GetAllPiiCategories returns a slice of all defined PII categories
func GetAllPiiCategories() []PiiCategory {
	return []PiiCategory{
		PiiCategoryNationalId,
		PiiCategoryContact,
		PiiCategoryFinancial,
		PiiCategoryPersonal,
		PiiCategoryLocation,
		PiiCategoryAuth,
	}
}

func GetAllPiiCategoriesAsStrings() []string {
	categories := GetAllPiiCategories()
	result := make([]string, len(categories))
	for i, category := range categories {
		result[i] = string(category)
	}
	return result
}

type PiiPattern struct {
	Pattern  *regexp.Regexp
	Category PiiCategory
}

var piiColumnPatterns []PiiPattern

func init() {
	// Define patterns and compile them
	patterns := []struct {
		expr     string
		category PiiCategory
	}{
		{`(?i)^ssn$|(?i)^social_security_number$`, PiiCategoryNationalId},
		{`(?i)^social[_-]?security`, PiiCategoryNationalId},
		{`(?i)^tax[_-]?id$`, PiiCategoryNationalId},
		{`(?i)^passport$|(?i)^passport[_-]?number$`, PiiCategoryNationalId},
		{`(?i)^driver[_-]?licen[sc]e$`, PiiCategoryNationalId},

		{`(?i)^email$|(?i)^email[_-]?address$`, PiiCategoryContact},
		{`(?i)^(phone|telephone|mobile)[_-]?number$`, PiiCategoryContact},

		{`(?i)^birth[_-]?date$`, PiiCategoryPersonal},
		{`(?i)^(first|last|full)[_-]?name$`, PiiCategoryPersonal},
		{`(?i)^dob$|(?i)^date[_-]?of[_-]?birth$`, PiiCategoryPersonal},

		{`(?i)^(credit[_-]?card|cc)[_-]?(num(ber)?)?$`, PiiCategoryFinancial},
		{`(?i)^account[_-]?num(ber)?$`, PiiCategoryFinancial},
		{`(?i)^bank[_-]?account[_-]?(num(ber)?)?$`, PiiCategoryFinancial},

		{`(?i)^password$|(?i)^passwd$`, PiiCategoryAuth},
		{`(?i)^auth[_-]?token$|(?i)^secret$`, PiiCategoryAuth},

		{`(?i)^address$|(?i)^mailing[_-]?address$`, PiiCategoryLocation},
		{`(?i)^(zip|postal)[_-]?code$`, PiiCategoryLocation},
		{`(?i)^ip[_-]?address$`, PiiCategoryLocation},
	}

	// Compile patterns
	for _, p := range patterns {
		compiled, err := regexp.Compile(p.expr)
		if err != nil {
			// In production, you might want to handle this differently
			panic(fmt.Sprintf("invalid regex pattern %q: %v", p.expr, err))
		}
		piiColumnPatterns = append(piiColumnPatterns, PiiPattern{
			Pattern:  compiled,
			Category: p.category,
		})
	}
}

type DetectPiiRegexRequest struct {
	ColumnData []*ColumnData
}

type DetectPiiRegexResponse struct {
	PiiColumns map[string]PiiCategory // Changed to map column names to their PII category
}

func (a *Activities) DetectPiiRegex(ctx context.Context, req *DetectPiiRegexRequest) (*DetectPiiRegexResponse, error) {
	logger := activity.GetLogger(ctx)

	piiColumns := make(map[string]PiiCategory)

	for _, dbCol := range req.ColumnData {
		column := dbCol.Column
		if category, isPii := isPiiColumn(column); isPii {
			piiColumns[column] = category
		}
	}

	logger.Debug("Regex PII detection complete")

	return &DetectPiiRegexResponse{
		PiiColumns: piiColumns,
	}, nil
}

type DetectPiiLLMRequest struct {
	TableName    string
	ColumnData   []*ColumnData
	ShouldSample bool
}

type DetectPiiLLMResponse struct {
	PiiColumns map[string]PiiDetectReport
}

type PiiDetectReport struct {
	Category   PiiCategory
	Confidence float64
}

func (a *Activities) DetectPiiLLM(ctx context.Context, req *DetectPiiLLMRequest) (*DetectPiiLLMResponse, error) {
	logger := activity.GetLogger(ctx)

	columnNames := make([]string, len(req.ColumnData))
	for i, col := range req.ColumnData {
		columnNames[i] = col.Column
	}

	if req.ShouldSample {
		// todo: sample data from the table
		logger.Warn("Sampling data from table is not yet implemented")
	}

	chatResp, err := a.openaiclient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Temperature:    openai.F(0.0),
		Model:          openai.F(openai.ChatModelGPT4o),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](openai.ResponseFormatJSONObjectParam{Type: openai.F(openai.ResponseFormatJSONObjectTypeJSONObject)}),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant that classifies database fields for PII."),
			openai.UserMessage(fmt.Sprintf(
				`You are a data classification expert tasked with identifying if database fields contain Personally Identifiable Information (PII).
Classify each field based on its name into one of these categories: %s

Provide your response as a JSON object that has the key called "output", where the value is an array of objects, each with the following keys:
- "field_name": The name of the field.
- "category": The most likely category.
- "confidence": A number between 0 and 1 indicating your confidence in this classification (1 being the most confident).

Here is the table name: %s

Here are the fields: %s`, strings.Join(GetAllPiiCategoriesAsStrings(), ", "), req.TableName, strings.Join(columnNames, ", "))),
		}),
	})
	if err != nil {
		return nil, err
	}

	var openAiResp openAiResponse
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &openAiResp); err != nil {
		return nil, err
	}

	piiColumns := make(map[string]PiiDetectReport)
	for _, column := range openAiResp.Output {
		piiColumns[column.FieldName] = PiiDetectReport{
			Category:   column.Category,
			Confidence: column.Confidence,
		}
	}

	logger.Debug("LLM PII detection complete")

	return &DetectPiiLLMResponse{
		PiiColumns: piiColumns,
	}, nil
}

type openAiResponse struct {
	Output []openAiColumnResponse `json:"output"`
}
type openAiColumnResponse struct {
	FieldName  string      `json:"field_name"`
	Category   PiiCategory `json:"category"`
	Confidence float64     `json:"confidence"`
}

// isPiiColumn checks if a column name matches any PII patterns and returns the category
func isPiiColumn(columnName string) (PiiCategory, bool) {
	for _, pattern := range piiColumnPatterns {
		if pattern.Pattern.MatchString(columnName) {
			return pattern.Category, true
		}
	}
	return "", false
}

func getSchemasByTable(databaseColumns []*mgmtv1alpha1.DatabaseColumn, schema, table string) []*mgmtv1alpha1.DatabaseColumn {
	output := []*mgmtv1alpha1.DatabaseColumn{}
	for _, databaseColumn := range databaseColumns {
		if databaseColumn.Schema == schema && databaseColumn.Table == table {
			output = append(output, databaseColumn)
		}
	}
	return output
}
