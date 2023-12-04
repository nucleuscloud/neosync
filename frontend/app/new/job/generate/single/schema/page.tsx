'use client';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import {
  SchemaTable,
  getConnectionSchema,
} from '@/components/jobs/SchemaTable/schema-table';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
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
import { useToast } from '@/components/ui/use-toast';
import { useGetConnectionSchema } from '@/libs/hooks/useGetConnectionSchema';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import {
  Connection,
  DatabaseColumn,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  CreateJobRequest,
  CreateJobResponse,
  GenerateSourceOptions,
  GenerateSourceSchemaOption,
  GenerateSourceTableOption,
  JobDestination,
  JobMapping,
  JobMappingTransformer,
  JobSource,
  JobSourceOptions,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import {
  Passthrough,
  TransformerConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { getErrorMessage } from '@/util/util';
import { toJobDestinationOptions } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, { DATA_GEN_STEPS } from '../../../JobsProgressSteps';
import {
  DefineFormValues,
  SINGLE_TABLE_SCHEMA_FORM_SCHEMA,
  SingleTableConnectFormValues,
  SingleTableSchemaFormValues,
} from '../../../schema';
const isBrowser = () => typeof window !== 'undefined';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const { toast } = useToast();

  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/new/job`);
    }
  }, [searchParams?.sessionId]);
  const { data: connectionsData } = useGetConnections(account?.id ?? '');
  const connections = connectionsData?.connections ?? [];

  const sessionPrefix = searchParams?.sessionId ?? '';

  // Used to complete the whole form
  const defineFormKey = `${sessionPrefix}-new-job-define`;
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    defineFormKey,
    { jobName: '' }
  );

  const connectFormKey = `${sessionPrefix}-new-job-single-table-connect`;
  const [connectFormValues] = useSessionStorage<SingleTableConnectFormValues>(
    connectFormKey,
    {
      destination: { connectionId: '', destinationOptions: {} },
    }
  );
  const { data: connSchemaData } = useGetConnectionSchema(
    connectFormValues.destination.connectionId
  );

  const formKey = `${sessionPrefix}-new-job-single-table-schema`;

  const [defaultValues] = useSessionStorage<SingleTableSchemaFormValues>(
    formKey,
    {
      mappings: [],
      numRows: 10,
      schema: '',
      table: '',
    }
  );

  async function getSchema(): Promise<SingleTableSchemaFormValues> {
    try {
      const res = await getConnectionSchema(
        connectFormValues.destination.connectionId
      );
      if (!res) {
        return defaultValues;
      }

      return {
        ...defaultValues,
        mappings: res.schemas.map((r) => {
          var pt = new JobMappingTransformer({
            source: 'passthrough',
            name: 'passthrough',
            config: new TransformerConfig({
              config: {
                case: 'passthroughConfig',
                value: new Passthrough({}),
              },
            }),
          }) as {
            source: string;
            name: string;
            config: {
              config: {
                case?: string;
                value: {};
              };
            };
          };
          return {
            ...r,
            transformer: pt,
          };
        }),
      };
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to get connection schema',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
      return defaultValues;
    }
  }

  const form = useForm({
    resolver: yupResolver<SingleTableSchemaFormValues>(
      SINGLE_TABLE_SCHEMA_FORM_SCHEMA
    ),
    defaultValues: async () => {
      return getSchema();
    },
  });

  useFormPersist(formKey, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });

  async function onSubmit(values: SingleTableSchemaFormValues) {
    if (!account) {
      return;
    }
    try {
      const job = await createNewJob(
        defineFormValues,
        connectFormValues,
        values,
        account.id,
        connections
      );
      toast({
        title: 'Successfully created job!',
        variant: 'success',
      });
      window.sessionStorage.removeItem(defineFormKey);
      window.sessionStorage.removeItem(connectFormKey);
      window.sessionStorage.removeItem(formKey);
      if (job.job?.id) {
        router.push(`/jobs/${job.job.id}`);
      } else {
        router.push(`/jobs`);
      }
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to create job',
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

  const uniqueSchemas = Array.from(
    new Set(connSchemaData?.schemas.map((s) => s.schema))
  );
  const schemaTableMap = getSchemaTableMap(connSchemaData?.schemas ?? []);

  const selectedSchemaTables = schemaTableMap.get(formValues.schema) ?? [];

  /* turning the input field into a controlled component due to console warning a component going from uncontrolled to controlled. The input at first receives an undefined value because the async getSchema() call hasn't returned yet then once it returns it sets the value which throws the error. Ideally, react hook form should just handle this but for some reason it's throwing an error. Revist this in the future.
   */

  const [rowNum, setRowNum] = useState<number>(
    form.getValues('numRows')
      ? form.getValues('numRows')
      : defaultValues.numRows
  );
  useEffect(() => {
    form.setValue(`numRows`, rowNum);
  }, [rowNum]);

  return (
    <div className="flex flex-col gap-20">
      <OverviewContainer
        Header={
          <PageHeader
            header="Schema"
            progressSteps={
              <JobsProgressSteps steps={DATA_GEN_STEPS} stepName={'schema'} />
            }
          />
        }
        containerClassName="connect-page"
      >
        <div />
      </OverviewContainer>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name="schema"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Schema</FormLabel>
                <FormDescription>The name of the schema.</FormDescription>
                <FormControl>
                  <Select
                    onValueChange={(value: string) => {
                      field.onChange(value);
                      form.setValue('table', ''); // reset the table value because it may no longer apply
                    }}
                    value={field.value}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a schema..." />
                    </SelectTrigger>
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
                </FormControl>
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
                <FormControl>
                  <Select
                    disabled={!formValues.schema}
                    onValueChange={field.onChange}
                    value={field.value}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a table..." />
                    </SelectTrigger>
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
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <Controller
            control={form.control}
            name="numRows"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Number of Rows</FormLabel>
                <FormDescription>
                  The number of rows to generate.
                </FormDescription>
                <FormControl>
                  <Input
                    value={rowNum}
                    onChange={(e) => {
                      setRowNum(Number(e.target.value));
                    }}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          {formValues.schema && formValues.table && (
            <SchemaTable data={schemaTableData} excludeTransformers />
          )}
          <div className="flex flex-row gap-1 justify-between">
            <Button key="back" type="button" onClick={() => router.back()}>
              Back
            </Button>
            <Button key="submit" type="submit">
              Submit
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}

async function createNewJob(
  define: DefineFormValues,
  connect: SingleTableConnectFormValues,
  schema: SingleTableSchemaFormValues,
  accountId: string,
  connections: Connection[]
): Promise<CreateJobResponse> {
  const connectionIdMap = new Map(
    connections.map((connection) => [connection.id, connection])
  );
  const body = new CreateJobRequest({
    accountId,
    jobName: define.jobName,
    cronSchedule: define.cronSchedule,
    initiateJobRun: define.initiateJobRun,
    mappings: schema.mappings.map((m) => {
      const jmt = new JobMappingTransformer({
        name: m.transformer.name,
        source: m.transformer.source,
        config: new TransformerConfig({
          config: {
            case: m.transformer.config.config.case as any,
            value: m.transformer.config.config.value,
          },
        }),
      });
      return new JobMapping({
        schema: schema.schema,
        table: schema.table,
        column: m.column,
        transformer: jmt,
      });
    }),
    source: new JobSource({
      options: new JobSourceOptions({
        config: {
          case: 'generate',
          value: new GenerateSourceOptions({
            fkSourceConnectionId: connect.destination.connectionId,
            schemas: [
              new GenerateSourceSchemaOption({
                schema: schema.schema,
                tables: [
                  new GenerateSourceTableOption({
                    rowCount: BigInt(schema.numRows),
                    table: schema.table,
                  }),
                ],
              }),
            ],
          }),
        },
      }),
    }),
    destinations: [
      new JobDestination({
        connectionId: connect.destination.connectionId,
        options: toJobDestinationOptions(
          connect.destination,
          connectionIdMap.get(connect.destination.connectionId)
        ),
      }),
    ],
  });

  const res = await fetch(`/api/jobs`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateJobResponse.fromJson(await res.json());
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
