'use client';
import ButtonText from '@/components/ButtonText';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { Job, UserAccount } from '@neosync/sdk';
import { nanoid } from 'nanoid';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { GrClone } from 'react-icons/gr';
import { NewJobType } from '../../../new/job/job-form-validations';
import { setDefaultNewJobFormValues } from '../../util';

interface Props {
  job: Job;
}

export default function JobCloneButton(props: Props): ReactElement {
  const { job } = props;
  const { account } = useAccount();
  const router = useRouter();

  function onCloneClick(): void {
    if (!account) {
      return;
    }
    const sessionId = nanoid();
    setDefaultNewJobFormValues(window.sessionStorage, job, sessionId);
    router.push(getJobCloneUrlFromJob(account, job, sessionId));
  }

  return (
    <Button variant="outline" onClick={onCloneClick}>
      <ButtonText text="Clone Job" leftIcon={<GrClone />} />
    </Button>
  );
}

export function getJobCloneUrlFromJob(
  account: UserAccount,
  job: Job,
  sessionId: string
): string {
  const urlParams = new URLSearchParams({
    jobType: getNewJobTypeFromJob(job),
    sessionId: sessionId,
  });
  return `/${account.name}/new/job/define?${urlParams.toString()}`;
}

function getNewJobTypeFromJob(job: Job): NewJobType {
  if (job.source?.options?.config.case === 'aiGenerate') {
    return 'ai-generate-table';
  }
  if (job.source?.options?.config.case === 'generate') {
    return 'generate-table';
  }
  if (job.jobType?.jobType.case === 'piiDetect') {
    return 'pii-detection';
  }
  return 'data-sync';
}
