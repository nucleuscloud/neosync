'use client';
import {
  SINGLE_TABLE_SCHEMA_FORM_SCHEMA,
  SingleTableSchemaFormValues,
} from '@/app/new/job/schema';
import { SchemaTable } from '@/components/jobs/SchemaTable/schema-table';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
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
import { useGetConnectionSchema } from '@/libs/hooks/useGetConnectionSchema';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { DatabaseColumn } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
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
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import {
  Passthrough,
  TransformerConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { getErrorMessage } from '@/util/util';
import { TransformerFormValues } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { SchemaMap, getColumnMapping } from './DataSyncConnectionCard';
import { getFkIdFromGenerateSource } from './util';

interface Props {
  jobId: string;
}

export default function DataGenConnectionCard({ jobId }: Props): ReactElement {
  const { toast } = useToast();

  const { data, mutate, isLoading: isJobLoading } = useGetJob(jobId);
  const fkSourceConnectionId = getFkIdFromGenerateSource(data?.job?.source);
  const { data: schema, isLoading: isGetConnectionsSchemaLoading } =
    useGetConnectionSchema(fkSourceConnectionId);

  const form = useForm<SingleTableSchemaFormValues>({
    resolver: yupResolver(SINGLE_TABLE_SCHEMA_FORM_SCHEMA),
    defaultValues: {
      mappings: [],
      numRows: 0,
      schema: 'foo',
      table: 'bar',
    },
    values: getJobSource(data?.job, schema?.schemas),
  });

  if (isJobLoading || isGetConnectionsSchemaLoading) {
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
      await updateJobConnection(job, values);
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

  // changed to getValues() from watch() because watch would listen for a change and then trigger a render which would overwrite the values that were changed (which we wanted to keep)

  const formValues = form.getValues();
  const schemaTableData = formValues.mappings?.map((mapping) => ({
    ...mapping,
    schema: formValues.schema,
    table: formValues.table,
  }));

  const uniqueSchemas = Array.from(
    new Set(schema?.schemas.map((s) => s.schema))
  );
  const schemaTableMap = getSchemaTableMap(schema?.schemas ?? []);

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
                  {uniqueSchemas.map((schema) => (
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
                <Input value={field.value} onChange={field.onChange} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {formValues.schema && formValues.table && (
          <SchemaTable data={schemaTableData} excludeInputReqTransformers />
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

function getJobSource(
  job?: Job,
  dbCols?: DatabaseColumn[]
): SingleTableSchemaFormValues {
  if (!job || !dbCols) {
    return {
      mappings: [],
      numRows: 0,
      schema: 'foo2',
      table: 'bar2',
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

  const schemaMap: SchemaMap = {};
  job?.mappings.forEach((c) => {
    if (!schemaMap[c.schema]) {
      schemaMap[c.schema] = {
        [c.table]: {
          [c.column]: {
            transformer:
              c?.transformer ??
              new JobMappingTransformer({
                source: 'passthrough',
                config: new TransformerConfig({
                  config: {
                    case: 'passthroughConfig',
                    value: new Passthrough({}),
                  },
                }),
              }),
          },
        },
      };
    } else if (!schemaMap[c.schema][c.table]) {
      schemaMap[c.schema][c.table] = {
        [c.column]: {
          transformer:
            c?.transformer ??
            new JobMappingTransformer({
              source: 'passthrough',
              config: new TransformerConfig({
                config: {
                  case: 'passthroughConfig',
                  value: new Passthrough({}),
                },
              }),
            }),
        },
      };
    } else {
      schemaMap[c.schema][c.table][c.column] = {
        transformer:
          c.transformer ??
          new JobMappingTransformer({
            source: 'passthrough',
            config: new TransformerConfig({
              config: {
                case: 'passthroughConfig',
                value: new Passthrough({}),
              },
            }),
          }),
      };
    }
  });

  const mappings: SingleTableSchemaFormValues['mappings'] = dbCols.map((c) => {
    const colMapping = getColumnMapping(schemaMap, c.schema, c.table, c.column);
    const transformer =
      colMapping?.transformer ??
      new JobMappingTransformer({
        source: 'passthrough',
        config: new TransformerConfig({
          config: {
            case: 'passthroughConfig',
            value: new Passthrough({}),
          },
        }),
      });

    return {
      column: c.column,
      dataType: c.dataType,
      transformer: transformer as TransformerFormValues,
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
  job: Job,
  values: SingleTableSchemaFormValues
): Promise<UpdateJobSourceConnectionResponse> {
  const res = await fetch(`/api/jobs/${job.id}/source-connection`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new UpdateJobSourceConnectionRequest({
        id: job.id,
        mappings: values.mappings.map((m) => {
          const jmt = new JobMappingTransformer({
            source: m.transformer.source,
            config: m.transformer.config as TransformerConfig,
          });

          return new JobMapping({
            schema: values.schema,
            table: values.table,
            column: m.column,
            transformer: jmt,
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
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobSourceConnectionResponse.fromJson(await res.json());
}

function getSchemaTableMap(schemas: DatabaseColumn[]): Map<string, string[]> {
  const map = new Map<string, Set<string>>();
  schemas.forEach((schema) => {
    const set = map.get(schema.schema);
    if (set) {
      set.add(schema.table);
    } else {
      map.set(schema.schema, new Set([schema.table]));
    }
  });

  const outMap = new Map<string, string[]>();
  map.forEach((tableSet, schema) => outMap.set(schema, Array.from(tableSet)));
  return outMap;
}
