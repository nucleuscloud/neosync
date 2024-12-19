package gob

import (
	"encoding/gob"
	"time"

	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jackc/pgx/v5/pgtype"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
)

// need to register all the types that are used in the connection data service
// because we use interfaces for the types
func RegisterGobTypes() {
	gob.Register(map[string]any{})
	gob.Register(map[string]any{})
	gob.Register(pgtype.Numeric{}) // TODO fix this
	gob.Register([]any{})
	gob.Register(time.Time{})
	gob.Register([][]uint8{})
	gob.RegisterName("neosynctypes.NeosyncDateTime", &neosynctypes.NeosyncDateTime{})
	gob.RegisterName("neosynctypes.Bits", &neosynctypes.Bits{})
	gob.RegisterName("neosynctypes.Binary", &neosynctypes.Binary{})
	gob.RegisterName("neosynctypes.NeosyncArray", &neosynctypes.NeosyncArray{})
	gob.RegisterName("neosynctypes.Interval", &neosynctypes.Interval{})
	gob.RegisterName("dynamodb.AttributeValueMemberB", &dynamotypes.AttributeValueMemberB{})
	gob.RegisterName("dynamodb.AttributeValueMemberBOOL", &dynamotypes.AttributeValueMemberBOOL{})
	gob.RegisterName("dynamodb.AttributeValueMemberBS", &dynamotypes.AttributeValueMemberBS{})
	gob.RegisterName("dynamodb.AttributeValueMemberL", &dynamotypes.AttributeValueMemberL{})
	gob.RegisterName("dynamodb.AttributeValueMemberM", &dynamotypes.AttributeValueMemberM{})
	gob.RegisterName("dynamodb.AttributeValueMemberN", &dynamotypes.AttributeValueMemberN{})
	gob.RegisterName("dynamodb.AttributeValueMemberNS", &dynamotypes.AttributeValueMemberNS{})
	gob.RegisterName("dynamodb.AttributeValueMemberNULL", &dynamotypes.AttributeValueMemberNULL{})
	gob.RegisterName("dynamodb.AttributeValueMemberS", &dynamotypes.AttributeValueMemberS{})
	gob.RegisterName("dynamodb.AttributeValueMemberSS", &dynamotypes.AttributeValueMemberSS{})
}
