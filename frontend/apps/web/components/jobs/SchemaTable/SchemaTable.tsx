'use client';
import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import DualListBox, {
  Action,
  Option,
} from '@/components/DualListBox/DualListBox';
import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useGetTransformersHandler } from '@/libs/hooks/useGetTransformersHandler';
import {
  JobMappingFormValues,
  SchemaFormValues,
  VirtualForeignConstraintFormValues,
} from '@/yup-validations/jobs';
import {
  GetConnectionSchemaResponse,
  JobMapping,
  ValidateJobMappingsResponse,
} from '@neosync/sdk';
import { TableIcon } from '@radix-ui/react-icons';
import { ReactElement, useMemo } from 'react';
import { FieldErrors } from 'react-hook-form';
import FormErrorsCard, { FormError } from './FormErrorsCard';
import { ImportMappingsConfig } from './ImportJobMappingsButton';
import { getSchemaColumns } from './SchemaColumns';
import SchemaPageTable from './SchemaPageTable';
import { getVirtualForeignKeysColumns } from './VirtualFkColumns';
import VirtualFkPageTable from './VirtualFkPageTable';
import { VirtualForeignKeyForm } from './VirtualForeignKeyForm';
import { JobType, SchemaConstraintHandler } from './schema-constraint-handler';
import { useOnExportMappings } from './useOnExportMappings';

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
  } = props;
  const { account } = useAccount();
  const { handler, isLoading, isValidating } = useGetTransformersHandler(
    account?.id ?? ''
  );

  const columns = useMemo(() => {
    return getSchemaColumns({
      transformerHandler: handler,
      constraintHandler,
      jobType,
    });
  }, [handler, constraintHandler, jobType]);

  const virtualForeignKeyColumns = useMemo(() => {
    return getVirtualForeignKeysColumns({ removeVirtualForeignKey });
  }, [removeVirtualForeignKey]);

  // it is imperative that this is stable to not cause infinite re-renders of the listbox(s)
  const dualListBoxOpts = useMemo(
    () => getDualListBoxOptions(new Set(Object.keys(schema)), data),
    [schema, data]
  );

  const { onClick: onExportMappingsClick } = useOnExportMappings({
    jobMappings: data,
  });

  if (isLoading || !data) {
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
              <div>{isValidating ? <Spinner /> : null}</div>
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
            <SchemaPageTable
              columns={columns}
              data={data}
              transformerHandler={handler}
              constraintHandler={constraintHandler}
              jobType={jobType}
              onExportMappingsClick={onExportMappingsClick}
              onImportMappingsClick={onImportMappingsClick}
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
        <SchemaPageTable
          columns={columns}
          data={data}
          transformerHandler={handler}
          constraintHandler={constraintHandler}
          jobType={jobType}
          onExportMappingsClick={onExportMappingsClick}
          onImportMappingsClick={onImportMappingsClick}
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
    };
  });
  const dbErr = validationErrors.databaseErrors?.errors.map((e) => {
    return {
      path: '',
      message: e,
    };
  });
  messages = messages.concat(colErr, formErr);
  if (dbErr) {
    messages = messages.concat(dbErr);
  }

  return messages;
}
