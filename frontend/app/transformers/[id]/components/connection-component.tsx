'use client';
import ConnectionIcon from '@/components/connections/ConnectionIcon';
import PageHeader from '@/components/headers/PageHeader';
import {
  Connection,
  UpdateConnectionResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { ReactElement } from 'react';
import PostgresForm from './PostgresForm';

interface ConnectionComponent {
  name: string;
  summary?: ReactElement;
  body: ReactElement;
  header: ReactElement;
}

interface GetConnectionComponentDetailsProps {
  connection?: Connection;
  onSaved(updatedConnResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
  extraPageHeading?: ReactElement;
}

export function getConnectionComponentDetails(
  props: GetConnectionComponentDetailsProps
): ConnectionComponent {
  const { connection, onSaved, extraPageHeading, onSaveFailed } = props;

  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig':
      const value = connection.connectionConfig.config.value;
      switch (value.connectionConfig.case) {
        case 'connection':
          return {
            name: connection.name,
            summary: (
              <div>
                <p>No summary found.</p>
              </div>
            ),
            header: (
              <PageHeader
                header="PostgreSQL"
                description="Update this connection"
                leftIcon={<ConnectionIcon name="postgres" />}
                extraHeading={extraPageHeading}
              />
            ),
            body: (
              <PostgresForm
                connectionId={connection.id}
                defaultValues={{
                  connectionName: connection.name,
                  db: {
                    host: value.connectionConfig.value.host,
                    port: value.connectionConfig.value.port,
                    name: value.connectionConfig.value.name,
                    user: value.connectionConfig.value.user,
                    pass: value.connectionConfig.value.pass,
                    sslMode: value.connectionConfig.value.sslMode,
                  },
                }}
                onSaved={(resp) => onSaved(resp)}
                onSaveFailed={onSaveFailed}
              />
            ),
          };
      }
      return {
        name: connection.name,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
        header: (
          <PageHeader
            header="Unknown Connection"
            description="Update this connection"
          />
        ),
        body: (
          <div>
            No connection component found for: (
            {connection?.name ?? 'unknown name'})
          </div>
        ),
      };
    case 'awsS3Config':
      return {
        name: connection.name,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
        header: (
          <PageHeader
            header="AWS S3"
            description="Update this connection"
            leftIcon={<ConnectionIcon name="awsS3" />}
            extraHeading={extraPageHeading}
          />
        ),
        body: (
          <div>
            No connection component found for: (
            {connection?.name ?? 'unknown name'})
          </div>
        ),
      };
    default:
      return {
        name: 'Invalid Connection',
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
        header: (
          <PageHeader
            header="Unknown Connection"
            description="Update this connection"
          />
        ),
        body: (
          <div>
            No connection component found for: (
            {connection?.name ?? 'unknown name'})
          </div>
        ),
      };
  }
}
