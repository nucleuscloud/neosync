import { ConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { PlainMessage } from '@bufbuild/protobuf';
import {
  DatabaseColumn,
  ForeignConstraintTables,
  ForeignKey,
  PrimaryConstraint,
  UniqueConstraint,
} from '@neosync/sdk';

export interface SchemaConstraintHandler {
  getIsPrimaryKey(key: ColumnKey): boolean;
  getIsForeignKey(key: ColumnKey): [boolean, string[]];
  getIsNullable(key: ColumnKey): boolean;
  getDataType(key: ColumnKey): string;
  getIsInSchema(key: ColumnKey): boolean;
  getIsUniqueConstraint(key: ColumnKey): boolean;
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
  };
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
