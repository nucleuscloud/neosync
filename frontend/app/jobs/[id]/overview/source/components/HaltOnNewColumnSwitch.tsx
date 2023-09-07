'use client';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import {
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import { getJob } from '../../util';

const CONNECTIONS_FORM_SCHEMA = Yup.object({
  sourceId: Yup.string().uuid().required(),
  destinationId: Yup.string().uuid().required(),
});

export type ConnectionsFormValues = Yup.InferType<
  typeof CONNECTIONS_FORM_SCHEMA
>;

interface Props {
  jobId: string;
}

export default function HaltOnNewColumnSwitch({ jobId }: Props): ReactElement {
  const { toast } = useToast();
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections();

  const connections = connectionsData?.connections ?? [];

  const form = useForm({
    resolver: yupResolver<ConnectionsFormValues>(CONNECTIONS_FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getJob(jobId);
      if (!res) {
        return { sourceId: '', destinationId: '' };
      }
      return {
        sourceId: res.job?.connectionSourceId || '',
        destinationId: res.job?.connectionDestinationIds[0] || '',
      };
    },
  });

  async function onSubmit(values: ConnectionsFormValues) {
    try {
      await updateJobConnection(jobId, values.sourceId);
      toast({
        title: 'Successfully updated job schedule!',
        variant: 'default',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update job schedule',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <div className="w-96">
      <div className="flex flex-row items-center justify-between rounded-lg border p-4">
        <div className="space-y-0.5">
          <Label className="text-base">Halt Job on new column addition</Label>
          <p className="text-sm text-muted-foreground">
            Stops job runs if new column is detected
          </p>
        </div>
        <Switch
          checked={data?.job?.sourceOptions?.haltOnNewColumnAddition}
          onCheckedChange={() => {}}
        />
      </div>
    </div>
  );
}

async function updateJobConnection(
  jobId: string,
  connectionId: string
): Promise<UpdateJobSourceConnectionResponse> {
  const res = await fetch(`/api/jobs/${jobId}/schedule`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new UpdateJobSourceConnectionRequest({
        id: jobId,
        connectionId: connectionId,
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobSourceConnectionResponse.fromJson(await res.json());
}
