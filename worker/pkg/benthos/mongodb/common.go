package neosync_benthos_mongodb

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	neosync_mongodb "github.com/nucleuscloud/neosync/backend/pkg/mongomanager"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
	neosync_benthos_metadata "github.com/nucleuscloud/neosync/worker/pkg/benthos/metadata"
	"github.com/warpstreamlabs/bento/public/bloblang"
	"github.com/warpstreamlabs/bento/public/service"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// JSONMarshalMode represents the way in which BSON should be marshaled to JSON.
type JSONMarshalMode string

const (
	// JSONMarshalModeCanonical Canonical BSON to JSON marshal mode.
	JSONMarshalModeCanonical JSONMarshalMode = "canonical"
	// JSONMarshalModeRelaxed Relaxed BSON to JSON marshal mode.
	JSONMarshalModeRelaxed JSONMarshalMode = "relaxed"
)

//------------------------------------------------------------------------------

const (
	// Common Client Fields
	commonFieldClientConnectionId = "connection_id"
	commonFieldClientDatabase     = "database"
)

func clientFields() []*service.ConfigField {
	return []*service.ConfigField{
		service.NewURLField(commonFieldClientConnectionId).
			Description("The URL of the target MongoDB server.").
			Example("mongodb://localhost:27017"),
		service.NewStringField(commonFieldClientDatabase).
			Description("The name of the target MongoDB database."),
	}
}

//------------------------------------------------------------------------------

// Operation represents the operation that will be performed by MongoDB.
type Operation string

const (
	// OperationInsertOne Insert One operation.
	OperationInsertOne Operation = "insert-one"
	// OperationDeleteOne Delete One operation.
	OperationDeleteOne Operation = "delete-one"
	// OperationDeleteMany Delete many operation.
	OperationDeleteMany Operation = "delete-many"
	// OperationReplaceOne Replace one operation.
	OperationReplaceOne Operation = "replace-one"
	// OperationUpdateOne Update one operation.
	OperationUpdateOne Operation = "update-one"
	// OperationFindOne Find one operation.
	OperationFindOne Operation = "find-one"
	// OperationInvalid Invalid operation.
	OperationInvalid Operation = "invalid"
)

func (op Operation) isDocumentAllowed() bool {
	switch op {
	case OperationInsertOne,
		OperationReplaceOne,
		OperationUpdateOne:
		return true
	default:
		return false
	}
}

func (op Operation) isFilterAllowed() bool {
	switch op {
	case OperationDeleteOne,
		OperationDeleteMany,
		OperationReplaceOne,
		OperationUpdateOne,
		OperationFindOne:
		return true
	default:
		return false
	}
}

func (op Operation) isHintAllowed() bool {
	switch op {
	case OperationDeleteOne,
		OperationDeleteMany,
		OperationReplaceOne,
		OperationUpdateOne,
		OperationFindOne:
		return true
	default:
		return false
	}
}

func (op Operation) isUpsertAllowed() bool {
	switch op {
	case OperationReplaceOne,
		OperationUpdateOne:
		return true
	default:
		return false
	}
}

// NewOperation converts a string operation to a strongly-typed Operation.
func NewOperation(op string) Operation {
	switch op {
	case "insert-one":
		return OperationInsertOne
	case "delete-one":
		return OperationDeleteOne
	case "delete-many":
		return OperationDeleteMany
	case "replace-one":
		return OperationReplaceOne
	case "update-one":
		return OperationUpdateOne
	case "find-one":
		return OperationFindOne
	default:
		return OperationInvalid
	}
}

const (
	// Common Operation Fields
	commonFieldOperation = "operation"
)

const (
	// Common Write Map Fields
	commonFieldDocumentMap = "document_map"
	commonFieldFilterMap   = "filter_map"
	commonFieldHintMap     = "hint_map"
	commonFieldUpsert      = "upsert"
)

func writeMapsFields() []*service.ConfigField {
	return []*service.ConfigField{
		service.NewBloblangField(commonFieldDocumentMap).
			Description("A bloblang map representing a document to store within MongoDB, expressed as [extended JSON in canonical form](https://www.mongodb.com/docs/manual/reference/mongodb-extended-json/). The document map is required for the operations " +
				"insert-one, replace-one and update-one.").
			Default(""),
		service.NewBloblangField(commonFieldFilterMap).
			Description("A bloblang map representing a filter for a MongoDB command, expressed as [extended JSON in canonical form](https://www.mongodb.com/docs/manual/reference/mongodb-extended-json/). The filter map is required for all operations except " +
				"insert-one. It is used to find the document(s) for the operation. For example in a delete-one case, the filter map should " +
				"have the fields required to locate the document to delete.").
			Default(""),
		service.NewBloblangField(commonFieldHintMap).
			Description("A bloblang map representing the hint for the MongoDB command, expressed as [extended JSON in canonical form](https://www.mongodb.com/docs/manual/reference/mongodb-extended-json/). This map is optional and is used with all operations " +
				"except insert-one. It is used to improve performance of finding the documents in the mongodb.").
			Default(""),
		service.NewBoolField(commonFieldUpsert).
			Description("The upsert setting is optional and only applies for update-one and replace-one operations. If the filter specified in filter_map matches, the document is updated or replaced accordingly, otherwise it is created.").
			Version("3.60.0").
			Default(false),
	}
}

const (
	// Common Write Concern Fields
	commonFieldWriteConcern         = "write_concern"
	commonFieldWriteConcernW        = "w"
	commonFieldWriteConcernJ        = "j"
	commonFieldWriteConcernWTimeout = "w_timeout"
)

func writeConcernDocs() *service.ConfigField {
	return service.NewObjectField(commonFieldWriteConcern,
		service.NewStringField(commonFieldWriteConcernW).
			Description("W requests acknowledgement that write operations propagate to the specified number of mongodb instances.").
			Default(""),
		service.NewBoolField(commonFieldWriteConcernJ).
			Description("J requests acknowledgement from MongoDB that write operations are written to the journal.").
			Default(false),
		service.NewStringField(commonFieldWriteConcernWTimeout).
			Description("The write concern timeout.").
			Default(""),
	).Description("The write concern settings for the mongo connection.")
}

func writeConcernCollectionOptionFromParsed(pConf *service.ParsedConfig) (opt *options.CollectionOptions, err error) {
	pConf = pConf.Namespace(commonFieldWriteConcern)

	var w string
	if w, err = pConf.FieldString(commonFieldWriteConcernW); err != nil {
		return nil, err
	}

	var j bool
	if j, err = pConf.FieldBool(commonFieldWriteConcernJ); err != nil {
		return nil, err
	}

	var wTimeout time.Duration
	if dStr, _ := pConf.FieldString(commonFieldWriteConcernWTimeout); dStr != "" {
		if wTimeout, err = pConf.FieldDuration(commonFieldWriteConcernWTimeout); err != nil {
			return nil, err
		}
	}

	writeConcern := &writeconcern.WriteConcern{
		Journal:  &j,
		WTimeout: wTimeout,
	}
	if wInt, err := strconv.Atoi(w); err != nil {
		writeConcern.W = w
	} else {
		writeConcern.W = wInt
	}

	return options.Collection().SetWriteConcern(writeConcern), nil
}

func outputOperationDocs(defaultOperation Operation) *service.ConfigField {
	return service.NewStringEnumField("operation",
		string(OperationInsertOne),
		string(OperationDeleteOne),
		string(OperationDeleteMany),
		string(OperationReplaceOne),
		string(OperationUpdateOne),
	).Description("The mongodb operation to perform.").
		Default(string(defaultOperation))
}

func operationFromParsed(pConf *service.ParsedConfig) (operation Operation, err error) {
	var operationStr string
	if operationStr, err = pConf.FieldString(commonFieldOperation); err != nil {
		return
	}

	if operation = NewOperation(operationStr); operation == OperationInvalid {
		err = fmt.Errorf("mongodb operation '%s' unknown: must be insert-one, delete-one, delete-many, replace-one or update-one", operationStr)
	}
	return
}

type writeMaps struct {
	filterMap   *bloblang.Executor
	documentMap *bloblang.Executor
	hintMap     *bloblang.Executor
	upsert      bool
}

func writeMapsFromParsed(conf *service.ParsedConfig, operation Operation) (maps writeMaps, err error) {
	if probeStr, _ := conf.FieldString(commonFieldFilterMap); probeStr != "" {
		if maps.filterMap, err = conf.FieldBloblang(commonFieldFilterMap); err != nil {
			return maps, err
		}
	}
	if probeStr, _ := conf.FieldString(commonFieldDocumentMap); probeStr != "" {
		if maps.documentMap, err = conf.FieldBloblang(commonFieldDocumentMap); err != nil {
			return maps, err
		}
	}
	if probeStr, _ := conf.FieldString(commonFieldHintMap); probeStr != "" {
		if maps.hintMap, err = conf.FieldBloblang(commonFieldHintMap); err != nil {
			return maps, err
		}
	}
	if maps.upsert, err = conf.FieldBool(commonFieldUpsert); err != nil {
		return maps, err
	}

	if operation.isFilterAllowed() {
		if maps.filterMap == nil {
			err = errors.New("mongodb filter_map must be specified")
			return maps, err
		}
	} else if maps.filterMap != nil {
		err = fmt.Errorf("mongodb filter_map not allowed for '%s' operation", operation)
		return maps, err
	}

	if operation.isDocumentAllowed() {
		if maps.documentMap == nil {
			err = errors.New("mongodb document_map must be specified")
			return maps, err
		}
	} else if maps.documentMap != nil {
		err = fmt.Errorf("mongodb document_map not allowed for '%s' operation", operation)
		return maps, err
	}

	if !operation.isHintAllowed() && maps.hintMap != nil {
		err = fmt.Errorf("mongodb hint_map not allowed for '%s' operation", operation)
		return maps, err
	}

	if !operation.isUpsertAllowed() && maps.upsert {
		err = fmt.Errorf("mongodb upsert not allowed for '%s' operation", operation)
		return maps, err
	}

	return maps, nil
}

func (w writeMaps) extractFromMessage(operation Operation, i int, batch service.MessageBatch) (
	docJSON, filterJSON, hintJSON any, err error,
) {
	filterValWanted := operation.isFilterAllowed()
	documentValWanted := operation.isDocumentAllowed()

	if filterValWanted && w.filterMap != nil {
		if filterJSON, err = extJSONFromMap(batch, i, w.filterMap); err != nil {
			err = fmt.Errorf("failed to execute filter_map: %v", err)
			return
		}
	}

	if documentValWanted && w.documentMap != nil {
		if docJSON, err = extJSONFromMap(batch, i, w.documentMap); err != nil {
			err = fmt.Errorf("failed to execute document_map: %v", err)
			return
		}
	}

	if w.hintMap != nil {
		if hintJSON, err = extJSONFromMap(batch, i, w.hintMap); err != nil {
			return
		}
	}
	return
}

func extJSONFromMap(b service.MessageBatch, i int, m *bloblang.Executor) (any, error) {
	executor := b.BloblangExecutor(m)
	msg, err := executor.Query(i)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, nil
	}

	keyTypeMap, err := getKeyTypMap(msg)
	if err != nil {
		return nil, err
	}
	root, err := msg.AsStructured()
	if err != nil {
		return nil, err
	}
	bsonmap := neosync_mongodb.MarshalJSONToBSONDocument(root, keyTypeMap)
	return bsonmap, nil
}

func getKeyTypMap(p *service.Message) (map[string]neosync_types.KeyType, error) {
	keyTypeMap := map[string]neosync_types.KeyType{}
	meta, ok := p.MetaGetMut(neosync_benthos_metadata.MetaTypeMapStr)
	if ok {
		kt, err := convertToMapStringKeyType(meta)
		if err != nil {
			return nil, err
		}
		keyTypeMap = kt
	}
	ktm := map[string]neosync_types.KeyType{}
	for k, v := range keyTypeMap {
		if k == "_id" {
			ktm[k] = v
		}
		ktm[fmt.Sprintf("$set.%s", k)] = v
	}
	return ktm, nil
}

func convertToMapStringKeyType(i any) (map[string]neosync_types.KeyType, error) {
	if m, ok := i.(map[string]neosync_types.KeyType); ok {
		return m, nil
	}

	return nil, errors.New("input is not of type map[string]KeyType")
}
