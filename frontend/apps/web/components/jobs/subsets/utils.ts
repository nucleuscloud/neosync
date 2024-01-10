import { SubsetFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import { TableRow } from './subset-table/column';

interface DbCol {
  schema: string;
  table: string;
}
export function buildTableRowData(
  dbCols: DbCol[],
  existingSubsets: SubsetFormValues['subsets']
): Record<string, TableRow> {
  const tableMap: Record<string, TableRow> = {};

  dbCols.forEach((mapping) => {
    const key = buildRowKey(mapping.schema, mapping.table);
    tableMap[key] = { schema: mapping.schema, table: mapping.table };
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
