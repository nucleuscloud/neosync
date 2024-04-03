import { ConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { PlainMessage } from '@bufbuild/protobuf';
import {
  DatabaseColumn,
  ForeignConstraintTables,
  ForeignKey,
  PrimaryConstraint,
  TransformerDataType,
  UniqueConstraint,
} from '@neosync/sdk';

export type JobType = 'sync' | 'generate';

export interface SchemaConstraintHandler {
  getIsPrimaryKey(key: ColumnKey): boolean;
  getIsForeignKey(key: ColumnKey): [boolean, string[]];
  getIsNullable(key: ColumnKey): boolean;
  getDataType(key: ColumnKey): string;
  getConvertedDataType(key: ColumnKey): TransformerDataType; // Returns the databases data types transformed to the Neosync Transformer Data Types
  getIsInSchema(key: ColumnKey): boolean;
  getIsUniqueConstraint(key: ColumnKey): boolean;
  getHasDefault(key: ColumnKey): boolean;
}

interface ColumnKey {
  schema: string;
  table: string;
  column: string;
}

interface ColDetails {
  isPrimaryKey: boolean;
  fk: [boolean, string[]];
  isNullable: boolean;
  dataType: string;
  isUniqueConstraint: boolean;
}

export function getSchemaConstraintHandler(
  schema: ConnectionSchemaMap,
  primaryConstraints: Record<string, PrimaryConstraint>,
  foreignConstraints: Record<string, ForeignConstraintTables>,
  uniqueConstraints: Record<string, UniqueConstraint>
): SchemaConstraintHandler {
  const colmap = buildColDetailsMap(
    schema,
    primaryConstraints,
    foreignConstraints,
    uniqueConstraints
  );
  return {
    getDataType(key) {
      return colmap[fromColKey(key)]?.dataType ?? '';
    },
    getConvertedDataType(key) {
      const datatype = colmap[fromColKey(key)]?.dataType ?? '';
      if (datatype === '') {
        return TransformerDataType.UNSPECIFIED;
      }
      return dbDataTypeToTransformerDataType(datatype);
    },
    getIsForeignKey(key) {
      return colmap[fromColKey(key)]?.fk ?? [false, []];
    },
    getIsNullable(key) {
      return colmap[fromColKey(key)]?.isNullable ?? false;
    },
    getIsPrimaryKey(key) {
      return colmap[fromColKey(key)]?.isPrimaryKey ?? false;
    },
    getIsUniqueConstraint(key) {
      return colmap[fromColKey(key)]?.isUniqueConstraint ?? false;
    },
    getIsInSchema(key) {
      return !!colmap[fromColKey(key)];
    },
    getHasDefault(key) {
      return true; // todo - NEOS-969
    },
  };
}

function dbDataTypeToTransformerDataType(
  dataType: string
): TransformerDataType {
  const dt = postgresTypeToTransformerDataType(dataType);
  if (dt === TransformerDataType.UNSPECIFIED) {
    return mysqlTypeToTransformerDataType(dataType);
  }
  return dt;
}

function postgresTypeToTransformerDataType(
  postgresType: string
): TransformerDataType {
  const isArray = postgresType.endsWith('[]');
  const baseType = postgresType
    .replace('[]', '')
    .split('(')[0]
    .trim()
    .toLowerCase();

  if (isArray) {
    return TransformerDataType.ANY;
  }

  switch (baseType) {
    case 'bigint':
    case 'integer':
    case 'smallint':
    case 'bigserial':
    case 'serial':
      return TransformerDataType.INT64;
    case 'text':
    case 'varchar':
    case 'char':
    case 'citext':
    case 'character varying':
      return TransformerDataType.STRING;
    case 'boolean':
      return TransformerDataType.BOOLEAN;
    case 'real':
    case 'double precision':
    case 'numeric':
      return TransformerDataType.FLOAT64;
    case 'uuid':
      return TransformerDataType.UUID;
    case 'timestamp':
    case 'date':
    case 'time':
      return TransformerDataType.TIME;
    case 'json':
    case 'jsonb':
      return TransformerDataType.ANY;
    default:
      return TransformerDataType.UNSPECIFIED;
  }
}

function mysqlTypeToTransformerDataType(
  mysqlType: string
): TransformerDataType {
  const baseType = mysqlType.split('(')[0].trim().toLowerCase();

  switch (baseType) {
    case 'int':
    case 'integer':
    case 'smallint':
    case 'mediumint':
    case 'bigint':
    case 'tinyint': // Could be BOOLEAN, but treated as INT64 for consistency
      return TransformerDataType.INT64;
    case 'varchar':
    case 'text':
    case 'char':
    case 'enum':
    case 'set':
    case 'mediumtext':
    case 'longtext':
      return TransformerDataType.STRING;
    case 'tinyint(1)':
      return TransformerDataType.BOOLEAN;
    case 'float':
    case 'double':
    case 'decimal':
      return TransformerDataType.FLOAT64;
    case 'datetime':
    case 'timestamp':
    case 'date':
    case 'time':
    case 'year':
      return TransformerDataType.TIME;
    case 'json':
      return TransformerDataType.ANY;
    case 'uuid': // Note: MySQL doesn't have a native UUID type, but this is for compatibility
      return TransformerDataType.UUID;
    default:
      return TransformerDataType.UNSPECIFIED;
  }
}

function buildColDetailsMap(
  schema: ConnectionSchemaMap,
  primaryConstraints: Record<string, PrimaryConstraint>,
  foreignConstraints: Record<string, ForeignConstraintTables>,
  uniqueConstraints: Record<string, UniqueConstraint>
): Record<string, ColDetails> {
  const colmap: Record<string, ColDetails> = {};
  //<schema.table: dbCols>
  Object.entries(schema).forEach(([key, dbcols]) => {
    const tablePkeys = primaryConstraints[key] ?? new PrimaryConstraint();
    const primaryCols = new Set(tablePkeys.columns);
    const foreignFkeys =
      foreignConstraints[key] ?? new ForeignConstraintTables();
    const tableUniqueConstraints =
      uniqueConstraints[key] ?? new UniqueConstraint({});
    const uniqueConstraintCols = new Set(tableUniqueConstraints.columns);
    const fkConstraints = foreignFkeys.constraints;
    const fkconstraintsMap: Record<string, ForeignKey> = {};
    fkConstraints.forEach((constraint) => {
      fkconstraintsMap[constraint.column] =
        constraint.foreignKey ?? new ForeignKey();
    });

    dbcols.forEach((dbcol) => {
      const fk: ForeignKey | undefined = fkconstraintsMap[dbcol.column];
      colmap[fromDbCol(dbcol)] = {
        isNullable: dbcol.isNullable === 'YES',
        dataType: dbcol.dataType,
        fk: [
          fk !== undefined,
          fk !== undefined ? [`${fk.table}.${fk.column}`] : [],
        ],
        isPrimaryKey: primaryCols.has(dbcol.column),
        isUniqueConstraint: uniqueConstraintCols.has(dbcol.column),
      };
    });
  });
  return colmap;
}
function fromColKey(key: ColumnKey): string {
  return `${key.schema}.${key.table}.${key.column}`;
}
function fromDbCol(dbcol: PlainMessage<DatabaseColumn>): string {
  return `${dbcol.schema}.${dbcol.table}.${dbcol.column}`;
}
