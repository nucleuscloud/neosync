import { ConnectionConfigCase } from '@/app/(mgmt)/[account]/connections/util';
import { SubsetFormValues } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import { Job, JobMapping } from '@neosync/sdk';
import { SubsetTableRow } from './SubsetTable/Columns';

// Valid ConnectionConfigCase types. Using Extract here to ensure they stay consistent with what is available in ConnectionConfigCase
export type ValidSubsetConnectionType = Extract<
  ConnectionConfigCase,
  'pgConfig' | 'mysqlConfig' | 'dynamodbConfig' | 'mssqlConfig'
>;

interface DbCol {
  schema: string;
  table: string;
}

export function buildTableRowData(
  dbCols: DbCol[],
  rootTables: Set<string>,
  existingSubsets: SubsetFormValues['subsets']
): Record<string, SubsetTableRow> {
  const tableMap: Record<string, SubsetTableRow> = {};

  dbCols.forEach((mapping) => {
    const key = buildRowKey(mapping.schema, mapping.table);
    tableMap[key] = {
      schema: mapping.schema,
      table: mapping.table,
      isRootTable: rootTables.has(key),
      where: undefined,
    };
  });
  existingSubsets.forEach((subset) => {
    const key = buildRowKey(subset.schema, subset.table);
    if (tableMap[key]) {
      tableMap[key].where = subset.whereClause;
    }
  });

  return tableMap;
}

export function buildRowKey(schema: string, table: string): string {
  return `${schema}.${table}`;
}

export function GetColumnsForSqlAutocomplete(
  mappings: JobMapping[],
  itemToEdit: SubsetTableRow | undefined
): string[] {
  let cols: string[] = [];
  mappings.map((row) => {
    if (row.schema == itemToEdit?.schema && row.table == itemToEdit.table) {
      cols.push(row.column);
    }
  });

  return cols;
}

export function isJobSubsettable(job: Job): boolean {
  switch (job.source?.options?.config.case) {
    case 'postgres':
    case 'mysql':
    case 'dynamodb':
    case 'mssql':
      return true;
    default:
      return false;
  }
}

export function isValidSubsetType(
  connectionType: ConnectionConfigCase | null
): connectionType is ValidSubsetConnectionType {
  return (
    connectionType === 'pgConfig' ||
    connectionType === 'mysqlConfig' ||
    connectionType === 'dynamodbConfig' ||
    connectionType === 'mssqlConfig'
  );
}

export function isSubsetRowCountSupported(
  connectionType: ConnectionConfigCase | null
): boolean {
  return (
    connectionType === 'pgConfig' ||
    connectionType === 'mysqlConfig' ||
    connectionType === 'mssqlConfig'
  );
}

export function isSubsetValidationSupported(
  connectionType: ConnectionConfigCase | null
): boolean {
  return (
    connectionType === 'pgConfig' ||
    connectionType === 'mysqlConfig' ||
    connectionType === 'mssqlConfig'
  );
}
