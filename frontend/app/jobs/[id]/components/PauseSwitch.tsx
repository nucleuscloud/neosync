'use client';
import SwitchCard from '@/components/switches/SwitchCard';
import { useToast } from '@/components/ui/use-toast';
import {
  JobStatus,
  PauseJobRequest,
  PauseJobResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { ReactElement } from 'react';

interface Props {
  jobId: string;
  status?: JobStatus;
  mutate: () => void;
}

export default function JobPauseSwitch({
  status,
  mutate,
  jobId,
}: Props): ReactElement {
  const { toast } = useToast();

  async function onClick(isPaused: boolean) {
    try {
      await pauseJob(jobId, isPaused);
      toast({
        title: `Successfully ${isPaused ? 'paused' : 'unpaused'}  job!`,
        variant: 'default',
      });
      mutate();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to pause',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <div className="max-w-[300px]">
      <SwitchCard
        isChecked={status == JobStatus.PAUSED || false}
        onCheckedChange={async (value) => {
          onClick(value);
        }}
        title="Pause job"
        description="Prevents future job runs."
      />
    </div>
  );
}

async function pauseJob(
  jobId: string,
  isPaused: boolean
): Promise<PauseJobResponse> {
  const res = await fetch(`/api/jobs/${jobId}/pause`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new PauseJobRequest({
        id: jobId,
        pause: isPaused,
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return PauseJobResponse.fromJson(await res.json());
}
