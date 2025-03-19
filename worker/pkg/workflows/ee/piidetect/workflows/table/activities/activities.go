package piidetect_table_activities

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"regexp"
	"sync"
	"text/template"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/internal/connectiondata"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/tiktoken-go/tokenizer"
	"go.temporal.io/sdk/activity"
)

type OpenAiCompletionsClient interface {
	New(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (res *openai.ChatCompletion, err error)
}

type Activities struct {
	connclient            mgmtv1alpha1connect.ConnectionServiceClient
	openaiclient          OpenAiCompletionsClient
	connectiondatabuilder connectiondata.ConnectionDataBuilder
	jobclient             mgmtv1alpha1connect.JobServiceClient
}

func New(
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	openaiclient OpenAiCompletionsClient,
	connectiondatabuilder connectiondata.ConnectionDataBuilder,
	jobclient mgmtv1alpha1connect.JobServiceClient,
) *Activities {
	return &Activities{
		connclient:            connclient,
		openaiclient:          openaiclient,
		connectiondatabuilder: connectiondatabuilder,
		jobclient:             jobclient,
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

func (p PiiCategory) String() string {
	return string(p)
}

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
		{`(?i)^age$`, PiiCategoryPersonal},

		{`(?i)^(credit[_-]?card|cc)[_-]?(num(ber)?)?$`, PiiCategoryFinancial},
		{`(?i)^account[_-]?num(ber)?$`, PiiCategoryFinancial},
		{`(?i)^bank[_-]?account[_-]?(num(ber)?)?$`, PiiCategoryFinancial},
		{`(?i)^salary$|(?i)^current[_-]?salary$`, PiiCategoryFinancial},

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
	TableSchema  string
	TableName    string
	ColumnData   []*ColumnData
	ShouldSample bool
	ConnectionId string
	UserPrompt   string
}

type DetectPiiLLMResponse struct {
	PiiColumns map[string]LLMPiiDetectReport
}

type LLMPiiDetectReport struct {
	Category   PiiCategory `json:"category"`
	Confidence float32     `json:"confidence"`
}

type RegexPiiDetectReport struct {
	Category PiiCategory `json:"category"`
}

type CombinedPiiDetectReport struct {
	Regex *RegexPiiDetectReport `json:"regex"`
	LLM   *LLMPiiDetectReport   `json:"llm"`
}

type sampleDataStream struct {
	records Records
	mu      sync.Mutex
}

var _ connectiondata.SampleDataStream = (*sampleDataStream)(nil)

func (s *sampleDataStream) Send(resp *mgmtv1alpha1.GetConnectionDataStreamResponse) error {
	decoder := gob.NewDecoder(bytes.NewReader(resp.GetRowBytes()))
	var record map[string]any
	err := decoder.Decode(&record)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, record)
	return nil
}

func sampleDataCollector() *sampleDataStream {
	return &sampleDataStream{
		records: []map[string]any{},
	}
}

func (s *sampleDataStream) GetRecords() Records {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Make a deep copy of the records
	copiedRecords := make(Records, len(s.records))
	for i, record := range s.records {
		// Copy each map
		copiedRecord := make(map[string]any, len(record))
		maps.Copy(copiedRecord, record)
		copiedRecords[i] = copiedRecord
	}
	return copiedRecords
}

type Records []map[string]any

func (r *Records) toAssociativeArray(maxRecords uint) map[string][]any {
	if len(*r) == 0 {
		return nil
	}

	// Create a map to hold all values for each column
	columnValues := make(map[string][]any)

	// Collect all column names first
	for _, record := range *r {
		for colName := range record {
			if _, exists := columnValues[colName]; !exists {
				columnValues[colName] = []any{}
			}
		}
	}

	// Populate the values for each column
	for _, record := range *r {
		for colName, value := range record {
			// skip empty maps
			if m, ok := value.(map[string]any); ok && len(m) == 0 {
				continue
			}
			columnValues[colName] = append(columnValues[colName], value)
		}
	}

	// limit the number of records to the maxRecords
	for colName, values := range columnValues {
		if uint(len(values)) > maxRecords {
			columnValues[colName] = values[:maxRecords]
		}
	}

	return columnValues
}

func (r *Records) ToPromptString(maxRecords uint) (string, error) {
	if len(*r) == 0 {
		return "{}", nil
	}

	columnValues := r.toAssociativeArray(maxRecords)

	jsonBytes, err := json.Marshal(columnValues)
	if err != nil {
		return "", fmt.Errorf("unable to convert column values to JSON in prompt string: %w", err)
	}

	return string(jsonBytes), nil
}

const (
	maxDataSamples = uint(5)
)

func (a *Activities) getSampleData(ctx context.Context, req *DetectPiiLLMRequest, logger *slog.Logger) (Records, error) {
	connResp, err := a.connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.ConnectionId,
	}))
	if err != nil {
		return nil, err
	}
	connection := connResp.Msg.GetConnection()
	connectionData, err := a.connectiondatabuilder.NewDataConnection(logger, connection)
	if err != nil {
		return nil, err
	}
	collector := sampleDataCollector()
	err = connectionData.SampleData(ctx, collector, req.TableSchema, req.TableName, maxDataSamples)
	if err != nil {
		return nil, err
	}
	return collector.GetRecords(), nil
}

const piiDetectionPrompt = `You are a data classification expert tasked with identifying if database fields contain Personally Identifiable Information (PII).
Classify each field based on its name into one of these categories: {{.Categories}}

Provide your response as a JSON object that has the key called "output", where the value is an array of objects, each with the following keys:
- "field_name": The name of the field.
- "category": The most likely category.
- "confidence": A number between 0 and 1 indicating your confidence in this classification (1 being the most confident).

Here is the table name: {{.TableName}}

{{if .UserPrompt}}{{.UserPrompt}}{{end}}

Here are the fields and (optionally) values: {{.RecordData}}`

var piiDetectionPromptTmpl = template.Must(template.New("pii_detection_prompt").Parse(piiDetectionPrompt))

const (
	systemMessage = "You are a helpful assistant that classifies database fields for PII."
	model         = openai.ChatModelGPT4oMini
	maxTokenLimit = 10_000 // dependent on model used
)

func countPromptTokens(prompt string) (int, error) {
	codec, err := tokenizer.ForModel(tokenizer.Model(model))
	if err != nil {
		return -1, err
	}
	tokens, err := codec.Count(prompt)
	if err != nil {
		return -1, err
	}
	return tokens, nil
}

func getPrompt(records Records, tableName, userPrompt string, maxRecords uint) (string, error) {
	// Try with initial maxRecords and keep reducing until we find a working size
	for currentMaxRecords := maxRecords; ; currentMaxRecords-- {
		recordPromptStr, err := records.ToPromptString(currentMaxRecords)
		if err != nil {
			return "", err
		}

		var promptBuf bytes.Buffer
		err = piiDetectionPromptTmpl.Execute(&promptBuf, map[string]any{
			"Categories": GetAllPiiCategoriesAsStrings(),
			"TableName":  tableName,
			"RecordData": recordPromptStr,
			"UserPrompt": userPrompt,
		})
		if err != nil {
			return "", err
		}

		prompt := promptBuf.String()

		// Count tokens for both system message and prompt
		totalTokens := 0
		for _, text := range []string{systemMessage, prompt} {
			tokens, err := countPromptTokens(text)
			if err != nil {
				return "", fmt.Errorf("failed to count tokens: %w", err)
			}
			totalTokens += tokens
		}

		// If we're within token limit, return this prompt
		if totalTokens <= maxTokenLimit {
			return prompt, nil
		}

		// If we're at 0 records and still over limit, something else is wrong
		if currentMaxRecords == 0 {
			return "", fmt.Errorf("prompt exceeds token limit (%d) even with no sample data", maxTokenLimit)
		}
	}
}

func (a *Activities) DetectPiiLLM(ctx context.Context, req *DetectPiiLLMRequest) (*DetectPiiLLMResponse, error) {
	logger := activity.GetLogger(ctx)
	slogger := temporallogger.NewSlogger(logger)

	records, err := a.getDataRecordsForLLM(ctx, req, slogger)
	if err != nil {
		return nil, err
	}

	userMessage, err := getPrompt(records, req.TableName, req.UserPrompt, maxDataSamples)
	if err != nil {
		return nil, err
	}
	logger.Debug("LLM PII detection prompt", "prompt", userMessage)

	chatResp, err := a.openaiclient.New(ctx, openai.ChatCompletionNewParams{
		Temperature:    openai.F(0.0),
		Model:          openai.F(model),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](openai.ResponseFormatJSONObjectParam{Type: openai.F(openai.ResponseFormatJSONObjectTypeJSONObject)}),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemMessage),
			openai.UserMessage(userMessage),
		}),
	})
	if err != nil {
		return nil, err
	}

	var openAiResp openAiResponse
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &openAiResp); err != nil {
		return nil, err
	}

	piiColumns := make(map[string]LLMPiiDetectReport)
	for _, column := range openAiResp.Output {
		piiColumns[column.FieldName] = LLMPiiDetectReport{
			Category:   column.Category,
			Confidence: column.Confidence,
		}
	}

	logger.Debug("LLM PII detection complete")

	return &DetectPiiLLMResponse{
		PiiColumns: piiColumns,
	}, nil
}

// Returns the data record for use with the LLM prompt.
// If shouldSample is true and connectionId is not empty, it will sample data from the database.
// Otherwise, it will return an empty record for each column.
// In other words, at a minimum, it will return a record that contains just the column names, but no data.
func (a *Activities) getDataRecordsForLLM(
	ctx context.Context,
	req *DetectPiiLLMRequest,
	slogger *slog.Logger,
) (Records, error) {
	if req.ShouldSample && req.ConnectionId != "" {
		return a.getSampleData(ctx, req, slogger)
	}

	records := Records{}
	for _, col := range req.ColumnData {
		records = append(records, map[string]any{
			col.Column: map[string]any{},
		})
	}
	return records, nil
}

type SaveTablePiiDetectReportRequest struct {
	ParentRunId    *string
	AccountId      string
	TableSchema    string
	TableName      string
	Report         map[string]CombinedPiiDetectReport
	ScannedColumns []string
}

type SaveTablePiiDetectReportResponse struct {
	Key *mgmtv1alpha1.RunContextKey
}

type ColumnReport struct {
	ColumnName string                  `json:"column_name"`
	Report     CombinedPiiDetectReport `json:"report"`
}

type TableReport struct {
	TableSchema   string         `json:"table_schema"`
	TableName     string         `json:"table_name"`
	ColumnReports []ColumnReport `json:"column_reports"`
	// Denotes all of the scanned columns as ColumnReports only includes columns that were detected as PII
	ScannedColumns []string `json:"scanned_columns,omitempty"`
}

func (a *Activities) SaveTablePiiDetectReport(
	ctx context.Context,
	req *SaveTablePiiDetectReportRequest,
	logger *slog.Logger,
) (*SaveTablePiiDetectReportResponse, error) {
	info := activity.GetInfo(ctx)
	jobRunId := info.WorkflowExecution.ID
	if req.ParentRunId != nil {
		jobRunId = *req.ParentRunId
	}

	columnReports := make([]ColumnReport, 0, len(req.Report))
	for columnName, report := range req.Report {
		columnReports = append(columnReports, ColumnReport{
			ColumnName: columnName,
			Report:     report,
		})
	}

	finalReport := &TableReport{
		TableSchema:    req.TableSchema,
		TableName:      req.TableName,
		ColumnReports:  columnReports,
		ScannedColumns: req.ScannedColumns,
	}

	reportBytes, err := json.Marshal(finalReport)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal report: %w", err)
	}

	key := &mgmtv1alpha1.RunContextKey{
		AccountId:  req.AccountId,
		JobRunId:   jobRunId,
		ExternalId: BuildTableReportExternalId(req.TableSchema, req.TableName),
	}

	_, err = a.jobclient.SetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.SetRunContextRequest{
		Id:    key,
		Value: reportBytes,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to set run context: %w", err)
	}
	return &SaveTablePiiDetectReportResponse{
		Key: key,
	}, nil
}

func BuildTableReportExternalId(tableSchema, tableName string) string {
	return fmt.Sprintf("%s.%s%s", tableSchema, tableName, PiiTableReportSuffix)
}

const (
	PiiTableReportSuffix = "--table-pii-report"
)

type openAiResponse struct {
	Output []openAiColumnResponse `json:"output"`
}
type openAiColumnResponse struct {
	FieldName  string      `json:"field_name"`
	Category   PiiCategory `json:"category"`
	Confidence float32     `json:"confidence"`
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
