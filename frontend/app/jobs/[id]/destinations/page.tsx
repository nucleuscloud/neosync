'use client';
import ButtonText from '@/components/ButtonText';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { PlusIcon } from '@radix-ui/react-icons';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import DestinationConnectionCard from './components/DestinationConnectionCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const account = useAccount();
  const { data, isLoading, mutate } = useGetJob(id);
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections(account?.id ?? '');

  const connections = connectionsData?.connections ?? [];
  const destinationIds = data?.job?.destinations.map((d) => d.connectionId);
  return (
    <div className="job-details-container">
      <SubPageHeader
        header="Destination Connections"
        description={`Manage job's destination connections`}
        extraHeading={<NewDestinationButton jobId={id} />}
      />

      {isLoading || isConnectionsLoading ? (
        <Skeleton className="w-full h-48 rounded-lg" />
      ) : (
        <div className="space-y-10">
          {data?.job?.destinations.map((destination) => {
            return (
              <DestinationConnectionCard
                key={destination.id}
                jobId={id}
                destination={destination}
                mutate={mutate}
                connections={connections}
                availableConnections={connections.filter(
                  (c) =>
                    c.id == destination.connectionId ||
                    (c.id != data?.job?.source?.connectionId &&
                      !destinationIds?.includes(c.id))
                )}
                isDeleteDisabled={data?.job?.destinations.length == 1}
              />
            );
          })}
        </div>
      )}
    </div>
  );
}

interface NewDestinationButtonProps {
  jobId: string;
}

function NewDestinationButton(props: NewDestinationButtonProps): ReactElement {
  const { jobId } = props;
  return (
    <NextLink href={`/new/job/${jobId}/destination`}>
      <Button>
        <ButtonText leftIcon={<PlusIcon />} text="New Destination" />
      </Button>
    </NextLink>
  );
}
