'use client';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { ReactElement } from 'react';
import DestinationConnectionCard from './components/DestinationConnectionCard';
import SourceConnectionCard from './components/SourceConnectionCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';

  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections();

  const connections = connectionsData?.connections ?? [];

  const form = useForm({
    resolver: yupResolver<ConnectionsFormValues>(CONNECTIONS_FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getJob(id);
      if (!res) {
        return { sourceId: '', destinationId: '' };
      }
      return {
        sourceId: res.job?.connectionSourceId || '',
        destinationId: res.job?.connectionDestinationIds[0] || '',
      };
    },
  });

  async function onSubmit(_values: ConnectionsFormValues) {}

  return (
    <div className="job-details-container">
      <PageHeader header="Connections" description="Manage job connections" />

      <div className="space-y-10">
        <SourceConnectionCard jobId={id} />
        <DestinationConnectionCard jobId={id} />
      </div>
    </div>
  );
}
