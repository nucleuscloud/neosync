package openaigenerate

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/warpstreamlabs/bento/public/service"
)

const (
	openaiApiUrl = "https://api.openai.com/v1"
)

func getSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("api_url").Default(openaiApiUrl)).
		Field(service.NewStringField("api_key")).
		Field(service.NewStringField("user_prompt").Optional()).
		Field(service.NewStringListField("columns")).
		Field(service.NewStringListField("data_types")).
		Field(service.NewStringField("model")).
		Field(service.NewIntField("count")).
		Field(service.NewIntField("batch_size"))
}

func RegisterOpenaiGenerate(env *service.Environment) error {
	return env.RegisterBatchInput("openai_generate", getSpec(), func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
		rdr, err := newGenerateReader(conf, mgr)
		if err != nil {
			return nil, err
		}
		return service.AutoRetryNacksBatched(rdr), nil
	})
}

type generateReader struct {
	apiUrl string
	apikey string
	// The number of records to generate. This is progressively decremented as the reads progress
	count int
	// The number of records to generate per batch
	batchsize int
	model     string

	columns   []string
	dataTypes []string

	client *azopenai.Client

	promptMut sync.Mutex

	conversation []azopenai.ChatRequestMessageClassification

	log *service.Logger
}

func newGenerateReader(conf *service.ParsedConfig, mgr *service.Resources) (*generateReader, error) {
	apiUrl, err := conf.FieldString("api_url")
	if err != nil {
		return nil, err
	}
	apikey, err := conf.FieldString("api_key")
	if err != nil {
		return nil, err
	}
	var userPrompt *string
	if conf.Contains("user_prompt") {
		p, err := conf.FieldString("user_prompt")
		if err != nil {
			return nil, err
		}
		userPrompt = &p
	}

	columns, err := conf.FieldStringList("columns")
	if err != nil {
		return nil, err
	}
	dataTypes, err := conf.FieldStringList("data_types")
	if err != nil {
		return nil, err
	}
	if len(columns) != len(dataTypes) {
		return nil, fmt.Errorf("length of columns and data types was not the same: %d v %d", len(columns), len(dataTypes))
	}

	count, err := conf.FieldInt("count")
	if err != nil {
		return nil, err
	}
	batchsize, err := conf.FieldInt("batch_size")
	if err != nil {
		return nil, err
	}
	model, err := conf.FieldString("model")
	if err != nil {
		return nil, err
	}
	systemPrompt := buildSystemPrompt(columns, dataTypes)
	conversation := []azopenai.ChatRequestMessageClassification{
		&azopenai.ChatRequestSystemMessage{
			Content: azopenai.NewChatRequestSystemMessageContent(systemPrompt),
		},
	}
	prompt := ""
	if userPrompt != nil {
		prompt = fmt.Sprintf("%s\n", *userPrompt)
	}
	prompt += fmt.Sprintf("Generate %d records", batchsize)
	conversation = append(conversation, &azopenai.ChatRequestUserMessage{
		Content: azopenai.NewChatRequestUserMessageContent(prompt),
	})
	return &generateReader{
		apiUrl:    apiUrl,
		apikey:    apikey,
		count:     count,
		batchsize: batchsize,
		model:     model,

		conversation: conversation,

		log: mgr.Logger(),

		columns:   columns,
		dataTypes: dataTypes,
	}, nil
}

func buildSystemPrompt(
	columns, dataTypes []string,
) string {
	csvPrompt :=
		"You generate valid CSV data. When generating records, include the headers, one record per line. Only return the CSV data. Ensure each record has exact number of fields as headers. Separate each record by a newline. Do NOT return anything other than the raw CSV data. Do NOT include the csv markdown wrapper."

	headerPrompt := getColumnPrompt(columns, dataTypes)
	return fmt.Sprintf("%s %s %s",
		csvPrompt,
		headerPrompt,
		"The data returned should be in the exact order the headers are defined.",
	)
}

func getColumnPrompt(columns, dataTypes []string) string {
	pieces := []string{"Headers and their data types as follows:"}
	for idx := range columns {
		pieces = append(pieces, fmt.Sprintf("%s is %s", columns[idx], dataTypes[idx]))
	}
	return strings.Join(pieces, " ")
}

var _ service.BatchInput = &generateReader{}

func (b *generateReader) Connect(ctx context.Context) error {
	b.promptMut.Lock()
	defer b.promptMut.Unlock()
	if b.client != nil {
		return nil
	}
	client, err := azopenai.NewClientForOpenAI(b.apiUrl, azcore.NewKeyCredential(b.apikey), &azopenai.ClientOptions{})
	if err != nil {
		return err
	}
	b.client = client
	return nil
}

func (b *generateReader) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	b.promptMut.Lock()
	defer b.promptMut.Unlock()
	if b.client == nil {
		return nil, nil, service.ErrNotConnected
	}
	batchSize := b.batchsize
	if b.count <= 0 {
		return nil, nil, service.ErrEndOfInput
	}
	if b.count < batchSize {
		batchSize = b.count
	}
	b.log.Infof("Starting batch read with batch size %d", batchSize)

	resp, err := b.client.GetChatCompletions(ctx, azopenai.ChatCompletionsOptions{
		Temperature:      ptr(float32(1.0)),
		DeploymentName:   &b.model,
		TopP:             ptr(float32(1.0)),
		FrequencyPenalty: ptr(float32(0)),
		N:                ptr(int32(1)),
		ResponseFormat:   &azopenai.ChatCompletionsTextResponseFormat{},
		Messages:         b.conversation,
	}, &azopenai.GetChatCompletionsOptions{})
	if err != nil {
		return nil, nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, nil, fmt.Errorf("received no choices back from openai")
	}
	choice := resp.Choices[0]
	switch *choice.FinishReason {
	case azopenai.CompletionsFinishReasonTokenLimitReached:
		b.log.Warn("openai_generate: hit token limit reached, trimmed conversation")
		b.conversation = trimNonSystemMessages(b.conversation, 1)
	case azopenai.CompletionsFinishReasonContentFiltered:
		return nil, nil, errors.New("openai: generation stopped due to openai content being filtered due to moderation policies")
	default:
	}
	b.conversation = append(
		b.conversation,
		&azopenai.ChatRequestAssistantMessage{Content: azopenai.NewChatRequestAssistantMessageContent(*choice.Message.Content)},
		&azopenai.ChatRequestUserMessage{Content: azopenai.NewChatRequestUserMessageContent(fmt.Sprintf("%d more records", batchSize))},
	)

	records, err := getCsvRecordsFromInput(*choice.Message.Content, b.log)
	if err != nil {
		return nil, nil, fmt.Errorf("openai_generate: unable to fully process records retrieved from openai: %w", err)
	}
	if len(records) == 0 {
		b.log.Warn("openai_generate: no records were returned from message")
		return nil, nil, service.ErrEndOfInput
	}
	if len(records) == 1 {
		b.log.Warn("openai_generate: only headers were returned from message")
		return nil, nil, service.ErrEndOfInput
	}

	messageBatch := service.MessageBatch{}
	// skipping the first record as it returns the headers
	for _, record := range records[1:] {
		if b.count == 0 {
			b.log.Infof("stopping openai_generate as we've reached a count of 0 even though we had more records to process")
			break
		}
		structuredRecord, err := convertCsvToStructuredRecord(record, b.columns, b.dataTypes)
		if err != nil {
			b.log.Errorf("unable to convert csv record to structured record: %s", err.Error())
			continue
		}
		msg := service.NewMessage(nil)
		msg.SetStructured(structuredRecord)
		messageBatch = append(messageBatch, msg)
		b.count -= 1
	}
	if len(messageBatch) == 0 {
		return nil, nil, errors.New("openai_generate: received response from openai but was unable to successfully process records to a structured format. see logs for more details.")
	}
	return messageBatch, emptyAck, nil
}

func (b *generateReader) Close(ctx context.Context) error {
	b.promptMut.Lock()
	defer b.promptMut.Unlock()
	b.client = nil
	b.conversation = nil
	return nil
}

func emptyAck(ctx context.Context, err error) error {
	// Nacks are handled by AutoRetryNacks because we don't have an explicit
	// ack mechanism right now.
	return nil
}

func ptr[T any](val T) *T {
	return &val
}

func convertCsvToStructuredRecord(record, headers, types []string) (map[string]any, error) {
	if len(record) != len(headers) && len(headers) != len(types) && len(record) != len(types) {
		return nil, fmt.Errorf("error converting csv record to structured record, record headers and types not equivalent in length")
	}
	output := map[string]any{}
	for idx, value := range record {
		header := headers[idx]
		valueType := types[idx]
		convertedValue, err := toStructuredRecordValueType(value, valueType)
		if err != nil {
			return nil, fmt.Errorf("unable to convert value to correct type from csv: %w", err)
		}
		output[header] = convertedValue
	}

	return output, nil
}

func toStructuredRecordValueType(value, dataType string) (any, error) {
	switch dataType {
	case "smallint", "integer", "int", "serial":
		return strconv.ParseInt(strings.TrimSpace(value), 10, 32)
	case "bigint", "bigserial":
		return strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	case "numeric", "decimal":
		return strconv.ParseFloat(strings.TrimSpace(value), 64)
	case "real":
		return strconv.ParseFloat(strings.TrimSpace(value), 32)
	case "double precision":
		return strconv.ParseFloat(strings.TrimSpace(value), 64)
	case "money":
		return strings.TrimSpace(value), nil // Keeping it as string due to currency formatting
	case "character varying", "varchar", "character", "char", "text":
		return value, nil
	case "date", "timestamp", "timestamp without time zone":
		//nolint:gocritic
		// return time.Parse("2006-01-02 15:04:05", value) // adjust format as needed
		return strings.TrimSpace(value), nil
	case "timestamp with time zone":
		//nolint:gocritic
		// return time.Parse(time.RFC3339, value)
		return strings.TrimSpace(value), nil
	case "time", "time without time zone":
		//nolint:gocritic
		// return time.Parse("15:04:05", value)
		return strings.TrimSpace(value), nil
	case "time with time zone":
		//nolint:gocritic
		// return time.Parse("15:04:05Z07:00", value)
		return strings.TrimSpace(value), nil
	case "interval":
		return strings.TrimSpace(value), nil // Parsing intervals can be complex; keeping it as string
	case "boolean":
		return strconv.ParseBool(strings.TrimSpace(value))
	case "uuid":
		return strings.TrimSpace(value), nil // UUIDs are usually handled as strings
	case "json", "jsonb":
		return strings.TrimSpace(value), nil // JSON is typically handled as a string or a map
	default:
		if strings.HasSuffix(dataType, "[]") {
			return parseBracketedArray(strings.TrimSpace(value)), nil
		}
		return strings.TrimSpace(value), nil
	}
}

func parseBracketedArray(value string) []any {
	value = strings.Trim(value, "[]")
	if value == "" {
		return []any{}
	}
	elements := strings.Split(value, ",")
	var array []any
	for _, element := range elements {
		array = append(array, strings.TrimSpace(element))
	}
	return array
}

func getCsvReader(input string) *csv.Reader {
	var buffer bytes.Buffer
	buffer.WriteString(input)
	return csv.NewReader(&buffer)
}

func getCsvRecordsFromInput(input string, logger *service.Logger) ([][]string, error) {
	reader := getCsvReader(input)
	reader.FieldsPerRecord = -1
	var headers []string
	for {
		row, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, errors.New("openai_generate: unable to process generated csv record response from openai")
			}
			return nil, fmt.Errorf("unable to process CSV row to retrieve headers: %w", err)
		}
		// handles the case where sometimes the llm returns the csv response wrapped in markdown csv block
		if len(row) == 1 && row[0] == "```csv" {
			logger.Debug("encountered opening markdown csv block in ai output, trimming...")
			continue
		}
		headers = row
		break
	}

	count := len(headers)
	records := [][]string{headers}
	totalProcessed := 0
	for {
		totalProcessed += 1
		row, err := reader.Read()
		if err != nil {
			if errors.Is(err, csv.ErrBareQuote) || errors.Is(err, csv.ErrQuote) {
				logger.Warnf("encountered malformed csv row, skipping. %s", err.Error())
				continue
			} else if errors.Is(err, io.EOF) {
				// including EOF as line count to inadvertingly account for header line
				logger.Debug(fmt.Sprintf("pulled %d/%d records/lines out of csv output", len(records), totalProcessed))
				return records, nil
			} else {
				logger.Warnf("encountered error when reading csv from openai: %s", err.Error())
				continue
			}
		}

		if len(row) == 1 && row[0] == "```" {
			// LLM returned the CSV block and we should not parse this.
			// next read should result in EOF
			logger.Debug("encountered closing markdown csv block in ai output, trimming...")
			continue
		}

		if len(row) == count {
			records = append(records, row)
		} else {
			logger.Warnf("encountered a csv row that did not have the correct number of columns. had: '%d' want: '%d'", len(row), count)
		}
	}
}

func trimNonSystemMessages(messages []azopenai.ChatRequestMessageClassification, count int) []azopenai.ChatRequestMessageClassification {
	if len(messages) <= count {
		return messages[:0] // Return an empty slice
	}

	nonSystemIdx := findFirstNonSystemIdx(messages)
	if nonSystemIdx == -1 {
		// No non-system messages found, return the original messages
		return messages
	}

	// Calculate the end index for trimming
	endIdx := nonSystemIdx + count
	if endIdx >= len(messages) {
		endIdx = len(messages)
	}

	// Return the trimmed slice
	return append(messages[:nonSystemIdx], messages[endIdx:]...)
}

func findFirstNonSystemIdx(messages []azopenai.ChatRequestMessageClassification) int {
	if len(messages) == 0 {
		return -1
	}
	for idx := range messages {
		msg := messages[idx]
		switch msg.(type) {
		case *azopenai.ChatRequestSystemMessage:
			continue
		default:
			return idx
		}
	}
	return -1
}
