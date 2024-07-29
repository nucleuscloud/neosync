import { Action } from '@/components/DualListBox/DualListBox';
import { DestinationDetails } from '@/components/jobs/NosqlTable/TableMappings/Columns';
import {
  JobMappingFormValues,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import {
  Connection,
  GetConnectionSchemaMapsResponse,
  GetConnectionSchemaResponse,
  JobDestination,
  JobMappingTransformer,
  JobSource,
} from '@neosync/sdk';

export function getConnectionIdFromSource(
  js: JobSource | undefined
): string | undefined {
  if (
    js?.options?.config.case === 'postgres' ||
    js?.options?.config.case === 'mysql' ||
    js?.options?.config.case === 'awsS3' ||
    js?.options?.config.case === 'mongodb' ||
    js?.options?.config.case === 'dynamodb'
  ) {
    return js.options.config.value.connectionId;
  }
  return undefined;
}

export function getFkIdFromGenerateSource(
  js: JobSource | undefined
): string | undefined {
  if (js?.options?.config.case === 'generate') {
    return js.options.config.value.fkSourceConnectionId;
  }
  if (js?.options?.config.case === 'aiGenerate') {
    return js.options.config.value.fkSourceConnectionId;
  }
  return undefined;
}

function getSetDelta(
  newSet: Set<string>,
  oldSet: Set<string>
): [Set<string>, Set<string>] {
  const added = new Set<string>();
  const removed = new Set<string>();

  oldSet.forEach((val) => {
    if (!newSet.has(val)) {
      removed.add(val);
    }
  });
  newSet.forEach((val) => {
    if (!oldSet.has(val)) {
      added.add(val);
    }
  });

  return [added, removed];
}

export function getOnSelectedTableToggle(
  schema: Record<string, GetConnectionSchemaResponse>,
  selectedTables: Set<string>,
  setSelectedTables: (newitems: Set<string>) => void,
  fields: { schema: string; table: string }[],
  remove: (indices: number[]) => void,
  append: (formValues: JobMappingFormValues[]) => void
): (items: Set<string>, action: Action) => void {
  return (items) => {
    if (items.size === 0) {
      const idxs = fields.map((_, idx) => idx);
      remove(idxs);
      setSelectedTables(new Set());
      return;
    }
    const [added, removed] = getSetDelta(items, selectedTables);
    const toRemove: number[] = [];
    const toAdd: JobMappingFormValues[] = [];
    fields.forEach((field, idx) => {
      if (removed.has(`${field.schema}.${field.table}`)) {
        toRemove.push(idx);
      }
    });

    added.forEach((item) => {
      const schemaResp = schema[item];
      if (!schemaResp) {
        return;
      }
      schemaResp.schemas.forEach((dbcol) => {
        toAdd.push({
          schema: dbcol.schema,
          table: dbcol.table,
          column: dbcol.column,
          transformer: convertJobMappingTransformerToForm(
            new JobMappingTransformer({})
          ),
        });
      });
    });
    if (toRemove.length > 0) {
      remove(toRemove);
    }
    if (toAdd.length > 0) {
      append(toAdd);
    }
    setSelectedTables(items);
  };
}

export function isNosqlSource(connection: Connection): boolean {
  switch (connection.connectionConfig?.config.case) {
    case 'mongoConfig':
    case 'dynamodbConfig':
      return true;
    default: {
      return false;
    }
  }
}

export function shouldShowDestinationTableMappings(
  sourceConnection: Connection,
  hasDynamoDbDestinations: boolean
): boolean {
  return isDynamoDBConnection(sourceConnection) && hasDynamoDbDestinations;
}

export function isDynamoDBConnection(connection: Connection): boolean {
  return connection.connectionConfig?.config.case === 'dynamodbConfig';
}

export function getDestinationDetailsRecord(
  dynamoDestinationConnections: Pick<JobDestination, 'id' | 'connectionId'>[],
  connectionsRecord: Record<string, Connection>,
  destinationSchemaMapsResp: GetConnectionSchemaMapsResponse
): Record<string, DestinationDetails> {
  const destSchemaRecord: Record<string, string[]> = {};
  destinationSchemaMapsResp.connectionIds.forEach((connid, idx) => {
    destSchemaRecord[connid] = Object.keys(
      destinationSchemaMapsResp.responses[idx].schemaMap
    ).map((table) => {
      const [, tableName] = table.split('.');
      return tableName;
    });
  });

  const output: Record<string, DestinationDetails> = {};

  dynamoDestinationConnections.forEach((d) => {
    const connection = connectionsRecord[d.connectionId];
    const availableTableNames = destSchemaRecord[d.connectionId] ?? [];
    output[d.id] = {
      destinationId: d.id,
      friendlyName: connection?.name ?? 'Unknown Name',
      availableTableNames,
    };
  });

  return output;
}

export function getDynamoDbDestinations(
  destinations: JobDestination[]
): JobDestination[] {
  return destinations.filter(
    (d) => d.options?.config.case === 'dynamodbOptions'
  );
}
