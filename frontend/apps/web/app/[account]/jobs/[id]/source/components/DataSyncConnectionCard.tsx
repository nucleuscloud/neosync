'use client';
import SourceOptionsForm from '@/components/jobs/Form/SourceOptionsForm';
import {
  SchemaTable,
  getConnectionSchema,
} from '@/components/jobs/SchemaTable/schema-table';
import { useAccount } from '@/components/providers/account-provider';
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
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { getErrorMessage } from '@/util/util';
import {
  SCHEMA_FORM_SCHEMA,
  SOURCE_FORM_SCHEMA,
  TransformerFormValues,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  DatabaseColumn,
  Job,
  JobMapping,
  JobMappingTransformer,
  JobSource,
  JobSourceOptions,
  MysqlSourceConnectionOptions,
  Passthrough,
  PostgresSourceConnectionOptions,
  TransformerConfig,
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
} from '@neosync/sdk';
import { ReactElement, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import { getConnection } from '../../util';

interface Props {
  jobId: string;
}

const FORM_SCHEMA = SOURCE_FORM_SCHEMA.concat(
  Yup.object({
    destinationIds: Yup.array().of(Yup.string().required()),
  })
).concat(SCHEMA_FORM_SCHEMA);
type SourceFormValues = Yup.InferType<typeof FORM_SCHEMA>;
export interface SchemaMap {
  [schema: string]: {
    [table: string]: {
      [column: string]: {
        transformer: JobMappingTransformer;
      };
    };
  };
}

function getConnectionIdFromSource(
  js: JobSource | undefined
): string | undefined {
  if (
    js?.options?.config.case === 'postgres' ||
    js?.options?.config.case === 'mysql' ||
    js?.options?.config.case === 'awsS3'
  ) {
    return js.options.config.value.connectionId;
  }
  return undefined;
}

export default function DataSyncConnectionCard({ jobId }: Props): ReactElement {
  const { toast } = useToast();
  const { account } = useAccount();
  const { data, mutate } = useGetJob(account?.id ?? '', jobId);
  const sourceConnectionId = getConnectionIdFromSource(data?.job?.source);
  const { data: schema, error } = useGetConnectionSchema(
    account?.id ?? '',
    sourceConnectionId
  );
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections(account?.id ?? '');

  const connections = connectionsData?.connections ?? [];

  useEffect(() => {
    if (error) {
      toast({
        title: 'Unable to get connection schema',
        description: getErrorMessage(error),
        variant: 'destructive',
      });
    }
  }, [error]);

  const form = useForm({
    resolver: yupResolver<SourceFormValues>(FORM_SCHEMA),
    defaultValues: {
      sourceId: '',
      sourceOptions: {
        haltOnNewColumnAddition: false,
      },
      destinationIds: [],
      mappings: [],
    },
    values: getJobSource(data?.job, schema?.schemas),
  });

  async function onSourceChange(value: string): Promise<void> {
    try {
      const newValues = await getUpdatedValues(
        account?.id ?? '',
        value,
        form.getValues()
      );
      form.reset(newValues);
    } catch (err) {
      form.reset({ ...form.getValues, mappings: [], sourceId: value });
      toast({
        title: 'Unable to get connection schema',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  async function onSubmit(values: SourceFormValues) {
    const connection = connections.find((c) => (c.id = values.sourceId));
    const job = data?.job;
    if (!job || !connection) {
      return;
    }
    try {
      await updateJobConnection(account?.id ?? '', job, values, connection);
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

  if (isConnectionsLoading) {
    return (
      <div className="space-y-10">
        <Skeleton className="w-full h-12" />
        <Skeleton className="w-1/2 h-12" />
        <SkeletonTable />
      </div>
    );
  }

  const source = connections.find((item) => item.id == sourceConnectionId);

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <div className="space-y-8">
          <FormField
            control={form.control}
            name="sourceId"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Source</FormLabel>
                <FormDescription>
                  The location of the source data set.
                </FormDescription>
                <FormControl>
                  <Select
                    value={field.value}
                    onValueChange={async (value) => {
                      if (!value) {
                        return;
                      }
                      field.onChange(value);
                      await onSourceChange(value);
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={source?.name} />
                    </SelectTrigger>
                    <SelectContent>
                      {connections
                        .filter(
                          (c) =>
                            !form.getValues().destinationIds?.includes(c.id) &&
                            c.connectionConfig?.config.case != 'awsS3Config'
                        )
                        .map((connection) => (
                          <SelectItem
                            className="cursor-pointer"
                            key={connection.id}
                            value={connection.id}
                          >
                            {connection.name}
                          </SelectItem>
                        ))}
                    </SelectContent>
                  </Select>
                </FormControl>

                <FormMessage />
              </FormItem>
            )}
          />
          <SourceOptionsForm
            connection={connections.find(
              (c) => c.id == form.getValues().sourceId
            )}
            maxColNum={2}
          />

          <SchemaTable data={form.getValues().mappings} />
          <div className="flex flex-row items-center justify-end w-full mt-4">
            <Button type="submit">Save</Button>
          </div>
        </div>
      </form>
    </Form>
  );
}

async function updateJobConnection(
  accountId: string,
  job: Job,
  values: SourceFormValues,
  connection: Connection
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
            const jmt = new JobMappingTransformer({
              source: m.transformer.source,
              config: m.transformer.config as TransformerConfig,
            });

            return new JobMapping({
              schema: m.schema,
              table: m.table,
              column: m.column,
              transformer: jmt,
            });
          }),
          source: new JobSource({
            options: toJobSourceOptions(
              values,
              job,
              connection,
              values.sourceId
            ),
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

function toJobSourceOptions(
  values: SourceFormValues,
  job: Job,
  connection: Connection,
  newSourceId: string
): JobSourceOptions {
  switch (connection.connectionConfig?.config.case) {
    case 'pgConfig':
      return new JobSourceOptions({
        config: {
          case: 'postgres',
          value: new PostgresSourceConnectionOptions({
            ...getExistingPostgresSourceConnectionOptions(job),
            connectionId: newSourceId,
            haltOnNewColumnAddition:
              values.sourceOptions.haltOnNewColumnAddition,
          }),
        },
      });
    case 'mysqlConfig':
      return new JobSourceOptions({
        config: {
          case: 'mysql',
          value: new MysqlSourceConnectionOptions({
            ...getExistingMysqlSourceConnectionOptions(job),
            connectionId: newSourceId,
            haltOnNewColumnAddition:
              values.sourceOptions.haltOnNewColumnAddition,
          }),
        },
      });
    default:
      throw new Error('unsupported connection type');
  }
}

function getExistingPostgresSourceConnectionOptions(
  job: Job
): PostgresSourceConnectionOptions | undefined {
  return job.source?.options?.config.case === 'postgres'
    ? job.source.options.config.value
    : undefined;
}

function getExistingMysqlSourceConnectionOptions(
  job: Job
): MysqlSourceConnectionOptions | undefined {
  return job.source?.options?.config.case === 'mysql'
    ? job.source.options.config.value
    : undefined;
}

function getJobSource(job?: Job, schema?: DatabaseColumn[]): SourceFormValues {
  if (!job || !schema) {
    return {
      sourceId: '',
      sourceOptions: {
        haltOnNewColumnAddition: false,
      },
      destinationIds: [],
      mappings: [],
    };
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

  const mappings = schema.map((c) => {
    const colMapping = getColumnMapping(schemaMap, c.schema, c.table, c.column);

    return {
      schema: c.schema,
      table: c.table,
      column: c.column,
      dataType: c.dataType,
      transformer:
        colMapping?.transformer ??
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
  });
  const destinationIds = job?.destinations.map((d) => d.connectionId);
  const values = {
    sourceOptions: {},
    destinationIds: destinationIds,
    mappings: mappings || [],
  };

  const yupValidationValues = {
    ...values,
    sourceId: getConnectionIdFromSource(job.source) || '',
    mappings: values.mappings.map((mapping) => ({
      ...mapping,
      transformer: mapping.transformer as TransformerFormValues,
    })),
  };

  switch (job?.source?.options?.config.case) {
    case 'postgres':
      return {
        ...yupValidationValues,
        sourceId: getConnectionIdFromSource(job.source) || '',
        sourceOptions: {
          haltOnNewColumnAddition:
            job?.source?.options?.config.value.haltOnNewColumnAddition,
        },
      };
    case 'mysql':
      return {
        ...yupValidationValues,
        sourceId: getConnectionIdFromSource(job.source) || '',
        sourceOptions: {
          haltOnNewColumnAddition:
            job?.source?.options?.config.value.haltOnNewColumnAddition,
        },
      };
    default:
      return yupValidationValues;
  }
}

export function getColumnMapping(
  schemaMap: SchemaMap,
  schema: string,
  table: string,
  column: string
): { transformer: JobMappingTransformer } | undefined {
  if (!schemaMap[schema]) {
    return;
  }
  if (!schemaMap[schema][table]) {
    return;
  }

  return schemaMap[schema][table][column];
}

async function getUpdatedValues(
  accountId: string,
  connectionId: string,
  originalValues: SourceFormValues
): Promise<SourceFormValues> {
  const [schemaRes, connRes] = await Promise.all([
    getConnectionSchema(accountId, connectionId),
    getConnection(accountId, connectionId),
  ]);

  if (!schemaRes || !connRes) {
    return originalValues;
  }

  const mappings = schemaRes.schemas.map((r) => {
    return {
      ...r,
      transformer: new JobMappingTransformer({}),
    };
  });

  const values = {
    sourceId: connectionId || '',
    sourceOptions: {},
    destinationIds: originalValues.destinationIds,
    mappings: mappings || [],
  };

  const yupValidationValues = {
    ...values,
    mappings: values.mappings.map((mapping) => ({
      ...mapping,
      transformer: mapping.transformer as {
        source: string;
        config: { config: { case: string; value: {} } };
      },
    })),
  };

  switch (connRes.connection?.connectionConfig?.config.case) {
    case 'pgConfig':
      return {
        ...yupValidationValues,
        sourceOptions: {
          haltOnNewColumnAddition: false,
        },
      };
    default:
      return yupValidationValues;
  }
}
