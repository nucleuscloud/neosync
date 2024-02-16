'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { SymbolIcon } from '@radix-ui/react-icons';
import { nanoid } from 'nanoid';
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { AiOutlineExperiment } from 'react-icons/ai';

export type NewJobType = 'data-sync' | 'generate-table';

export default function NewJob({ params }: PageProps): ReactElement {
  const [sessionToken, setSessionToken] = useState<string>('');
  const searchParams = useSearchParams();
  const { account } = useAccount();

  useEffect(() => {
    // Generate the session token only on the client side
    setSessionToken(params?.sessionToken ?? nanoid());
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

  const jobData = [
    {
      name: 'Data Synchronization',
      description:
        'Synchronize and anonymize data between a source and destination. ',
      href: `/${account?.name}/new/job/define?${dataSyncParams.toString()}`,
      icon: <SymbolIcon />,
      type: 'data-sync',
    },
    {
      name: 'Data Generation',
      description:
        'Generate synthetic data from scratch for a chosen destination.',
      href: `/${account?.name}/new/job/define?${dataGenParams.toString()}`,
      icon: <AiOutlineExperiment />,
      type: 'generate-table',
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

  return (
    <div
      id="newjobdefine"
      className="px-12 md:px-48 lg:px-96 flex flex-col pt-4 gap-16"
    >
      <OverviewContainer Header={<PageHeader header="Select a Job type" />}>
        <div className="flex flex-col justify-center gap-6 pt-8">
          <RadioGroup
            value={selectedJobType}
            onChange={() => setSelectedJobType}
          >
            {jobData.map((jd) => (
              <div key={jd.name}>
                <Card
                  className={`cursor-pointer p-2 w-full min-w-[400px] ${selectedJobType === jd.type ? 'border border-black shadow-sm' : 'hover:border hover:border-gray-500'}`}
                  onClick={() => handleJobSelection(jd.type, jd.href)}
                >
                  <CardHeader>
                    <div className="flex flex-row justify-between items-center h-full">
                      <div>
                        <CardTitle>
                          <div className="flex flex-row items-center gap-2 ">
                            <div>{jd.icon}</div>
                            <p>{jd.name}</p>
                          </div>
                        </CardTitle>
                        <CardDescription className="pl-6 pt-2">
                          {jd.description}
                        </CardDescription>
                      </div>
                      <RadioGroupItem
                        value={jd.type}
                        id={jd.type}
                        className={`${selectedJobType === jd.type ? 'bg-black text-white' : 'bg-white text-black'}`}
                      />
                    </div>
                  </CardHeader>
                </Card>
              </div>
            ))}
          </RadioGroup>
        </div>
      </OverviewContainer>
      <div className="flex flex-row justify-between">
        <Button
          variant="outline"
          type="reset"
          onClick={() => router.push(`/${account?.name}`)}
        >
          Back
        </Button>
        <Button
          type="submit"
          disabled={!selectedJobType}
          onClick={() =>
            router.push(
              href ??
                `/${account?.name}/new/job/define?${dataSyncParams.toString()}`
            )
          }
        >
          Next
        </Button>
      </div>
    </div>
  );
}
