'use client';
import SourceOptionsForm from '@/components/jobs/Form/SourceOptionsForm';
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
  JobSource,
  JobSourceOptions,
  SqlSourceConnectionOptions,
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { SOURCE_FORM_SCHEMA } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import { getJob } from '../../util';

interface Props {
  jobId: string;
}

const FORM_SCHEMA = SOURCE_FORM_SCHEMA.concat(
  Yup.object({
    destinationId: Yup.string().required(),
  })
);
export type SourceFormValues = Yup.InferType<typeof FORM_SCHEMA>;

export default function SourceConnectionCard({ jobId }: Props): ReactElement {
  const { toast } = useToast();
  const {
    isLoading: isConnectionsLoading,
    data: connectionsData,
    mutate,
  } = useGetConnections();

  const connections = connectionsData?.connections ?? [];

  const form = useForm({
    resolver: yupResolver<SourceFormValues>(FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getJob(jobId);
      if (!res) {
        return {
          sourceId: '',
          sourceOptions: {
            haltOnNewColumnAddition: false,
          },
          destinationId: '',
        };
      }
      const destinationIds = res.job?.destinations.map((d) => d.connectionId);
      const values = {
        sourceId: res.job?.source?.connectionId || '',
        sourceOptions: {},
        destinationId: destinationIds ? destinationIds[0] : '',
      };
      switch (res.job?.source?.options?.config.case) {
        case 'sqlOptions':
          return {
            ...values,
            sourceOptions: {
              haltOnNewColumnAddition:
                res.job?.source?.options?.config.value.haltOnNewColumnAddition,
            },
          };
        default:
          return values;
      }
    },
  });

  async function onSubmit(values: SourceFormValues) {
    try {
      await updateJobConnection(jobId, values);
      mutate();
      toast({
        title: 'Successfully updated job source connection!',
        variant: 'default',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update job source connection',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
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
                  {isConnectionsLoading ? (
                    <Skeleton />
                  ) : (
                    <Select onValueChange={field.onChange} value={field.value}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {connections
                          .filter(
                            (c) => c.id !== form.getValues().destinationId
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
                  )}
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
