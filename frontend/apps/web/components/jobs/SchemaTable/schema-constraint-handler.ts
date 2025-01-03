import { create } from '@bufbuild/protobuf';
import {
  DatabaseColumn,
  ForeignConstraintTables,
  ForeignConstraintTablesSchema,
  ForeignKey,
  ForeignKeySchema,
  GetConnectionSchemaResponse,
  PrimaryConstraint,
  PrimaryConstraintSchema,
  TransformerDataType,
  UniqueConstraints,
  UniqueConstraintsSchema,
  VirtualForeignConstraint,
  VirtualForeignKey,
  VirtualForeignKeySchema,
} from '@neosync/sdk';

export type JobType = 'sync' | 'generate';

export interface SchemaConstraintHandler {
  getIsPrimaryKey(key: ColumnKey): boolean;
  getIsForeignKey(key: ColumnKey): [boolean, string[]];
  getIsVirtualForeignKey(key: ColumnKey): [boolean, string[]];
  getIsNullable(key: ColumnKey): boolean;
  getDataType(key: ColumnKey): string;
  getConvertedDataType(key: ColumnKey): TransformerDataType; // Returns the databases data types transformed to the Neosync Transformer Data Types
  getIsInSchema(key: ColumnKey): boolean;
  getIsUniqueConstraint(key: ColumnKey): boolean;
  getHasDefault(key: ColumnKey): boolean;
  getIsGenerated(key: ColumnKey): boolean;
  getGeneratedType(key: ColumnKey): string | undefined;
  getIdentityType(key: ColumnKey): string | undefined;
}

export interface ColumnKey {
  schema: string;
  table: string;
  column: string;
}

interface ColDetails {
  isPrimaryKey: boolean;
  fk: [boolean, string[]];
  virtualForeignKey: [boolean, string[]];
  isNullable: boolean;
  dataType: string;
  isUniqueConstraint: boolean;
  columnDefault?: string;
  generatedType?: string;
  identityGeneration?: string;
}

export function getSchemaConstraintHandler(
  schema: Record<string, GetConnectionSchemaResponse>,
  primaryConstraints: Record<string, PrimaryConstraint>,
  foreignConstraints: Record<string, ForeignConstraintTables>,
  uniqueConstraints: Record<string, UniqueConstraints>,
  virtualForeignConstraints: VirtualForeignConstraint[]
): SchemaConstraintHandler {
  const vfkMap = virtualForeignConstraints.reduce(
    (vfkMap, vfk) => {
      const key = `${vfk.schema}.${vfk.table}`;
      const vfkArray = vfkMap[key];
      if (!vfkArray) {
        vfkMap[key] = [vfk];
      } else {
        vfkMap[key].push(vfk);
      }
      return vfkMap;
    },
    {} as Record<string, VirtualForeignConstraint[]>
  );
  const colmap = buildColDetailsMap(
    schema,
    primaryConstraints,
    foreignConstraints,
    uniqueConstraints,
    vfkMap
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
    getIsVirtualForeignKey(key) {
      return colmap[fromColKey(key)]?.virtualForeignKey ?? [false, []];
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
      const ckey = fromColKey(key);
      return (
        !!colmap[ckey]?.columnDefault || !!colmap[ckey]?.identityGeneration
      );
    },
    getIsGenerated(key) {
      return !!colmap[fromColKey(key)]?.generatedType;
    },
    getGeneratedType(key) {
      return colmap[fromColKey(key)]?.generatedType;
    },
    getIdentityType(key) {
      return colmap[fromColKey(key)]?.identityGeneration;
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
  schema: Record<string, GetConnectionSchemaResponse>,
  primaryConstraints: Record<string, PrimaryConstraint>,
  foreignConstraints: Record<string, ForeignConstraintTables>,
  uniqueConstraints: Record<string, UniqueConstraints>,
  virtualForeignConstraints: Record<string, VirtualForeignConstraint[]>
): Record<string, ColDetails> {
  const colmap: Record<string, ColDetails> = {};
  //<schema.table: dbCols>
  Object.entries(schema).forEach(([key, schemaResp]) => {
    const tablePkeys =
      primaryConstraints[key] ?? create(PrimaryConstraintSchema);
    const primaryCols = new Set(tablePkeys.columns);
    const foreignFkeys =
      foreignConstraints[key] ?? create(ForeignConstraintTablesSchema);
    const virtualForeignKeys = virtualForeignConstraints[key] ?? [];
    const tableUniqueConstraints =
      uniqueConstraints[key] ?? create(UniqueConstraintsSchema);
    const uniqueConstraintCols = tableUniqueConstraints.constraints.reduce(
      (prev, curr) => {
        curr.columns.forEach((c) => prev.add(c));
        return prev;
      },
      new Set()
    );
    const fkConstraints = foreignFkeys.constraints;
    const fkconstraintsMap: Record<string, ForeignKey> = {};
    fkConstraints.forEach((constraint) => {
      constraint.columns.forEach((col, idx) => {
        if (constraint.foreignKey) {
          fkconstraintsMap[col] = create(ForeignKeySchema, {
            table: constraint.foreignKey?.table,
            columns: [constraint.foreignKey?.columns[idx]],
          });
        } else {
          fkconstraintsMap[col] = create(ForeignKeySchema);
        }
      });
    });

    const virtualFkMap: Record<string, VirtualForeignKey> = {};
    virtualForeignKeys.forEach((vfk) => {
      vfk.columns.forEach((col, idx) => {
        if (vfk.foreignKey) {
          virtualFkMap[col] = create(VirtualForeignKeySchema, {
            schema: vfk.foreignKey.schema,
            table: vfk.foreignKey?.table,
            columns: [vfk.foreignKey?.columns[idx]],
          });
        } else {
          virtualFkMap[col] = create(VirtualForeignKeySchema);
        }
      });
    });

    schemaResp.schemas.forEach((dbcol) => {
      const fk: ForeignKey | undefined = fkconstraintsMap[dbcol.column];
      const vfk: VirtualForeignKey | undefined = virtualFkMap[dbcol.column];
      colmap[fromDbCol(dbcol)] = {
        isNullable: dbcol.isNullable === 'YES',
        dataType: dbcol.dataType,
        fk: [
          fk !== undefined,
          fk !== undefined
            ? fk.columns.map((column) => `${fk.table}.${column}`)
            : [],
        ],
        virtualForeignKey: [
          vfk !== undefined,
          vfk !== undefined
            ? vfk.columns.map(
                (column) => `${vfk.schema}.${vfk.table}.${column}`
              )
            : [],
        ],
        isPrimaryKey: primaryCols.has(dbcol.column),
        isUniqueConstraint: uniqueConstraintCols.has(dbcol.column),
        columnDefault: dbcol.columnDefault,
        generatedType: dbcol.generatedType,
        identityGeneration: dbcol.identityGeneration,
      };
    });
  });
  return colmap;
}
function fromColKey(key: ColumnKey): string {
  return `${key.schema}.${key.table}.${key.column}`;
}
function fromDbCol(dbcol: DatabaseColumn): string {
  return `${dbcol.schema}.${dbcol.table}.${dbcol.column}`;
}
