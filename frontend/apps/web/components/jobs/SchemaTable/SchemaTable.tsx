'use client';
import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import DualListBox, {
  Action,
  Option,
} from '@/components/DualListBox/DualListBox';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Transformer } from '@/shared/transformers';
import {
  JobMappingFormValues,
  JobMappingTransformerForm,
  SchemaFormValues,
  VirtualForeignConstraintFormValues,
} from '@/yup-validations/jobs';
import {
  GetConnectionSchemaResponse,
  JobMapping,
  ValidateJobMappingsResponse,
} from '@neosync/sdk';
import { TableIcon } from '@radix-ui/react-icons';
import { Row } from '@tanstack/react-table';
import { ReactElement, useMemo } from 'react';
import { FieldErrors } from 'react-hook-form';
import {
  getGeneratedStatement,
  getIdentityStatement,
} from '../JobMappingTable/AttributesCell';
import { JobMappingRow, SQL_COLUMNS } from '../JobMappingTable/Columns';
import JobMappingTable from '../JobMappingTable/JobMappingTable';
import FormErrorsCard, { ErrorLevel, FormError } from './FormErrorsCard';
import { ImportMappingsConfig } from './ImportJobMappingsButton';
import { getVirtualForeignKeysColumns } from './VirtualFkColumns';
import VirtualFkPageTable from './VirtualFkPageTable';
import { VirtualForeignKeyForm } from './VirtualForeignKeyForm';
import { JobType, SchemaConstraintHandler } from './schema-constraint-handler';
import { TransformerResult } from './transformer-handler';
import { useOnExportMappings } from './useOnExportMappings';
import { handleDataTypeBadge } from './util';

interface Props {
  data: JobMappingFormValues[];
  virtualForeignKeys?: VirtualForeignConstraintFormValues[];
  addVirtualForeignKey?: (vfk: VirtualForeignConstraintFormValues) => void;
  removeVirtualForeignKey?: (index: number) => void;
  jobType: JobType;
  schema: Record<string, GetConnectionSchemaResponse>;
  isSchemaDataReloading: boolean;
  constraintHandler: SchemaConstraintHandler;
  selectedTables: Set<string>;
  onSelectedTableToggle(items: Set<string>, action: Action): void;
  isJobMappingsValidating?: boolean;

  onValidate?(): void;

  formErrors: FormError[];
  onImportMappingsClick(
    jobmappings: JobMapping[],
    importConfig: ImportMappingsConfig
  ): void;
  onTransformerUpdate(index: number, config: JobMappingTransformerForm): void;
  getAvailableTransformers(index: number): TransformerResult;
  getTransformerFromField(index: number): Transformer;
  onTransformerBulkUpdate(
    indices: number[],
    config: JobMappingTransformerForm
  ): void;
  getAvailableTransformersForBulk(
    rows: Row<JobMappingRow>[]
  ): TransformerResult;
  getTransformerFromFieldValue(value: JobMappingTransformerForm): Transformer;
  onApplyDefaultClick(override: boolean): void;
}

export function SchemaTable(props: Props): ReactElement {
  const {
    data,
    virtualForeignKeys,
    addVirtualForeignKey,
    removeVirtualForeignKey,
    constraintHandler,
    jobType,
    schema,
    selectedTables,
    onSelectedTableToggle,
    formErrors,
    isJobMappingsValidating,
    onValidate,
    onImportMappingsClick,
    onTransformerUpdate,
    getAvailableTransformers,
    getTransformerFromField,
    getAvailableTransformersForBulk,
    getTransformerFromFieldValue,
    onApplyDefaultClick,
    onTransformerBulkUpdate,
  } = props;
  const tableData = useMemo((): JobMappingRow[] => {
    return data.map((d): JobMappingRow => {
      const colKey = {
        schema: d.schema,
        table: d.table,
        column: d.column,
      };
      const isPrimaryKey = constraintHandler.getIsPrimaryKey(colKey);
      const [isForeignKey, fkCols] = constraintHandler.getIsForeignKey(colKey);
      const [isVirtualForeignKey, vfkCols] =
        constraintHandler.getIsVirtualForeignKey(colKey);
      const isUnique = constraintHandler.getIsUniqueConstraint(colKey);

      const constraintPieces: string[] = [];
      if (isPrimaryKey) {
        constraintPieces.push('Primary Key');
      }
      if (isForeignKey) {
        fkCols.forEach((col) => constraintPieces.push(`Foreign Key: ${col}`));
      }
      if (isVirtualForeignKey) {
        vfkCols.forEach((col) =>
          constraintPieces.push(`Virtual Foreign Key: ${col}`)
        );
      }
      if (isUnique) {
        constraintPieces.push('Unique');
      }
      const constraints = constraintPieces.join('\n');

      const generatedType = constraintHandler.getGeneratedType(colKey);
      const identityType = constraintHandler.getIdentityType(colKey);

      const attributePieces: string[] = [];
      if (generatedType) {
        attributePieces.push(getGeneratedStatement(generatedType));
      } else if (identityType) {
        attributePieces.push(getIdentityStatement(identityType));
      }
      attributePieces.push(
        constraintHandler.getIsNullable(colKey) ? 'Is Nullable' : 'Not Nullable'
      );
      const attributes = attributePieces.join('\n');

      return {
        schema: d.schema,
        table: d.table,
        column: d.column,
        dataType: handleDataTypeBadge(constraintHandler.getDataType(colKey)),
        attributes: {
          value: attributes,
          generatedType: generatedType,
          identityType: identityType,
        },
        constraints: {
          value: constraints,
          foreignKey: [isForeignKey, fkCols],
          virtualForeignKey: [isVirtualForeignKey, vfkCols],
          isPrimaryKey: isPrimaryKey,
          isUnique: isUnique,
        },
        isNullable: constraintHandler.getIsNullable(colKey),
        transformer: d.transformer,
      };
    });
  }, [data]);

  const virtualForeignKeyColumns = useMemo(() => {
    return getVirtualForeignKeysColumns({ removeVirtualForeignKey });
  }, [removeVirtualForeignKey]);

  // it is imperative that this is stable to not cause infinite re-renders of the listbox(s)
  const dualListBoxOpts = useMemo(
    () => getDualListBoxOptions(new Set(Object.keys(schema)), data),
    [schema, data]
  );

  const { onClick: onExportMappingsClick } = useOnExportMappings<JobMappingRow>(
    {
      jobMappings: data,
    }
  );

  if (!data) {
    return <SkeletonTable />;
  }

  return (
    <div className="flex flex-col gap-10">
      <div className="flex flex-col md:flex-row gap-3">
        <Card className="w-full">
          <CardHeader className="flex flex-col gap-2">
            <div className="flex flex-row items-center gap-2">
              <div className="flex">
                <TableIcon className="h-4 w-4" />
              </div>
              <CardTitle>Table Selection</CardTitle>
            </div>
            <CardDescription>
              Select the tables that you want to transform and move them from
              the source to destination table.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <DualListBox
              options={dualListBoxOpts}
              selected={selectedTables}
              onChange={onSelectedTableToggle}
              mode={jobType === 'generate' ? 'single' : 'many'}
            />
          </CardContent>
        </Card>
        <FormErrorsCard
          formErrors={formErrors}
          isValidating={isJobMappingsValidating}
          onValidate={onValidate}
        />
      </div>

      {virtualForeignKeys && addVirtualForeignKey ? (
        <Tabs defaultValue="mappings">
          <TabsList>
            <TabsTrigger value="mappings">Transformer Mappings</TabsTrigger>
            <TabsTrigger value="virtualforeignkeys">
              Virtual Foreign Keys
            </TabsTrigger>
          </TabsList>
          <TabsContent value="mappings">
            <JobMappingTable
              data={tableData}
              columns={SQL_COLUMNS}
              onTransformerUpdate={onTransformerUpdate}
              getAvailableTransformers={getAvailableTransformers}
              getTransformerFromField={getTransformerFromField}
              onExportMappingsClick={onExportMappingsClick}
              onImportMappingsClick={onImportMappingsClick}
              isApplyDefaultTransformerButtonDisabled={data.length === 0}
              getAvalableTransformersForBulk={getAvailableTransformersForBulk}
              getTransformerFromFieldValue={getTransformerFromFieldValue}
              onTransformerBulkUpdate={onTransformerBulkUpdate}
              onApplyDefaultClick={onApplyDefaultClick}
              onDeleteRow={() =>
                console.warn('on delete row is not implemented')
              }
              onDuplicateRow={() =>
                console.warn('on duplicate row is not implemented')
              }
              canRenameColumn={() => false}
              onRowUpdate={() => console.warn('onRowUpdate is not implemented')}
              getAvailableCollectionsByRow={() => {
                console.warn('getAvailableCollections is not implemented');
                return [];
              }}
            />
          </TabsContent>
          <TabsContent value="virtualforeignkeys">
            <div className="flex flex-col gap-6 pt-4">
              <VirtualForeignKeyForm
                schema={schema}
                constraintHandler={constraintHandler}
                selectedTables={selectedTables}
                addVirtualForeignKey={addVirtualForeignKey}
              />
              <VirtualFkPageTable
                columns={virtualForeignKeyColumns}
                data={virtualForeignKeys}
              />
            </div>
          </TabsContent>
        </Tabs>
      ) : (
        <JobMappingTable
          data={tableData}
          columns={SQL_COLUMNS}
          onTransformerUpdate={onTransformerUpdate}
          getAvailableTransformers={getAvailableTransformers}
          getTransformerFromField={getTransformerFromField}
          onExportMappingsClick={onExportMappingsClick}
          onImportMappingsClick={onImportMappingsClick}
          isApplyDefaultTransformerButtonDisabled={data.length === 0}
          getAvalableTransformersForBulk={getAvailableTransformersForBulk}
          getTransformerFromFieldValue={getTransformerFromFieldValue}
          onTransformerBulkUpdate={onTransformerBulkUpdate}
          onApplyDefaultClick={onApplyDefaultClick}
          onDeleteRow={() => console.warn('on delete row is not implemented')}
          onDuplicateRow={() =>
            console.warn('on duplicate row is not implemented')
          }
          canRenameColumn={() => false}
          onRowUpdate={() => console.warn('onRowUpdate is not implemented')}
          getAvailableCollectionsByRow={() => {
            console.warn('getAvailableCollections is not implemented');
            return [];
          }}
        />
      )}
    </div>
  );
}

function getDualListBoxOptions(
  tables: Set<string>,
  jobmappings: JobMappingFormValues[]
): Option[] {
  jobmappings.forEach((jm) => tables.add(`${jm.schema}.${jm.table}`));
  return Array.from(tables).map((table): Option => ({ value: table }));
}

function extractAllFormErrors(
  errors: FieldErrors<SchemaFormValues | SingleTableSchemaFormValues>,
  values: JobMappingFormValues[],
  path = ''
): FormError[] {
  let messages: FormError[] = [];

  for (const key in errors) {
    let newPath = path;

    if (!isNaN(Number(key))) {
      const index = Number(key);
      if (index < values.length) {
        const value = values[index];
        const column = `${value.schema}.${value.table}.${value.column}`;
        newPath = path ? `${path}.${column}` : column;
      }
    }
    const error = (errors as any)[key as unknown as any] as any; // eslint-disable-line @typescript-eslint/no-explicit-any

    if (!error) {
      continue;
    }
    if (error.message) {
      messages.push({
        path: newPath,
        message: error.message,
        type: error.type,
        level: 'error',
      });
    } else {
      messages = messages.concat(extractAllFormErrors(error, values, newPath));
    }
  }
  return messages;
}

export function getAllFormErrors(
  formErrors: FieldErrors<SchemaFormValues | SingleTableSchemaFormValues>,
  values: JobMappingFormValues[],
  validationErrors: ValidateJobMappingsResponse | undefined
): FormError[] {
  let messages: FormError[] = [];
  const formErr = extractAllFormErrors(formErrors, values);
  if (!validationErrors) {
    return formErr;
  }
  const colErr = validationErrors.columnErrors.map((e) => {
    return {
      path: `${e.schema}.${e.table}.${e.column}`,
      message: e.errors.join('. '),
      level: 'error' as ErrorLevel,
    };
  });
  const colWarnings = validationErrors.columnWarnings.map((e) => {
    return {
      path: `${e.schema}.${e.table}.${e.column}`,
      message: e.warnings.join('. '),
      level: 'warning' as ErrorLevel,
    };
  });
  const dbErr = validationErrors.databaseErrors?.errors.map((e) => {
    return {
      path: '',
      message: e,
      level: 'error' as ErrorLevel,
    };
  });
  messages = messages.concat(colErr, formErr, colWarnings);
  if (dbErr) {
    messages = messages.concat(dbErr);
  }

  return messages;
}
