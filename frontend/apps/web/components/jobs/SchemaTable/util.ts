import { Row } from '@tanstack/react-table';
import { JobMappingRow, NosqlJobMappingRow } from '../JobMappingTable/Columns';
import {
  ColumnKey,
  JobType,
  SchemaConstraintHandler,
} from './schema-constraint-handler';
import { toSupportedJobtype, TransformerFilters } from './transformer-handler';

export function fromRowDataToColKey(row: Row<JobMappingRow>): ColumnKey {
  return {
    schema: row.getValue('schema'),
    table: row.getValue('table'),
    column: row.getValue('column'),
  };
}
export function fromNosqlRowDataToColKey(
  row: Row<NosqlJobMappingRow>
): ColumnKey {
  const [schema, table] = splitCollection(row.getValue('collection'));
  return {
    schema,
    table,
    column: row.getValue('column'),
  };
}

export function toColKey(
  schema: string,
  table: string,
  column: string
): ColumnKey {
  return {
    schema,
    table,
    column,
  };
}

// cleans up the data type values since some are too long , can add on more here
export function handleDataTypeBadge(dataType: string): string {
  // Check for "timezone" and replace accordingly without entering the switch
  if (dataType.includes('timezone')) {
    return dataType
      .replace('timestamp with time zone', 'timestamp(tz)')
      .replace('timestamp without time zone', 'timestamp');
  }

  const splitDt = dataType.split('(');
  switch (splitDt[0]) {
    case 'character varying':
      // The condition inside the if statement seemed reversed. It should return 'varchar' directly if splitDt[1] is undefined.
      return splitDt[1] !== undefined ? `varchar(${splitDt[1]}` : 'varchar';
    default:
      return dataType;
  }
}

export function getTransformerFilter(
  constraintHandler: SchemaConstraintHandler,
  colkey: ColumnKey,
  jobType: JobType
): TransformerFilters {
  const [isForeignKey] = constraintHandler.getIsForeignKey(colkey);
  const [isVirtualForeignKey] =
    constraintHandler.getIsVirtualForeignKey(colkey);
  const isNullable = constraintHandler.getIsNullable(colkey);
  const convertedDataType = constraintHandler.getConvertedDataType(colkey);
  const hasDefault = constraintHandler.getHasDefault(colkey);
  const isGenerated = constraintHandler.getIsGenerated(colkey);
  return {
    dataType: convertedDataType,
    hasDefault,
    isForeignKey,
    isVirtualForeignKey,
    isNullable,
    jobType: toSupportedJobtype(jobType),
    isGenerated,
    identityType: constraintHandler.getIdentityType(colkey),
  };
}

export function splitCollection(collection: string): [string, string] {
  const lastDotIndex = collection.lastIndexOf('.');
  const schema = collection.substring(0, lastDotIndex);
  const table = collection.substring(lastDotIndex + 1);
  return [schema, table];
}
