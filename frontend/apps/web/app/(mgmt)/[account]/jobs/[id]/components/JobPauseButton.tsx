'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import { Button } from '@/components/ui/button';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { JobService, JobStatus } from '@neosync/sdk';
import { PauseIcon, PlayIcon } from '@radix-ui/react-icons';
import { ReactElement, useEffect, useState } from 'react';
import { toast } from 'sonner';

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
  const { mutateAsync: setJobPaused } = useMutation(JobService.method.pauseJob);
  const [buttonText, setButtonText] = useState(
    status === JobStatus.PAUSED ? 'Resume Job' : 'Pause Job'
  );
  const [buttonIcon, setButtonIcon] = useState<ReactElement>(
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
      await setJobPaused({
        id: jobId,
        pause: isPaused,
      });
      toast.success(`Successfully ${isPaused ? 'paused' : 'unpaused'}  job!`);
      onNewStatus(isPaused ? JobStatus.PAUSED : JobStatus.ENABLED);
    } catch (err) {
      console.error(err);
      toast.error('Unable to update job paused status', {
        description: getErrorMessage(err),
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
