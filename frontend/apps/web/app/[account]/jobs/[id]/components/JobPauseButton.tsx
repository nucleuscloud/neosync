'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { JobStatus, PauseJobRequest, PauseJobResponse } from '@neosync/sdk';
import { PauseIcon, PlayIcon } from '@radix-ui/react-icons';
import { ReactElement, useEffect, useState } from 'react';

interface Props {
  jobId: string;
  status?: JobStatus;
  onNewStatus(status: JobStatus): void;
}

export default function JobPauseButton({
  status,
  onNewStatus,
  jobId,
}: Props): ReactElement {
  const { account } = useAccount();
  const { toast } = useToast();
  const [buttonText, setButtonText] = useState(
    status === JobStatus.PAUSED ? 'Resume Job' : 'Pause Job'
  );
  const [buttonIcon, setButtonIcon] = useState<JSX.Element>(
    status === JobStatus.PAUSED ? <PlayIcon /> : <PauseIcon />
  );
  const [isTrying, setIsTrying] = useState<boolean>(false);

  useEffect(() => {
    setButtonText(status === JobStatus.PAUSED ? 'Resume Job' : 'Pause Job');
    if (isTrying) {
      setButtonIcon(<Spinner />);
    } else {
      setButtonIcon(status === JobStatus.PAUSED ? <PlayIcon /> : <PauseIcon />);
    }
  }, [status, isTrying]);

  async function updateJobStatus(isPaused: boolean): Promise<void> {
    if (isTrying) {
      return;
    }
    try {
      setIsTrying(true);
      await pauseJob(account?.id ?? '', jobId, isPaused);
      toast({
        title: `Successfully ${isPaused ? 'paused' : 'unpaused'}  job!`,
        variant: 'success',
      });
      onNewStatus(isPaused ? JobStatus.PAUSED : JobStatus.ENABLED);
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to pause',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    } finally {
      setIsTrying(false);
    }
  }

  return (
    <div className="max-w-[300px]">
      <Button
        variant="outline"
        onClick={async () => {
          const isCurrentlyPaused = status === JobStatus.PAUSED;
          updateJobStatus(!isCurrentlyPaused);
        }}
      >
        <ButtonText leftIcon={buttonIcon} text={buttonText} />
      </Button>
    </div>
  );
}

async function pauseJob(
  accountId: string,
  jobId: string,
  isPaused: boolean
): Promise<PauseJobResponse> {
  const res = await fetch(`/api/accounts/${accountId}/jobs/${jobId}/pause`, {
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
