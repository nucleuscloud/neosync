'use client';
import SwitchCard from '@/components/switches/SwitchCard';
import { useToast } from '@/components/ui/use-toast';
import {
  Job,
  PauseJobRequest,
  PauseJobResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { ReactElement } from 'react';
import * as Yup from 'yup';

const FORM_SCHEMA = Yup.object({
  note: Yup.string().optional(),
  isPaused: Yup.boolean().required(),
  error: Yup.string().optional(),
});

export type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

interface Props {
  job: Job;
  mutate: () => void;
}

export default function JobPauseSwitch({ job, mutate }: Props): ReactElement {
  const { toast } = useToast();

  async function onClick(isPaused: boolean) {
    try {
      await pauseJob(job.id, isPaused);
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
    <SwitchCard
      isChecked={job.pauseStatus?.isPaused || false}
      onCheckedChange={async (value) => {
        onClick(value);
      }}
      title="Pause job"
      description="Prevents future job runs."
    />
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
