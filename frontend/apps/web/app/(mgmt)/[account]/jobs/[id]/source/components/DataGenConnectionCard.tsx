'use client';
import {
  SINGLE_TABLE_SCHEMA_FORM_SCHEMA,
  SingleTableSchemaFormValues,
} from '@/app/(mgmt)/[account]/new/job/schema';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/SchemaColumns';
import { SchemaTable } from '@/components/jobs/SchemaTable/SchemaTable';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnectionForeignConstraints } from '@/libs/hooks/useGetConnectionForeignConstraints';
import { useGetConnectionPrimaryConstraints } from '@/libs/hooks/useGetConnectionPrimaryConstraints';
import { useGetConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { getErrorMessage } from '@/util/util';
import {
  JobMappingTransformerForm,
  convertJobMappingTransformerFormToJobMappingTransformer,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { PlainMessage } from '@bufbuild/protobuf';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  DatabaseColumn,
  GenerateSourceOptions,
  GenerateSourceSchemaOption,
  GenerateSourceTableOption,
  Job,
  JobMapping,
  JobMappingTransformer,
  JobSource,
  JobSourceOptions,
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ReactElement, useMemo } from 'react';
import { useForm } from 'react-hook-form';
import { getFkIdFromGenerateSource } from './util';

interface Props {
  jobId: string;
}

export default function DataGenConnectionCard({ jobId }: Props): ReactElement {
  const { toast } = useToast();
  const { account } = useAccount();

  const {
    data,
    mutate,
    isLoading: isJobLoading,
  } = useGetJob(account?.id ?? '', jobId);
  const fkSourceConnectionId = getFkIdFromGenerateSource(data?.job?.source);

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaDataMapLoading,
    isValidating: isSchemaMapValidating,
  } = useGetConnectionSchemaMap(account?.id ?? '', fkSourceConnectionId ?? '');

  const { data: primaryConstraints, isValidating: isPkValidating } =
    useGetConnectionPrimaryConstraints(
      account?.id ?? '',
      fkSourceConnectionId ?? ''
    );

  const { data: foreignConstraints, isValidating: isFkValidating } =
    useGetConnectionForeignConstraints(
      account?.id ?? '',
      fkSourceConnectionId ?? ''
    );

  const allJobMappings =
    Object.values(connectionSchemaDataMap?.schemaMap ?? {}).flatMap(
      (dbcols) => {
        const t: JobMappingTransformerForm = convertJobMappingTransformerToForm(
          new JobMappingTransformer({})
        );
        return dbcols.map((dbcol) => ({ ...dbcol, transformer: t }));
      }
    ) || [];
  const form = useForm<SingleTableSchemaFormValues>({
    resolver: yupResolver(SINGLE_TABLE_SCHEMA_FORM_SCHEMA),
    defaultValues: {
      mappings: [],
      numRows: 0,
      schema: '',
      table: '',
    },
    values: getJobSource(data?.job),
  });

  const schemaConstraintHandler = useMemo(
    () =>
      getSchemaConstraintHandler(
        connectionSchemaDataMap?.schemaMap ?? {},
        primaryConstraints?.tableConstraints ?? {},
        foreignConstraints?.tableConstraints ?? {}
      ),
    [isSchemaMapValidating, isPkValidating, isFkValidating]
  );

  if (isJobLoading || isSchemaDataMapLoading) {
    return (
      <div className="space-y-10">
        <Skeleton className="w-full h-12" />
        <Skeleton className="w-1/2 h-12" />
        <SkeletonTable />
      </div>
    );
  }

  async function onSubmit(values: SingleTableSchemaFormValues) {
    const job = data?.job;
    if (!job) {
      return;
    }
    try {
      await updateJobConnection(account?.id ?? '', job, values);
      toast({
        title: 'Successfully updated job source connection!',
        variant: 'success',
      });
      mutate();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update job source connection',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  const formValues = form.watch();
  const schemaTableData = formValues.mappings?.map((mapping) => ({
    ...mapping,
    schema: formValues.schema,
    table: formValues.table,
  }));

  const [uniqueSchemas, schemaTableMap] = getUniqueSchemasAndTables(
    connectionSchemaDataMap?.schemaMap ?? {}
  );

  const selectedSchemaTables = schemaTableMap.get(formValues.schema) ?? [];

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="schema"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Schema</FormLabel>
              <FormDescription>The name of the schema.</FormDescription>
              <Select
                onValueChange={(value: string) => {
                  if (!value) {
                    return;
                  }
                  field.onChange(value);
                  form.setValue('table', ''); // reset the table value because it may no longer apply
                }}
                value={field.value}
              >
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a schema..." />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {Array.from(uniqueSchemas).map((schema) => (
                    <SelectItem
                      className="cursor-pointer"
                      key={schema}
                      value={schema}
                    >
                      {schema}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="table"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Table Name</FormLabel>
              <FormDescription>The name of the table.</FormDescription>
              <Select
                disabled={!formValues.schema}
                onValueChange={(value: string) => {
                  if (!value) {
                    return;
                  }
                  field.onChange(value);
                  form.setValue(
                    'mappings',
                    allJobMappings
                      .filter(
                        (m) =>
                          m.schema === formValues.schema && m.table === value
                      )
                      .map((r) => {
                        return {
                          schema: formValues.schema,
                          table: value,
                          column: r.column,
                          dataType: r.dataType,
                          transformer: r.transformer,
                        };
                      })
                  );
                }}
                value={field.value}
              >
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a table..." />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {selectedSchemaTables.map((table) => (
                    <SelectItem
                      className="cursor-pointer"
                      key={table}
                      value={table}
                    >
                      {table}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="numRows"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Number of Rows</FormLabel>
              <FormDescription>The number of rows to generate.</FormDescription>
              <FormControl>
                <Input
                  type="text"
                  value={field.value
                    .toString()
                    .replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                  onChange={(e) => {
                    const numberValue = parseFloat(
                      e.target.value.replace(/,/g, '')
                    );
                    if (!isNaN(numberValue)) {
                      field.onChange(numberValue);
                    } else {
                      field.onChange(0);
                    }
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {formValues.schema && formValues.table && (
          <SchemaTable
            data={schemaTableData}
            excludeInputReqTransformers
            jobType="generate"
            constraintHandler={schemaConstraintHandler}
            schema={connectionSchemaDataMap?.schemaMap ?? {}}
          />
        )}
        {form.formState.errors.mappings && (
          <Alert variant="destructive">
            <AlertTitle className="flex flex-row space-x-2 justify-center">
              <ExclamationTriangleIcon />
              <p>Please fix form errors and try again.</p>
            </AlertTitle>
          </Alert>
        )}
        <div className="flex flex-row gap-1 justify-end">
          <Button key="submit" type="submit">
            Submit
          </Button>
        </div>
      </form>
    </Form>
  );
}

function getJobSource(job?: Job): SingleTableSchemaFormValues {
  if (!job) {
    return {
      mappings: [],
      numRows: 0,
      schema: '',
      table: '',
    };
  }
  let schema = '';
  let table = '';
  let numRows = 0;
  if (job.source?.options?.config.case === 'generate') {
    const srcSchemas = job.source.options.config.value.schemas;
    if (srcSchemas.length > 0) {
      schema = srcSchemas[0].schema;
      const tables = srcSchemas[0].tables;
      if (tables.length > 0) {
        table = tables[0].table;
        numRows = Number(tables[0].rowCount); // this will be an issue if the number is bigger than what js allows
      }
    }
  }

  const mappings: SingleTableSchemaFormValues['mappings'] = (
    job.mappings ?? []
  ).map((mapping) => {
    return {
      schema: mapping.schema,
      table: mapping.table,
      column: mapping.column,
      transformer: mapping.transformer
        ? convertJobMappingTransformerToForm(mapping.transformer)
        : convertJobMappingTransformerToForm(new JobMappingTransformer()),
    };
  });
  return {
    mappings: mappings,
    numRows: numRows,
    schema: schema,
    table: table,
  };
}

async function updateJobConnection(
  accountId: string,
  job: Job,
  values: SingleTableSchemaFormValues
): Promise<UpdateJobSourceConnectionResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${job.id}/source-connection`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new UpdateJobSourceConnectionRequest({
          id: job.id,
          mappings: values.mappings.map((m) => {
            return new JobMapping({
              schema: values.schema,
              table: values.table,
              column: m.column,
              transformer:
                convertJobMappingTransformerFormToJobMappingTransformer(
                  m.transformer
                ),
            });
          }),
          source: new JobSource({
            options: new JobSourceOptions({
              config: {
                case: 'generate',
                value: new GenerateSourceOptions({
                  fkSourceConnectionId: getFkIdFromGenerateSource(job.source),
                  schemas: [
                    new GenerateSourceSchemaOption({
                      schema: values.schema,
                      tables: [
                        new GenerateSourceTableOption({
                          table: values.table,
                          rowCount: BigInt(values.numRows),
                        }),
                      ],
                    }),
                  ],
                }),
              },
            }),
          }),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobSourceConnectionResponse.fromJson(await res.json());
}

export function getUniqueSchemasAndTables(
  // key: schema.table
  schemaMap: Record<string, PlainMessage<DatabaseColumn>[]>
): [Set<string>, Map<string, string[]>] {
  const uniqueSchemas = new Set<string>();
  const tableToSchemaMap = new Map<string, string[]>();

  // Can be sneaky here because the record is expected to be keyed by the table.
  // So the values become a list of columns an we can short circuit and only care about the first record to get the
  // objectified schema and table, which is easier than splitting the key
  Object.values(schemaMap).forEach((dbcols) => {
    if (dbcols.length === 0) {
      return;
    }
    const [dbcol] = dbcols;
    uniqueSchemas.add(dbcol.schema);
    const tables = tableToSchemaMap.get(dbcol.schema) ?? [];
    tableToSchemaMap.set(dbcol.schema, [...tables, dbcol.table]);
  });
  return [uniqueSchemas, tableToSchemaMap];
}
