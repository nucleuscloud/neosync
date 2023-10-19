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
import { DatabaseColumn } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  Job,
  JobMapping,
  JobSource,
  JobSourceOptions,
  SqlSourceConnectionOptions,
  Transformer,
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { SCHEMA_FORM_SCHEMA, SOURCE_FORM_SCHEMA } from '@/yup-validations/jobs';
import { toTransformerConfigOptions } from '@/yup-validations/transformers';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
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
interface SchemaMap {
  [schema: string]: {
    [table: string]: {
      [column: string]: {
        transformer: Transformer;
      };
    };
  };
}

export default function SourceConnectionCard({ jobId }: Props): ReactElement {
  const { toast } = useToast();
  const account = useAccount();
  const { data, mutate } = useGetJob(jobId);
  const { data: schema } = useGetConnectionSchema(
    data?.job?.source?.connectionId
  );
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections(account?.id ?? '');

  const connections = connectionsData?.connections ?? [];

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
      const newValues = await getUpdatedValues(value, form.getValues());
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
    const job = data?.job;
    if (!job) {
      return;
    }
    try {
      await updateJobConnection(job, values);
      toast({
        title: 'Successfully updated job source connection!',
        variant: 'default',
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
                <FormControl>
                  <Select
                    value={field.value}
                    onValueChange={async (value) => {
                      field.onChange(value);
                      await onSourceChange(value);
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue />
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
                <FormDescription>
                  The location of the source data set.
                </FormDescription>
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

          <SchemaTable data={form.getValues().mappings || []} />
          <div className="flex flex-row items-center justify-end w-full mt-4">
            <Button type="submit">Save</Button>
          </div>
        </div>
      </form>
    </Form>
  );
}

async function updateJobConnection(
  job: Job,
  values: SourceFormValues
): Promise<UpdateJobSourceConnectionResponse> {
  const existingSourceOpts = getExistingSqlSourceConnectionOptions(job);
  const res = await fetch(`/api/jobs/${job.id}/source-connection`, {
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
            transformer: toTransformerConfigOptions(m.transformer),
          });
        }),
        source: new JobSource({
          connectionId: values.sourceId,
          options: new JobSourceOptions({
            config: {
              case: 'sqlOptions',
              value: new SqlSourceConnectionOptions({
                ...existingSourceOpts,
                haltOnNewColumnAddition:
                  values.sourceOptions.haltOnNewColumnAddition,
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

function getExistingSqlSourceConnectionOptions(
  job: Job
): SqlSourceConnectionOptions | undefined {
  return job.source?.options?.config.case === 'sqlOptions'
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
            transformer: c.transformer ?? new Transformer(),
          },
        },
      };
    } else if (!schemaMap[c.schema][c.table]) {
      schemaMap[c.schema][c.table] = {
        [c.column]: {
          transformer:
            c.transformer ??
            new Transformer({
              value: 'passthrough',
            }),
        },
      };
    } else {
      schemaMap[c.schema][c.table][c.column] = {
        transformer: c.transformer ?? new Transformer({ value: 'passthrough' }),
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
        colMapping?.transformer ?? new Transformer({ value: 'passthrough' }),
    };
  });

  const destinationIds = job?.destinations.map((d) => d.connectionId);
  const values = {
    sourceId: job?.source?.connectionId || '',
    sourceOptions: {},
    destinationIds: destinationIds,
    mappings: mappings || [],
  };

  //update to map the tranformer values from proto defintion to the yup validation definition
  const yupValidationValues = {
    ...values,
    mappings: values.mappings.map((mapping) => ({
      ...mapping,
      transformer: { value: mapping.transformer.value, config: {} },
    })),
  };

  switch (job?.source?.options?.config.case) {
    case 'sqlOptions':
      return {
        ...yupValidationValues,
        sourceOptions: {
          haltOnNewColumnAddition:
            job?.source?.options?.config.value.haltOnNewColumnAddition,
        },
      };
    default:
      return yupValidationValues;
  }
}

function getColumnMapping(
  schemaMap: SchemaMap,
  schema: string,
  table: string,
  column: string
): { transformer: Transformer } | undefined {
  if (!schemaMap[schema]) {
    return;
  }
  if (!schemaMap[schema][table]) {
    return;
  }

  return schemaMap[schema][table][column];
}

async function getUpdatedValues(
  connectionId: string,
  originalValues: SourceFormValues
): Promise<SourceFormValues> {
  const [schemaRes, connRes] = await Promise.all([
    getConnectionSchema(connectionId),
    getConnection(connectionId),
  ]);

  if (!schemaRes || !connRes) {
    return originalValues;
  }

  const mappings = schemaRes.schemas.map((r) => {
    return {
      ...r,
      transformer: '',
    };
  });

  const values = {
    sourceId: connectionId || '',
    sourceOptions: {},
    destinationIds: originalValues.destinationIds,
    mappings: mappings || [],
  };

  //update to map the tranformer values from proto defintion to the yup validation definition
  const yupValidationValues = {
    ...values,
    mappings: values.mappings.map((mapping) => ({
      ...mapping,
      transformer: { value: mapping.transformer, config: {} },
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
