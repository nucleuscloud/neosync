'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { SymbolIcon } from '@radix-ui/react-icons';
import { nanoid } from 'nanoid';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
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
    },
    {
      name: 'Data Generation',
      description:
        'Generate synthetic data from scratch for a chosen destination.',
      href: `/${account?.name}/new/job/define?${dataGenParams.toString()}`,
      icon: <AiOutlineExperiment />,
    },
  ] as const;

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Create a new job"
          description="Select a job type to synchronize or generate data"
          pageHeaderContainerClassName="mx-24"
        />
      }
    >
      <div className="gap-6 rounded-lg grid lg:grid-cols-2 xl:grid-cols-3 content-stretch mx-24">
        {jobData.map((jd) => (
          <JobCard
            key={jd.name}
            name={jd.name}
            description={jd.description}
            href={jd.href}
            icon={jd.icon}
          />
        ))}
      </div>
    </OverviewContainer>
  );
}

interface JobCardProps {
  name: string;
  description: string;
  href: string;
  icon: JSX.Element;
}

function JobCard(props: JobCardProps): ReactElement {
  const { name, description, href, icon } = props;
  return (
    <Link href={href}>
      <Card className="cursor-pointer hover:border hover:border-gray-500 min-h-[110px]">
        <CardHeader>
          <CardTitle>
            <div className="flex flex-row items-center gap-2">
              <div>{icon}</div>
              <p>{name}</p>
            </div>
          </CardTitle>
          <CardDescription className="pl-6">{description}</CardDescription>
        </CardHeader>
      </Card>
    </Link>
  );
}
