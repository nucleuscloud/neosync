package neosync_benthos_mongodb

import (
	"github.com/benthosdev/benthos/v4/public/service"
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
	commonFieldClientURL      = "url"
	commonFieldClientDatabase = "database"
)

func clientFields() []*service.ConfigField {
	return []*service.ConfigField{
		service.NewURLField(commonFieldClientURL).
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

func (op Operation) isDocumentAllowed() bool { //nolint: unused
	switch op {
	case OperationInsertOne,
		OperationReplaceOne,
		OperationUpdateOne:
		return true
	default:
		return false
	}
}

func (op Operation) isFilterAllowed() bool { //nolint: unused
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

func (op Operation) isHintAllowed() bool { //nolint: unused
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

func (op Operation) isUpsertAllowed() bool { //nolint: unused
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
	commonFieldOperation = "operation" //nolint: unused
)
