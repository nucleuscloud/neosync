'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import {
  JobStatus,
  PauseJobRequest,
  PauseJobResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { PauseIcon, PlayIcon } from '@radix-ui/react-icons';
import { ReactElement, useEffect, useState } from 'react';

interface Props {
  jobId: string;
  status?: JobStatus;
  mutate: () => void;
}

export default function JobPauseButton({
  status,
  mutate,
  jobId,
}: Props): ReactElement {
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
    setButtonIcon(status === JobStatus.PAUSED ? <PlayIcon /> : <PauseIcon />);
  }, [status]);

  async function updateJobStatus(isPaused: boolean) {
    setIsTrying(true);
    try {
      await pauseJob(jobId, isPaused);
      toast({
        title: `Successfully ${isPaused ? 'paused' : 'unpaused'}  job!`,
        variant: 'default',
      });
      mutate();
      setIsTrying(false);
      setButtonText((val) => (val == 'Pause Job' ? 'Resume Job' : 'Pause Job'));
      setButtonIcon(handleIcon());
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to pause',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
      setIsTrying(false);
    }
  }

  const handleIcon = () => {
    if (isTrying) {
      return <Spinner />;
    } else if (!isTrying && buttonText == 'Resume Job') {
      return <PlayIcon />;
    } else {
      return <PauseIcon />;
    }
  };

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
