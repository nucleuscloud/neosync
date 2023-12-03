'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { nanoid } from 'nanoid';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { ReactElement, useState } from 'react';

export type NewJobType = 'data-sync' | 'generate-table';

export default function NewJob({ params }: PageProps): ReactElement {
  const [sessionToken] = useState(params?.sessionToken ?? nanoid());
  const searchParams = useSearchParams();

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
      description: 'Synchronize data between two different data sources',
      href: `/new/job/define?${dataSyncParams.toString()}`,
    },
    {
      name: 'Single Table Data Generation',
      description: 'Generate data for a single table in a chosen data source',
      href: `/new/job/define?${dataGenParams.toString()}`,
    },
  ] as const;

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Create a new job"
          description="Create a new job that can be triggered or scheduled in the system"
          pageHeaderContainerClassName="mx-24"
        />
      }
    >
      <div className="gap-6 rounded-lg md:grid lg:grid-cols-2 xl:grid-cols-3 content-stretch mx-24">
        {jobData.map((jd) => (
          <JobCard
            key={jd.name}
            name={jd.name}
            description={jd.description}
            href={jd.href}
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
}

function JobCard(props: JobCardProps): ReactElement {
  const { name, description, href } = props;
  return (
    <Link href={href}>
      <Card className="cursor-pointer">
        <CardHeader>
          <CardTitle>
            <div className="flex flex-row items-center gap-2">
              <p>{name}</p>
            </div>
          </CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
      </Card>
    </Link>
  );
}
