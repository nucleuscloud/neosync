package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
)

type PgxArray[T any] struct {
	pgtype.Array[T]
	colDataType string
}

// custom PGX array scanner
// properly handles scanning postgres arrays
func (a *PgxArray[T]) Scan(src any) error {
	m := pgtype.NewMap()
	// Register money types
	m.RegisterType(&pgtype.Type{
		Name:  "money",
		OID:   790,
		Codec: pgtype.TextCodec{},
	})
	m.RegisterType(&pgtype.Type{
		Name: "_money",
		OID:  791,
		Codec: &pgtype.ArrayCodec{
			ElementType: &pgtype.Type{
				Name:  "money",
				OID:   790,
				Codec: pgtype.TextCodec{},
			},
		},
	})

	// Register UUID types
	m.RegisterType(&pgtype.Type{
		Name:  "uuid",
		OID:   2950, // UUID type OID
		Codec: pgtype.TextCodec{},
	})

	m.RegisterType(&pgtype.Type{
		Name: "_uuid",
		OID:  2951,
		Codec: &pgtype.ArrayCodec{
			ElementType: &pgtype.Type{
				Name:  "uuid",
				OID:   2950,
				Codec: pgtype.TextCodec{},
			},
		},
	})

	// Register XML type
	m.RegisterType(&pgtype.Type{
		Name:  "xml",
		OID:   142,
		Codec: pgtype.TextCodec{},
	})

	m.RegisterType(&pgtype.Type{
		Name: "_xml",
		OID:  143,
		Codec: &pgtype.ArrayCodec{
			ElementType: &pgtype.Type{
				Name:  "xml",
				OID:   142,
				Codec: pgtype.TextCodec{},
			},
		},
	})

	// Try to get the type by OID first if colDataType is numeric
	var pgt *pgtype.Type
	var ok bool

	if oid, err := strconv.Atoi(a.colDataType); err == nil {
		pgt, ok = m.TypeForOID(uint32(oid)) //nolint:gosec
	} else {
		pgt, ok = m.TypeForName(strings.ToLower(a.colDataType))
	}
	if !ok {
		return fmt.Errorf("cannot convert to sql.Scanner: cannot find registered type for %s", a.colDataType)
	}

	v := &a.Array
	var bufSrc []byte
	if src != nil {
		switch src := src.(type) {
		case string:
			bufSrc = []byte(src)
		case []byte:
			bufSrc = src
		default:
			bufSrc = []byte(fmt.Sprint(src))
		}
	}

	return m.Scan(pgt.OID, pgtype.TextFormatCode, bufSrc, v)
}

type NullableJSON struct {
	json.RawMessage
	Valid bool
}

// Nullable JSON scanner
func (n *NullableJSON) Scan(value any) error {
	if value == nil {
		n.RawMessage, n.Valid = nil, false
		return nil
	}

	n.Valid = true
	switch v := value.(type) {
	case []byte:
		n.RawMessage = json.RawMessage(v)
		return nil
	case string:
		n.RawMessage = json.RawMessage(v)
		return nil
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", value, n.RawMessage)
	}
}

func (n *NullableJSON) Unmarshal() (any, error) {
	if !n.Valid {
		return nil, nil
	}
	var js any
	if err := json.Unmarshal(n.RawMessage, &js); err != nil {
		return nil, err
	}
	return js, nil
}

func SqlRowToPgTypesMap(rows *sql.Rows) (map[string]any, error) {
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	cTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	values := make([]any, len(columnNames))
	scanTargets := make([]any, 0, len(columnNames))
	// Debug log column types
	for _, ct := range cTypes {
		fmt.Printf("Column %s: type=%s, scanType=%v\n",
			ct.Name(),
			ct.DatabaseTypeName(),
			ct.ScanType())
	}
	for i := range values {
		dbTypeName := cTypes[i].DatabaseTypeName()
		switch {
		case isXmlDataType(dbTypeName):
			values[i] = &sql.NullString{}
			scanTargets = append(scanTargets, values[i])
		case IsJsonPgDataType(dbTypeName):
			values[i] = &NullableJSON{}
			scanTargets = append(scanTargets, values[i])
		case strings.EqualFold(dbTypeName, "_interval"):
			values[i] = &PgxArray[*pgtype.Interval]{colDataType: dbTypeName}
			scanTargets = append(scanTargets, values[i])
		case strings.EqualFold(dbTypeName, "_bytea") || strings.EqualFold(dbTypeName, "_varbit"):
			values[i] = &PgxArray[[]byte]{colDataType: dbTypeName}
			scanTargets = append(scanTargets, values[i])
		case strings.EqualFold(dbTypeName, "_bit"):
			values[i] = &PgxArray[*pgtype.Bits]{colDataType: dbTypeName}
			scanTargets = append(scanTargets, values[i])
		case strings.EqualFold(dbTypeName, "interval"):
			values[i] = &pgtype.Interval{}
			scanTargets = append(scanTargets, values[i])
		// case strings.EqualFold(dbTypeName, "timestampz"):
		// 	values[i] = &pgtype.Timestamptz{}
		// 	scanTargets = append(scanTargets, values[i])
		case isPgxPgArrayType(dbTypeName):
			values[i] = &PgxArray[any]{colDataType: dbTypeName}
			scanTargets = append(scanTargets, values[i])
		default:
			scanTargets = append(scanTargets, &values[i])
		}
	}
	if err := rows.Scan(scanTargets...); err != nil {
		return nil, err
	}
	// Debug log scanned values
	for i, v := range values {
		fmt.Printf("Scanned value for %s: %#v\n",
			columnNames[i],
			v)
	}

	jObj := parsePgRowValues(values, columnNames, cTypes)

	// Debug log final converted values
	for k, v := range jObj {
		fmt.Printf("Final value for %s: %#v\n", k, v)
	}

	jsonF, _ := json.MarshalIndent(jObj, "", " ")
	fmt.Printf("\n jObj: %s \n\n", string(jsonF))

	return jObj, nil
}

func parsePgRowValues(values []any, columnNames []string, columnTypes []*sql.ColumnType) map[string]any {
	jObj := map[string]any{}
	for i, v := range values {
		col := columnNames[i]
		colType := columnTypes[i].DatabaseTypeName()
		switch t := v.(type) {
		case nil:
			jObj[col] = t
		case *sql.NullString:
			var val any = nil
			if t.Valid {
				val = t.String
			}
			jObj[col] = val
		case *NullableJSON:
			js, err := t.Unmarshal()
			if err != nil {
				js = t
			}
			jObj[col] = js
		case *PgxArray[*pgtype.Interval]:
			ia, err := toIntervalArray(t)
			if err != nil {
				jObj[col] = t
				continue
			}
			jObj[col] = ia
		case *PgxArray[[]byte]:
			ba, err := toBinaryArray(t)
			if err != nil {
				jObj[col] = t
				continue
			}
			jObj[col] = ba
		case *PgxArray[*pgtype.Bits]:
			ba, err := toBitsArray(t)
			if err != nil {
				jObj[col] = t
				continue
			}
			jObj[col] = ba
		case *PgxArray[any]:
			jObj[col] = pgArrayToGoSlice(t)
		case *pgtype.Interval:
			if !t.Valid {
				jObj[col] = nil
				continue
			}
			neoInterval, err := neosynctypes.NewIntervalFromPgx(t)
			if err != nil {
				jObj[col] = t
				continue
			}
			jObj[col] = neoInterval
		default:
			switch {
			case strings.EqualFold(colType, "date"), strings.EqualFold(colType, "timestamp"), strings.EqualFold(colType, "timestamptz"):
				fmt.Println("NEOSYNC DATETIME")
				dt, err := neosynctypes.NewDateTimeFromPgx(t)
				if err != nil {
					jObj[col] = t
					continue
				}
				jObj[col] = dt
			case strings.EqualFold(colType, "bit"), strings.EqualFold(colType, "varbit"):
				bits, err := neosynctypes.NewBitsFromPgx(t)
				if err != nil {
					jObj[col] = t
					continue
				}
				jObj[col] = bits
			case strings.EqualFold(colType, "bytea"):
				binary, err := neosynctypes.NewBinaryFromPgx(t)
				if err != nil {
					jObj[col] = t
					continue
				}
				jObj[col] = binary
			default:
				jObj[col] = t
			}
		}
	}
	return jObj
}

func isXmlDataType(colDataType string) bool {
	return strings.EqualFold(colDataType, "xml")
}

func IsJsonPgDataType(dataType string) bool {
	return strings.EqualFold(dataType, "json") || strings.EqualFold(dataType, "jsonb")
}
func isPgxPgArrayType(dbTypeName string) bool {
	return strings.HasPrefix(dbTypeName, "_") || dbTypeName == "791"
}

func IsPgArrayColumnDataType(colDataType string) bool {
	return strings.HasSuffix(colDataType, "[]")
}

func toBinaryArray(array *PgxArray[[]byte]) (any, error) {
	if array.Elements == nil {
		return nil, nil
	}

	dim := array.Dimensions()
	if len(dim) > 1 {
		return nil, errors.ErrUnsupported
	}

	binaryArray, err := neosynctypes.NewBinaryArrayFromPgx(array.Elements, []neosynctypes.NeosyncTypeOption{})
	if err != nil {
		return nil, err
	}
	return binaryArray, nil
}

func toBitsArray(array *PgxArray[*pgtype.Bits]) (any, error) {
	if array.Elements == nil {
		return nil, nil
	}

	dim := array.Dimensions()
	if len(dim) > 1 {
		return nil, errors.ErrUnsupported
	}

	bitsArray, err := neosynctypes.NewBitsArrayFromPgx(array.Elements, []neosynctypes.NeosyncTypeOption{})
	if err != nil {
		return nil, err
	}
	return bitsArray, nil
}

func toIntervalArray(array *PgxArray[*pgtype.Interval]) (any, error) {
	if array.Elements == nil {
		return nil, nil
	}

	dim := array.Dimensions()
	if len(dim) > 1 {
		return nil, errors.ErrUnsupported
	}

	neoIntervalArray, err := neosynctypes.NewIntervalArrayFromPgx(array.Elements, []neosynctypes.NeosyncTypeOption{})
	if err != nil {
		return nil, err
	}
	return neoIntervalArray, nil
}

func pgArrayToGoSlice(array *PgxArray[any]) any {
	if array.Elements == nil {
		return nil
	}

	dim := array.Dimensions()
	if len(dim) > 1 {
		dims := []int{}
		for _, d := range dim {
			dims = append(dims, int(d.Length))
		}
		return CreateMultiDimSlice(dims, array.Elements)
	}

	return array.Elements
}

// converts flat slice to multi-dimensional slice
func CreateMultiDimSlice(dims []int, elements []any) any {
	if len(elements) == 0 {
		return elements
	}
	if len(dims) == 0 || len(dims) == 1 {
		return elements
	}
	firstDim := dims[0]

	// creates nested any slice where depth = # of dimensions
	// 2 dimensions creates [][]any{}
	sliceType := reflect.TypeOf([]any{})
	for i := 0; i < len(dims)-1; i++ {
		sliceType = reflect.SliceOf(sliceType)
	}
	slice := reflect.MakeSlice(sliceType, firstDim, firstDim)

	// handles multi-dimensional slices
	subSize := 1
	for _, dim := range dims[1:] {
		subSize *= dim
	}

	for i := 0; i < firstDim; i++ {
		start := i * subSize
		end := start + subSize
		subSlice := CreateMultiDimSlice(dims[1:], elements[start:end])
		slice.Index(i).Set(reflect.ValueOf(subSlice))
	}

	return slice.Interface()
}

// returns string in this form ARRAY[[a,b],[c,d]]
func FormatPgArrayLiteral(arr any, castType string) string {
	arrayLiteral := "ARRAY" + formatArrayLiteral(arr)
	if castType == "" {
		return arrayLiteral
	}

	return arrayLiteral + "::" + castType
}

func formatArrayLiteral(arr any) string {
	v := reflect.ValueOf(arr)

	if v.Kind() == reflect.Slice {
		result := "["
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				result += ","
			}
			result += formatArrayLiteral(v.Index(i).Interface())
		}
		result += "]"
		return result
	}

	switch val := arr.(type) {
	case map[string]any:
		return formatMapLiteral(val)
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''"))
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatMapLiteral(m map[string]any) string {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("%v", m)
	}

	return fmt.Sprintf("'%s'", string(jsonBytes))
}
