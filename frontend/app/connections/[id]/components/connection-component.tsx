'use client';
import ConnectionIcon from '@/components/connections/ConnectionIcon';
import PageHeader from '@/components/headers/PageHeader';
import {
  Connection,
  UpdateConnectionResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { ReactElement } from 'react';
import AwsS3Form from './AwsS3Form';
import MysqlForm from './MysqlForm';
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
        header: <PageHeader header="Unknown Connection" />,
        body: (
          <div>
            No connection component found for: (
            {connection?.name ?? 'unknown name'})
          </div>
        ),
      };
    case 'mysqlConfig':
      const mysqlValue = connection.connectionConfig.config.value;
      switch (mysqlValue.connectionConfig.case) {
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
                header="Mysql"
                description="Update this connection"
                leftIcon={<ConnectionIcon name="mysql" />}
                extraHeading={extraPageHeading}
              />
            ),
            body: (
              <MysqlForm
                connectionId={connection.id}
                defaultValues={{
                  connectionName: connection.name,
                  db: {
                    host: mysqlValue.connectionConfig.value.host,
                    port: mysqlValue.connectionConfig.value.port,
                    name: mysqlValue.connectionConfig.value.name,
                    user: mysqlValue.connectionConfig.value.user,
                    pass: mysqlValue.connectionConfig.value.pass,
                    protocol: mysqlValue.connectionConfig.value.protocol,
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
        header: <PageHeader header="Unknown Connection" />,
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
            leftIcon={<ConnectionIcon name="awsS3" />}
            extraHeading={extraPageHeading}
          />
        ),
        body: (
          <AwsS3Form
            connectionId={connection.id}
            defaultValues={{
              connectionName: connection.name,
              s3: {
                bucketArn: connection.connectionConfig.config.value.bucketArn,
                pathPrefix: connection.connectionConfig.config.value.pathPrefix,
                credentials:
                  connection.connectionConfig.config.value.credentials,
                endpoint: connection.connectionConfig.config.value.endpoint,
                region: connection.connectionConfig.config.value.region,
              },
            }}
            onSaved={(resp) => onSaved(resp)}
            onSaveFailed={onSaveFailed}
          />
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
        header: <PageHeader header="Unknown Connection" />,
        body: (
          <div>
            No connection component found for: (
            {connection?.name ?? 'unknown name'})
          </div>
        ),
      };
  }
}
