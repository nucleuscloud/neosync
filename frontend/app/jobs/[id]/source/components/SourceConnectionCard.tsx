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
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import {
  JobMapping,
  JobSource,
  JobSourceOptions,
  SqlSourceConnectionOptions,
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { SCHEMA_FORM_SCHEMA, SOURCE_FORM_SCHEMA } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import { getConnection, getJob } from '../../util';

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
        dataType: string;
      };
    };
  };
}

export default function SourceConnectionCard({ jobId }: Props): ReactElement {
  const { toast } = useToast();
  const account = useAccount();
  const {
    isLoading: isConnectionsLoading,
    data: connectionsData,
    mutate,
  } = useGetConnections(account?.id ?? '');

  const connections = connectionsData?.connections ?? [];

  const form = useForm({
    resolver: yupResolver<SourceFormValues>(FORM_SCHEMA),
    defaultValues: async () => {
      try {
        const res = await getJobSource(jobId);
        return res;
      } catch (err) {
        console.error(err);
        toast({
          title: 'Unable to get connection schema',
          description: getErrorMessage(err),
          variant: 'destructive',
        });
        return {
          sourceId: '',
          sourceOptions: {
            haltOnNewColumnAddition: false,
          },
          destinationIds: [],
          mappings: [],
        };
      }
    },
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
    try {
      await updateJobConnection(jobId, values);
      toast({
        title: 'Successfully updated job source connection!',
        variant: 'default',
      });
      mutate();
      form.reset();
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

          <SchemaTable data={form.getValues().mappings} />
          <div className="flex flex-row items-center justify-end w-full mt-4">
            <Button disabled={!form.formState.isDirty} type="submit">
              Save
            </Button>
          </div>
        </div>
      </form>
    </Form>
  );
}

async function updateJobConnection(
  jobId: string,
  values: SourceFormValues
): Promise<UpdateJobSourceConnectionResponse> {
  const res = await fetch(`/api/jobs/${jobId}/source-connection`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new UpdateJobSourceConnectionRequest({
        id: jobId,
        mappings: values.mappings.map((m) => {
          return new JobMapping({
            schema: m.schema,
            table: m.table,
            column: m.column,
            transformer: m.transformer,
            exclude: m.exclude,
          });
        }),
        source: new JobSource({
          connectionId: values.sourceId,
          options: new JobSourceOptions({
            config: {
              case: 'sqlOptions',
              value: new SqlSourceConnectionOptions({
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

async function getJobSource(jobId?: string): Promise<SourceFormValues> {
  const emptyValues = {
    sourceId: '',
    sourceOptions: {
      haltOnNewColumnAddition: false,
    },
    destinationIds: [],
    mappings: [],
  };
  if (!jobId) {
    return emptyValues;
  }
  try {
    const jobRes = await getJob(jobId);
    if (!jobRes) {
      return emptyValues;
    }
    const job = jobRes?.job;

    const res = await getConnectionSchema(job?.source?.connectionId);
    if (!res) {
      return emptyValues;
    }

    const schemaMap: SchemaMap = {};
    res.schemas.forEach((c) => {
      if (!schemaMap[c.schema]) {
        schemaMap[c.schema] = {
          [c.table]: {
            [c.column]: {
              dataType: c.dataType,
            },
          },
        };
      } else if (!schemaMap[c.schema][c.table]) {
        schemaMap[c.schema][c.table] = {
          [c.column]: {
            dataType: c.dataType,
          },
        };
      } else {
        schemaMap[c.schema][c.table][c.column] = { dataType: c.dataType };
      }
    });

    const mappings = job?.mappings.map((r) => {
      const datatype = schemaMap[r.schema][r.table][r.column].dataType;
      return {
        ...r,
        transformer: r.transformer as unknown as string,
        dataType: datatype || '',
      };
    });
    const destinationIds = jobRes.job?.destinations.map((d) => d.connectionId);
    const values = {
      sourceId: jobRes.job?.source?.connectionId || '',
      sourceOptions: {},
      destinationIds: destinationIds,
      mappings: mappings || [],
    };
    switch (jobRes.job?.source?.options?.config.case) {
      case 'sqlOptions':
        return {
          ...values,
          sourceOptions: {
            haltOnNewColumnAddition:
              jobRes.job?.source?.options?.config.value.haltOnNewColumnAddition,
          },
        };
      default:
        return values;
    }
  } catch (err) {
    console.error(err);
    throw new Error(getErrorMessage(err));
  }
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
