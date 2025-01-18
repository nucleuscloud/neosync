import { Action } from '@/components/DualListBox/DualListBox';
import {
  JobMappingRow,
  NosqlJobMappingRow,
} from '@/components/jobs/JobMappingTable/Columns';
import { DestinationDetails } from '@/components/jobs/NosqlTable/TableMappings/Columns';
import {
  JobType,
  SchemaConstraintHandler,
} from '@/components/jobs/SchemaTable/schema-constraint-handler';
import {
  TransformerConfigCase,
  TransformerHandler,
  TransformerResult,
} from '@/components/jobs/SchemaTable/transformer-handler';
import {
  fromNosqlRowDataToColKey,
  fromRowDataToColKey,
  getTransformerFilter,
} from '@/components/jobs/SchemaTable/util';
import { isValidSubsetType } from '@/components/jobs/subsets/utils';
import {
  JobMappingFormValues,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { create } from '@bufbuild/protobuf';
import {
  Connection,
  GetConnectionSchemaMapsResponse,
  GetConnectionSchemaResponse,
  JobDestination,
  JobMappingTransformerSchema,
  JobSource,
  SystemTransformer,
  UserDefinedTransformer,
} from '@neosync/sdk';
import { Row } from '@tanstack/react-table';

export function getConnectionIdFromSource(
  js: JobSource | undefined
): string | undefined {
  if (
    js?.options?.config.case === 'postgres' ||
    js?.options?.config.case === 'mysql' ||
    js?.options?.config.case === 'awsS3' ||
    js?.options?.config.case === 'mongodb' ||
    js?.options?.config.case === 'dynamodb' ||
    js?.options?.config.case === 'mssql'
  ) {
    return js.options.config.value.connectionId;
  }
  if (js?.options?.config.case === 'generate') {
    return js.options.config.value.fkSourceConnectionId;
  }
  if (js?.options?.config.case === 'aiGenerate') {
    return js.options.config.value.aiConnectionId;
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
            create(JobMappingTransformerSchema)
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

export function isConnectionSubsettable(connection: Connection): boolean {
  return isValidSubsetType(connection.connectionConfig?.config.case ?? null);
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

export function getFilteredTransformersForBulkSet(
  rows: Row<JobMappingRow>[] | Row<NosqlJobMappingRow>[],
  transformerHandler: TransformerHandler,
  constraintHandler: SchemaConstraintHandler,
  jobType: JobType,
  sqlType: 'relational' | 'nosql'
): TransformerResult {
  const systemArrays: SystemTransformer[][] = [];
  const userDefinedArrays: UserDefinedTransformer[][] = [];

  rows.forEach((row) => {
    const colkey =
      sqlType === 'nosql'
        ? fromNosqlRowDataToColKey(row as Row<NosqlJobMappingRow>)
        : fromRowDataToColKey(row as Row<JobMappingRow>);
    const { system, userDefined } = transformerHandler.getFilteredTransformers(
      getTransformerFilter(constraintHandler, colkey, jobType)
    );
    systemArrays.push(system);
    userDefinedArrays.push(userDefined);
  });

  const uniqueSystemConfigCases = findCommonSystemConfigCases(systemArrays);
  const uniqueSystem = uniqueSystemConfigCases
    .map((configCase) =>
      transformerHandler.getSystemTransformerByConfigCase(configCase)
    )
    .filter((x): x is SystemTransformer => !!x);

  const uniqueIds = findCommonUserDefinedIds(userDefinedArrays);
  const uniqueUserDef = uniqueIds
    .map((id) => transformerHandler.getUserDefinedTransformerById(id))
    .filter((x): x is UserDefinedTransformer => !!x);

  return {
    system: uniqueSystem,
    userDefined: uniqueUserDef,
  };
}

function findCommonSystemConfigCases(
  arrays: SystemTransformer[][]
): TransformerConfigCase[] {
  const elementCount: Record<TransformerConfigCase, number> = {} as Record<
    TransformerConfigCase,
    number
  >;
  const subArrayCount = arrays.length;
  const commonElements: TransformerConfigCase[] = [];

  arrays.forEach((subArray) => {
    // Use a Set to ensure each element in a sub-array is counted only once
    new Set(subArray).forEach((element) => {
      if (!element.config?.config.case) {
        return;
      }
      if (!elementCount[element.config.config.case]) {
        elementCount[element.config.config.case] = 1;
      } else {
        elementCount[element.config.config.case]++;
      }
    });
  });

  for (const [element, count] of Object.entries(elementCount)) {
    if (count === subArrayCount) {
      commonElements.push(element as TransformerConfigCase);
    }
  }

  return commonElements;
}

function findCommonUserDefinedIds(
  arrays: UserDefinedTransformer[][]
): string[] {
  const elementCount: Record<string, number> = {};
  const subArrayCount = arrays.length;
  const commonElements: string[] = [];

  arrays.forEach((subArray) => {
    // Use a Set to ensure each element in a sub-array is counted only once
    new Set(subArray).forEach((element) => {
      if (!elementCount[element.id]) {
        elementCount[element.id] = 1;
      } else {
        elementCount[element.id]++;
      }
    });
  });

  for (const [element, count] of Object.entries(elementCount)) {
    if (count === subArrayCount) {
      commonElements.push(element);
    }
  }

  return commonElements;
}
