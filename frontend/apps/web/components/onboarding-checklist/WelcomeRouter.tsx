import { NewJobType } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { cn } from '@/libs/utils';
import { MagicWandIcon, SymbolIcon } from '@radix-ui/react-icons';
import { nanoid } from 'nanoid';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { AiOutlineExperiment } from 'react-icons/ai';
import { PageProps } from '../types';
import { Button } from '../ui/button';

interface Props {
  currentStep: number;
  setCurrentStep: (val: number) => void;
  setIsDialogOpen: (val: boolean) => void;
}

export default function WelcomeRouter({
  params,
  currentStep,
  setIsDialogOpen,
  setCurrentStep,
}: Props & PageProps): ReactElement {
  const [sessionToken, setSessionToken] = useState<string>('');
  const searchParams = useSearchParams();
  const { account } = useAccount();

  useEffect(() => {
    // Generate the session token only on the client side
    setSessionToken(params?.sessionId ?? nanoid());
  }, []);

  const dataSyncParams = new URLSearchParams(searchParams);
  dataSyncParams.set('jobType', 'data-sync');
  if (!dataSyncParams.has('sessionId')) {
    dataSyncParams.set('sessionId', sessionToken);
  }

  const dataGenParams = new URLSearchParams(searchParams);
  dataGenParams.set('jobType', 'generate-table');
  if (!dataGenParams.has('sessionId')) {
    dataGenParams.set('sessionId', sessionToken);
  }

  const aiDataGenParams = new URLSearchParams(searchParams);
  aiDataGenParams.set('jobType', 'ai-generate-table');
  if (!aiDataGenParams.has('sessionId')) {
    aiDataGenParams.set('sessionId', sessionToken);
  }

  const jobData = [
    {
      name: 'Data Synchronization',
      description: 'Sync and anonymize data between a source and destination. ',
      href: `/${account?.name}/new/job/define?${dataSyncParams.toString()}`,
      icon: <SymbolIcon />,
      type: 'data-sync',
      experimental: false,
    },
    {
      name: 'Data Generation',
      description: 'Generate synthetic data from using Neosync Transformers.',
      href: `/${account?.name}/new/job/define?${dataGenParams.toString()}`,
      icon: <AiOutlineExperiment />,
      type: 'generate-table',
      experimental: false,
    },
    {
      name: 'AI Data Generation',
      description: 'Generate synthetic data from scratch with AI.',
      href: `/${account?.name}/new/job/define?${aiDataGenParams.toString()}`,
      icon: <MagicWandIcon />,
      type: 'ai-generate-table',
      experimental: true,
    },
  ] as const;

  const [selectedJobType, setSelectedJobType] =
    useState<NewJobType>('data-sync');

  const [href, setHref] = useState<string>();

  const handleJobSelection = (jobType: NewJobType, href: string) => {
    setSelectedJobType(jobType);
    setHref(href);
  };

  const router = useRouter();
  const posthog = usePostHog();

  return (
    <div className="flex flex-col gap-8 justify-center items-center text-center ">
      <h1 className="font-semibold text-2xl">Select a Job type</h1>
      <p className="text-sm px-10">
        Select the type of job that you would like to create.
      </p>
      <div className="flex flex-col justify-center">
        <RadioGroup value={selectedJobType} onChange={() => setSelectedJobType}>
          {jobData.map((jd) => (
            <Card
              key={jd.name}
              className={cn(
                'cursor-pointer p-2',
                selectedJobType === jd.type
                  ? 'border border-black shadow-sm dark:border-gray-500'
                  : 'hover:border hover:border-gray-500 dark:border-gray-700 dark:hover:border-gray-600'
              )}
              onClick={() => handleJobSelection(jd.type, jd.href)}
            >
              <CardHeader>
                <div className="flex flex-col md:flex-row justify-between items-center">
                  <div>
                    <CardTitle>
                      <div className="flex flex-row items-center gap-2">
                        <div>{jd.icon}</div>
                        <p>{jd.name}</p>
                        {jd.experimental ? <Badge>Experimental</Badge> : null}
                      </div>
                    </CardTitle>
                    <CardDescription className="pl-6 pt-2">
                      {jd.description}
                    </CardDescription>
                  </div>
                  <RadioGroupItem
                    value={jd.type}
                    id={jd.type}
                    className={`${selectedJobType === jd.type ? 'bg-black text-white' : 'bg-white dark:bg-transparent text-black'}`}
                  />
                </div>
              </CardHeader>
            </Card>
          ))}
        </RadioGroup>
      </div>
      <div className="flex flex-row justify-between w-full py-6">
        <Button
          variant="outline"
          type="reset"
          onClick={() => setCurrentStep(currentStep - 1)}
        >
          Back
        </Button>
        <Button
          type="submit"
          disabled={!selectedJobType}
          onClick={() => {
            setIsDialogOpen(false);
            setCurrentStep(currentStep - 1);
            router.push(
              href ??
                `/${account?.name}/new/job/define?${dataSyncParams.toString()}`
            );
            posthog.capture('New Job Flow Started', {
              jobType: selectedJobType,
            });
          }}
        >
          Next
        </Button>
      </div>
    </div>
  );
}
