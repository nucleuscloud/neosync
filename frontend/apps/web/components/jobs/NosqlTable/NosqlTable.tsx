import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { useGetTransformersHandler } from '@/libs/hooks/useGetTransformersHandler';
import { Transformer } from '@/shared/transformers';
import {
  EditDestinationOptionsFormValues,
  JobMappingFormValues,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import { GetConnectionSchemaResponse, JobMapping } from '@neosync/sdk';
import { TableIcon } from '@radix-ui/react-icons';
import { Row } from '@tanstack/react-table';
import { nanoid } from 'nanoid';
import { ReactElement, useCallback, useMemo } from 'react';
import { NOSQL_COLUMNS, NosqlJobMappingRow } from '../JobMappingTable/Columns';
import JobMappingTable from '../JobMappingTable/JobMappingTable';
import FormErrorsCard, { FormError } from '../SchemaTable/FormErrorsCard';
import { ImportMappingsConfig } from '../SchemaTable/ImportJobMappingsButton';
import { TransformerResult } from '../SchemaTable/transformer-handler';
import { useOnExportMappings } from '../SchemaTable/useOnExportMappings';
import { splitCollection } from '../SchemaTable/util';
import AddNewNosqlRecord, {
  AddNewNosqlRecordFormValues,
} from './AddNewNosqlRecord';
import {
  DestinationDetails,
  OnTableMappingUpdateRequest,
} from './TableMappings/Columns';
import TableMappingsCard from './TableMappings/TableMappingsCard';

interface Props {
  data: JobMappingFormValues[];
  schema: Record<string, GetConnectionSchemaResponse>;
  isSchemaDataReloading: boolean;
  isJobMappingsValidating?: boolean;

  onValidate?(): void;

  formErrors: FormError[];
  onAddMappings(values: AddNewNosqlRecordFormValues[]): void;
  onRemoveMappings(indices: number[]): void;
  onEditMappings(values: JobMappingFormValues, index: number): void;

  destinationOptions: EditDestinationOptionsFormValues[];
  destinationDetailsRecord: Record<string, DestinationDetails>;
  onDestinationTableMappingUpdate(req: OnTableMappingUpdateRequest): void;
  showDestinationTableMappings: boolean;
  onImportMappingsClick(
    jobmappings: JobMapping[],
    importConfig: ImportMappingsConfig
  ): void;
  getAvailableTransformers(index: number): TransformerResult;
  getTransformerFromField(index: number): Transformer;
  onApplyDefaultClick(override: boolean): void;
  getAvailableTransformersForBulk(
    rows: Row<NosqlJobMappingRow>[]
  ): TransformerResult;
  getTransformerFromFieldValue(value: JobMappingTransformerForm): Transformer;
  onTransformerBulkUpdate(
    indices: number[],
    config: JobMappingTransformerForm
  ): void;
  hasMissingSourceColumnMappings: boolean;
  onRemoveMissingSourceColumnMappings(): void;
}

export default function NosqlTable(props: Props): ReactElement {
  const {
    data,
    schema,
    formErrors,
    isJobMappingsValidating,
    onValidate,
    onAddMappings,
    onRemoveMappings,
    onEditMappings,
    destinationOptions,
    destinationDetailsRecord,
    onDestinationTableMappingUpdate,
    showDestinationTableMappings,
    onImportMappingsClick,
    getAvailableTransformers,
    getTransformerFromField,
    onApplyDefaultClick,
    getAvailableTransformersForBulk,
    getTransformerFromFieldValue,
    onTransformerBulkUpdate,
    hasMissingSourceColumnMappings,
    onRemoveMissingSourceColumnMappings,
  } = props;
  const { account } = useAccount();
  const { handler, isValidating } = useGetTransformersHandler(
    account?.id ?? ''
  );

  const collections = Array.from(Object.keys(schema));

  // useMemo ensures that we don't recreate the set unless the data changes
  const keySet = useMemo(() => {
    const set = new Set<string>();
    data.forEach((item: JobMappingFormValues) => {
      set.add(`${item.schema}.${item.table}.${item.column}`);
    });
    return set;
  }, [data]);

  // useCallback ensures that we only re-run the function if the keySet changes
  const isDuplicateKey = useCallback(
    (index: number, newValue: string) => {
      const row = data[index];
      const key = `${row.schema}.${row.table}.${newValue}`;
      return keySet.has(key);
    },
    [keySet]
  );

  // used to calculate the collections that can be updated based on a given key value
  // for ex. if a collection.key is "a.b.c" and we want to update it to "d.e.c" but "d.e.c" already exists, then we don't want to show the "d.e." collection as an uption for the update
  const getAvailableCollectionsByRow = useCallback(
    (index: number) => {
      const currentColumn = data[index].column;
      const currentSchemaTable = `${data[index].schema}.${data[index].table}`;

      const conflictRows = data.filter(
        (obj) =>
          obj.column === currentColumn &&
          `${obj.schema}.${obj.table}` !== currentSchemaTable
      );

      return collections.filter(
        (item) =>
          !conflictRows.some((obj) => `${obj.schema}.${obj.table}` === item)
      );
    },
    [data, collections]
  );

  const tableData = useMemo(() => {
    return data.map((d): NosqlJobMappingRow => {
      return {
        collection: `${d.schema}.${d.table}`,
        column: d.column,
        transformer: d.transformer,
      };
    });
  }, [data.length]);

  const { onClick: onExportMappingsClick } =
    useOnExportMappings<NosqlJobMappingRow>({
      jobMappings: data,
    });

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col md:flex-row gap-3">
        <Card className="w-full">
          <CardHeader className="flex flex-col gap-2">
            <div className="flex flex-row items-center gap-2">
              <div className="flex">
                <TableIcon className="h-4 w-4" />
              </div>
              <CardTitle>Add new mapping</CardTitle>
              <div>{isValidating ? <Spinner /> : null}</div>
            </div>
            <CardDescription>
              Select a collection and input a document key to transform, along
              with specifying the relevant transformer.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AddNewNosqlRecord
              collections={collections}
              isDuplicateKey={(newVal, schema, table) => {
                return keySet.has(`${schema}.${table}.${newVal}`);
              }}
              onSubmit={(values) => {
                onAddMappings([values]);
              }}
              transformerHandler={handler}
            />
          </CardContent>
        </Card>
        <FormErrorsCard
          formErrors={formErrors}
          isValidating={isJobMappingsValidating}
          onValidate={onValidate}
        />
      </div>
      {showDestinationTableMappings && (
        <div>
          <TableMappingsCard
            mappings={destinationOptions}
            onUpdate={onDestinationTableMappingUpdate}
            destinationDetailsRecord={destinationDetailsRecord}
          />
        </div>
      )}
      <JobMappingTable
        data={tableData}
        columns={NOSQL_COLUMNS}
        onTransformerUpdate={(idx, config) => {
          const row = data[idx];
          onEditMappings({ ...row, transformer: config }, idx);
        }}
        getAvailableTransformers={getAvailableTransformers}
        getTransformerFromField={getTransformerFromField}
        onExportMappingsClick={onExportMappingsClick}
        onImportMappingsClick={onImportMappingsClick}
        isApplyDefaultTransformerButtonDisabled={data.length === 0}
        displayApplyDefaultTransformersButton={true}
        getAvalableTransformersForBulk={getAvailableTransformersForBulk}
        getTransformerFromFieldValue={getTransformerFromFieldValue}
        onTransformerBulkUpdate={onTransformerBulkUpdate}
        onApplyDefaultClick={onApplyDefaultClick}
        onDeleteRow={(idx) => onRemoveMappings([idx])}
        onDuplicateRow={(idx) => {
          const row = data[idx];
          onAddMappings([
            {
              collection: `${row.schema}.${row.table}`,
              key: createDuplicateKey(row.column),
              transformer: row.transformer,
            },
          ]);
        }}
        canRenameColumn={isDuplicateKey}
        onRowUpdate={(idx, val) => {
          const [schema, table] = splitCollection(val.collection);
          onEditMappings(
            {
              ...val,
              schema,
              table,
            },
            idx
          );
        }}
        getAvailableCollectionsByRow={getAvailableCollectionsByRow}
        hasMissingSourceColumnMappings={hasMissingSourceColumnMappings}
        onRemoveMissingSourceColumnMappings={
          onRemoveMissingSourceColumnMappings
        }
      />
    </div>
  );
}

function createDuplicateKey(key: string): string {
  const uniqueSuffix = nanoid(6);
  return `${key}_${uniqueSuffix}`;
}
