'use client';
import SourceOptionsForm from '@/components/jobs/Form/SourceOptionsForm';
import { SchemaTable } from '@/components/jobs/SchemaTable/SchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
import { useAccount } from '@/components/providers/account-provider';
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
import { useToast } from '@/components/ui/use-toast';
import { useGetConnectionForeignConstraints } from '@/libs/hooks/useGetConnectionForeignConstraints';
import { useGetConnectionPrimaryConstraints } from '@/libs/hooks/useGetConnectionPrimaryConstraints';
import {
  GetConnectionSchemaMapResponse,
  getConnectionSchema,
  useGetConnectionSchemaMap,
} from '@/libs/hooks/useGetConnectionSchemaMap';
import { useGetConnectionUniqueConstraints } from '@/libs/hooks/useGetConnectionUniqueConstraints';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { getErrorMessage } from '@/util/util';
import {
  SCHEMA_FORM_SCHEMA,
  SOURCE_FORM_SCHEMA,
  convertJobMappingTransformerFormToJobMappingTransformer,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  Job,
  JobMapping,
  JobMappingTransformer,
  JobSource,
  JobSourceOptions,
  MysqlSourceConnectionOptions,
  PostgresSourceConnectionOptions,
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
} from '@neosync/sdk';
import { ReactElement, useMemo } from 'react';
import { useForm } from 'react-hook-form';
import { KeyedMutator } from 'swr';
import * as Yup from 'yup';
import { getConnection } from '../../util';
import SchemaPageSkeleton from './SchemaPageSkeleton';

interface Props {
  jobId: string;
}

const FORM_SCHEMA = SOURCE_FORM_SCHEMA.concat(
  Yup.object({
    destinationIds: Yup.array().of(Yup.string().required()),
  })
).concat(SCHEMA_FORM_SCHEMA);
type SourceFormValues = Yup.InferType<typeof FORM_SCHEMA>;

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
  const {
    data,
    mutate,
    isLoading: isJobDataLoading,
  } = useGetJob(account?.id ?? '', jobId);
  const sourceConnectionId = getConnectionIdFromSource(data?.job?.source);

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaDataMapLoading,
    isValidating: isSchemaMapValidating,
    mutate: mutateGetConnectionSchemaMap,
  } = useGetConnectionSchemaMap(account?.id ?? '', sourceConnectionId ?? '');

  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections(account?.id ?? '');

  const connections = connectionsData?.connections ?? [];

  const form = useForm({
    resolver: yupResolver<SourceFormValues>(FORM_SCHEMA),
    values: getJobSource(data?.job),
  });

  const { data: primaryConstraints, isValidating: isPkValidating } =
    useGetConnectionPrimaryConstraints(
      account?.id ?? '',
      sourceConnectionId ?? ''
    );

  const { data: foreignConstraints, isValidating: isFkValidating } =
    useGetConnectionForeignConstraints(
      account?.id ?? '',
      sourceConnectionId ?? ''
    );

  const { data: uniqueConstraints, isValidating: isUCValidating } =
    useGetConnectionUniqueConstraints(
      account?.id ?? '',
      sourceConnectionId ?? ''
    );

  const schemaConstraintHandler = useMemo(
    () =>
      getSchemaConstraintHandler(
        connectionSchemaDataMap?.schemaMap ?? {},
        primaryConstraints?.tableConstraints ?? {},
        foreignConstraints?.tableConstraints ?? {},
        uniqueConstraints?.tableConstraints ?? {}
      ),
    [isSchemaMapValidating, isPkValidating, isFkValidating, isUCValidating]
  );

  async function onSourceChange(value: string): Promise<void> {
    try {
      const newValues = await getUpdatedValues(
        account?.id ?? '',
        value,
        form.getValues(),
        mutateGetConnectionSchemaMap
      );
      form.reset(newValues);
      mutateGetConnectionSchemaMap();
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
    const connection = connections.find((c) => c.id === values.sourceId);
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

  if (isConnectionsLoading || isSchemaDataMapLoading || isJobDataLoading) {
    return <SchemaPageSkeleton />;
  }

  const source = connections.find((item) => item.id === sourceConnectionId);
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
                            className="cursor-pointer ml-2"
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
              (c) => c.id === form.getValues().sourceId
            )}
            maxColNum={2}
          />

          <SchemaTable
            data={form.watch('mappings')}
            jobType="sync"
            constraintHandler={schemaConstraintHandler}
            schema={connectionSchemaDataMap?.schemaMap ?? {}}
            isSchemaDataReloading={isSchemaMapValidating}
          />
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
            return new JobMapping({
              schema: m.schema,
              table: m.table,
              column: m.column,
              transformer:
                convertJobMappingTransformerFormToJobMappingTransformer(
                  m.transformer
                ),
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

function getJobSource(job?: Job): SourceFormValues {
  if (!job) {
    return {
      sourceId: '',
      sourceOptions: {
        haltOnNewColumnAddition: false,
      },
      destinationIds: [],
      mappings: [],
      connectionId: '',
    };
  }

  const mappings = (job.mappings ?? []).map((mapping) => {
    return {
      ...mapping,
      transformer: mapping.transformer
        ? convertJobMappingTransformerToForm(mapping.transformer)
        : convertJobMappingTransformerToForm(new JobMappingTransformer()),
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
    mappings,
    connectionId: getConnectionIdFromSource(job.source) || '',
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

async function getUpdatedValues(
  accountId: string,
  connectionId: string,
  originalValues: SourceFormValues,
  mutateConnectionSchemaRes:
    | KeyedMutator<unknown>
    | KeyedMutator<GetConnectionSchemaMapResponse>
): Promise<SourceFormValues> {
  const [schemaRes, connRes] = await Promise.all([
    getConnectionSchema(accountId, connectionId),
    getConnection(accountId, connectionId),
  ]);

  if (!schemaRes || !connRes) {
    return originalValues;
  }

  const sameKeys = new Set(
    Object.values(schemaRes.schemaMap).flatMap((dbcols) =>
      dbcols.map((dbcol) => `${dbcol.schema}.${dbcol.table}.${dbcol.column}`)
    )
  );

  const mappings = originalValues.mappings.filter((mapping) =>
    sameKeys.has(`${mapping.schema}.${mapping.table}.${mapping.column}`)
  );

  const values = {
    sourceId: connectionId || '',
    sourceOptions: {},
    destinationIds: originalValues.destinationIds,
    mappings,
    connectionId: connectionId || '',
  };
  mutateConnectionSchemaRes(schemaRes);

  switch (connRes.connection?.connectionConfig?.config.case) {
    case 'pgConfig':
      return {
        ...values,
        sourceOptions: {
          haltOnNewColumnAddition: false,
        },
      };
    default:
      return values;
  }
}
