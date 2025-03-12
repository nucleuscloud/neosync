'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { cn } from '@/libs/utils';
import { MagicWandIcon, SymbolIcon } from '@radix-ui/react-icons';
import { nanoid } from 'nanoid';
import { useTheme } from 'next-themes';
import Image from 'next/image';
import {
  ReadonlyURLSearchParams,
  useRouter,
  useSearchParams,
} from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, ReactNode, use, useEffect, useState } from 'react';
import { AiOutlineExperiment } from 'react-icons/ai';
import { NewJobType } from './job-form-validations';

export default function NewJob(props: PageProps): ReactElement {
  const params = use(props.params);
  const searchParams = useSearchParams();
  const { account } = useAccount();

  const jobData = useGetJobData(params, searchParams);

  const [selectedJobType, setSelectedJobType] =
    useState<NewJobType>('data-sync');

  const [href, setHref] = useState<string | undefined>();

  const handleJobSelection = (jobType: NewJobType, href: string) => {
    setSelectedJobType(jobType);
    setHref(href);
  };

  const router = useRouter();
  const posthog = usePostHog();
  const theme = useTheme();

  return (
    <div
      id="newjobdefine"
      className="px-12 sm:px-24 md:px-48 lg:px-60 flex flex-col pt-4 gap-16"
    >
      <OverviewContainer Header={<PageHeader header="Select a Job type" />}>
        <RadioGroup
          value={selectedJobType}
          className="flex flex-col lg:flex-row flex-wrap justify-center items-center gap-6 pt-8"
        >
          {jobData.map((jd) => (
            <Card
              key={jd.name}
              className={cn(
                'cursor-pointer',
                selectedJobType === jd.type
                  ? 'border border-black shadow-md dark:border-gray-400'
                  : 'hover:border hover:border-gray-500 dark:border-gray-700 dark:hover:border-gray-600'
              )}
              onClick={() => handleJobSelection(jd.type, jd.href)}
            >
              <CardHeader className="w-[300px] relative">
                <div className="flex flex-col items-center text-left">
                  <div className="relative">
                    <Image
                      src={jd.image}
                      alt="image"
                      width="200"
                      height="200"
                    />
                    <div
                      className="absolute inset-0"
                      style={{
                        background:
                          theme.resolvedTheme == 'light'
                            ? 'linear-gradient(to top, rgba(255,255,255,1) 0%, rgba(255,255,255,1) 5%, rgba(255,255,255,0) 100%)'
                            : 'linear-gradient(to top, rgba(30,30,36,1) 0%, rgba(30,30,36,1) 5%, rgba(30,30,36,0) 100%)',
                        pointerEvents: 'none',
                      }}
                    />
                  </div>
                  <div className="pt-8">
                    <CardTitle>
                      <div className="flex flex-row items-center gap-2 text-nowrap">
                        <p>{jd.name}</p>
                        <div>{jd.icon}</div>
                        {jd.experimental ? <Badge>Experimental</Badge> : null}
                      </div>
                    </CardTitle>
                    <CardDescription className="pt-2">
                      {jd.description}
                    </CardDescription>
                  </div>
                </div>
                <div className="absolute top-0 right-2">
                  <RadioGroupItem
                    value={jd.type}
                    id={jd.type}
                    className={`${selectedJobType === jd.type ? 'bg-black dark:bg-white text-white dark:text-gray-900' : 'bg-white dark:bg-transparent text-black'}`}
                  />{' '}
                </div>
              </CardHeader>
            </Card>
          ))}
        </RadioGroup>
      </OverviewContainer>
      <div className="flex flex-col md:flex-row justify-between gap-1">
        <Button
          variant="outline"
          type="reset"
          onClick={() => router.push(`/${account?.name}`)}
        >
          Back
        </Button>
        <Button
          type="submit"
          disabled={!selectedJobType || !href}
          onClick={() => {
            if (!href) {
              return;
            }
            router.push(href);
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

interface JobData {
  name: string;
  description: string;
  href: string;
  icon: ReactNode;
  type: NewJobType;
  experimental: boolean;
  image: string;
}

function useGetJobData(
  params: Record<string, string>,
  searchParams: ReadonlyURLSearchParams
): JobData[] {
  const [sessionToken, setSessionToken] = useState<string>('');
  const { account } = useAccount();
  const { data: systemAppConfig } = useGetSystemAppConfig();
  const { resolvedTheme } = useTheme();

  useEffect(() => {
    // Generate the session token only on the client side
    setSessionToken(params?.sessionId ?? nanoid());
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

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

  const piiDetectionParams = new URLSearchParams(searchParams);
  piiDetectionParams.set('jobType', 'pii-detection');
  if (!piiDetectionParams.has('sessionId')) {
    piiDetectionParams.set('sessionId', sessionToken);
  }
  const jobData: JobData[] = [
    {
      name: 'Data Synchronization',
      description:
        'Synchronize and anonymize data between a source and destination. ',
      href: `/${account?.name}/new/job/define?${dataSyncParams.toString()}`,
      icon: <SymbolIcon />,
      type: 'data-sync',
      experimental: false,
      image:
        resolvedTheme === 'light'
          ? 'https://assets.nucleuscloud.com/neosync/app/jobsynclight.svg'
          : 'https://assets.nucleuscloud.com/neosync/app/prodsync-dark.svg',
    },
    {
      name: 'Data Generation',
      description:
        'Generate synthetic data from scratch for a chosen destination.',
      href: `/${account?.name}/new/job/define?${dataGenParams.toString()}`,
      icon: <AiOutlineExperiment />,
      type: 'generate-table',
      experimental: false,
      image:
        resolvedTheme === 'light'
          ? 'https://assets.nucleuscloud.com/neosync/app/gen-light.svg'
          : 'https://assets.nucleuscloud.com/neosync/app/datagen-dark.svg',
    },
    {
      name: 'AI Data Generation',
      description: 'Generate synthetic data from scratch with AI.',
      href: `/${account?.name}/new/job/define?${aiDataGenParams.toString()}`,
      icon: <MagicWandIcon />,
      type: 'ai-generate-table',
      experimental: true,
      image:
        resolvedTheme === 'light'
          ? 'https://assets.nucleuscloud.com/neosync/app/aigen.svg'
          : 'https://assets.nucleuscloud.com/neosync/app/aigen-dark.svg',
    },
  ];

  if (systemAppConfig?.isPiiDetectionJobEnabled) {
    jobData.push({
      name: 'PII Detection',
      description:
        'Scan your database for PII and sensitive data to identify security risks.',
      href: `/${account?.name}/new/job/define?${piiDetectionParams.toString()}`,
      icon: <MagicWandIcon />,
      type: 'pii-detection',
      experimental: true,
      image:
        resolvedTheme === 'light'
          ? 'https://assets.nucleuscloud.com/neosync/app/jobsynclight.svg'
          : 'https://assets.nucleuscloud.com/neosync/app/prodsync-dark.svg',
    });
  }

  return jobData;
}
