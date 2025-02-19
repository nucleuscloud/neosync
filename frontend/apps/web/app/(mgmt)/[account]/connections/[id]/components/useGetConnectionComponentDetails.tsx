'use client';
import ConnectionIcon from '@/components/connections/ConnectionIcon';
import DynamoDbConnectionForm from '@/components/connections/forms/dynamodb/DynamoDbConnectionForm';
import GcpCloudStorageConnectionForm from '@/components/connections/forms/gcp-cloud-storage/GcpCloudStorageConnectionForm';
import MongoDbConnectionForm from '@/components/connections/forms/mongodb/MongoDbConnectionForm';
import MysqlConnectionForm from '@/components/connections/forms/mysql/MysqlConnectionForm';
import OpenAiConnectionForm from '@/components/connections/forms/openai/OpenAiConnectionForm';
import PostgresConnectionForm from '@/components/connections/forms/postgres/PostgresConnectionForm';
import AwsS3ConnectionForm from '@/components/connections/forms/s3/AwsS3ConnectionForm';
import SqlServerConnectionForm from '@/components/connections/forms/sql-server/SqlServerConnectionForm';
import PageHeader from '@/components/headers/PageHeader';
import { Connection, PostgresConnectionConfig } from '@neosync/sdk';
import { ReactElement } from 'react';
import ModeView from './ModeView';

interface ConnectionComponent {
  body: ReactElement;
  header: ReactElement;
}

type BaseConnectionComponentDetailsProps = {
  connection?: Connection;
  extraPageHeading?: ReactElement;
  subHeading?: ReactElement;
};

type ViewModeProps = BaseConnectionComponentDetailsProps & {
  mode?: 'view';
  onSaved?: never;
};

type EditModeProps = BaseConnectionComponentDetailsProps & {
  mode: 'edit';
  onSaved(connection: Connection): Promise<void> | void;
};

type CloneModeProps = BaseConnectionComponentDetailsProps & {
  mode: 'clone';
  connection?: Connection;
  onSaved(connection: Connection): Promise<void> | void;
};

type CreateModeProps = BaseConnectionComponentDetailsProps & {
  mode: 'create';
  connection?: never;

  onSaved(connection: Connection): Promise<void> | void;
};

type GetConnectionComponentDetailsProps =
  | ViewModeProps
  | EditModeProps
  | CloneModeProps;

function getPgHeaderType(
  connection: PostgresConnectionConfig
): 'generic' | 'neon' {
  switch (connection.connectionConfig.case) {
    case 'connection':
      if (connection.connectionConfig.value.host.includes('neon')) {
        return 'neon';
      }
      return 'generic';
    case 'url':
      if (connection.connectionConfig.value.includes('neon')) {
        return 'neon';
      }
      return 'generic';
    case 'urlFromEnv':
      return 'generic';
    default:
      return 'generic';
  }
}

export function useGetConnectionComponentDetails(
  props: GetConnectionComponentDetailsProps
): ConnectionComponent {
  const {
    connection,

    onSaved,
    extraPageHeading,
    subHeading,
    mode = 'view',
  } = props;

  async function onSuccess(connection: Connection): Promise<void> {
    if (onSaved) {
      onSaved(connection);
    }
  }

  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig': {
      const value = connection.connectionConfig.config.value;
      const headerType = getPgHeaderType(value);

      return {
        header: (
          <PageHeader
            header={headerType == 'neon' ? 'Neon' : 'PostgreSQL'}
            leftIcon={
              headerType == 'neon' ? (
                <ConnectionIcon
                  connectionType="pgConfig"
                  connectionTypeVariant="neon"
                />
              ) : (
                <ConnectionIcon connectionType="pgConfig" />
              )
            }
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <ModeView
            mode={mode}
            view={() => (
              <PostgresConnectionForm mode="view" connection={connection} />
            )}
            edit={() => (
              <PostgresConnectionForm
                mode="edit"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
            clone={() => (
              <PostgresConnectionForm
                mode="clone"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
          />
        ),
      };
    }

    case 'mysqlConfig': {
      return {
        header: (
          <PageHeader
            header="Mysql"
            leftIcon={<ConnectionIcon connectionType="mysqlConfig" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <ModeView
            mode={mode}
            view={() => (
              <MysqlConnectionForm mode="view" connection={connection} />
            )}
            edit={() => (
              <MysqlConnectionForm
                mode="edit"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
            clone={() => (
              <MysqlConnectionForm
                mode="clone"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
          />
        ),
      };
    }

    case 'awsS3Config':
      return {
        header: (
          <PageHeader
            header="AWS S3"
            leftIcon={<ConnectionIcon connectionType="awsS3Config" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <ModeView
            mode={mode}
            view={() => (
              <AwsS3ConnectionForm mode="view" connection={connection} />
            )}
            edit={() => (
              <AwsS3ConnectionForm
                mode="edit"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
            clone={() => (
              <AwsS3ConnectionForm
                mode="clone"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
          />
        ),
      };
    case 'openaiConfig':
      return {
        header: (
          <PageHeader
            header="OpenAI"
            leftIcon={<ConnectionIcon connectionType="openaiConfig" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <ModeView
            mode={mode}
            view={() => (
              <OpenAiConnectionForm mode="view" connection={connection} />
            )}
            edit={() => (
              <OpenAiConnectionForm
                mode="edit"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
            clone={() => (
              <OpenAiConnectionForm
                mode="clone"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
          />
        ),
      };
    case 'mongoConfig':
      if (
        connection.connectionConfig.config.value.connectionConfig.case !== 'url'
      ) {
        return {
          header: <PageHeader header="Invalid MongoDB Connection" />,
          body: (
            <div>
              No connection component found for: (
              {connection?.name ?? 'unknown name'})
            </div>
          ),
        };
      }
      return {
        header: (
          <PageHeader
            header="MongoDB"
            leftIcon={<ConnectionIcon connectionType="mongoConfig" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <ModeView
            mode={mode}
            view={() => (
              <MongoDbConnectionForm mode="view" connection={connection} />
            )}
            edit={() => (
              <MongoDbConnectionForm
                mode="edit"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
            clone={() => (
              <MongoDbConnectionForm
                mode="clone"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
          />
        ),
      };
    case 'gcpCloudstorageConfig': {
      return {
        header: (
          <PageHeader
            header="GCP Cloud Storage"
            leftIcon={<ConnectionIcon connectionType="gcpCloudstorageConfig" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <ModeView
            mode={mode}
            view={() => (
              <GcpCloudStorageConnectionForm
                mode="view"
                connection={connection}
              />
            )}
            edit={() => (
              <GcpCloudStorageConnectionForm
                mode="edit"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
            clone={() => (
              <GcpCloudStorageConnectionForm
                mode="clone"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
          />
        ),
      };
    }
    case 'dynamodbConfig': {
      return {
        header: (
          <PageHeader
            header="DynamoDB"
            leftIcon={<ConnectionIcon connectionType="dynamodbConfig" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <ModeView
            mode={mode}
            view={() => (
              <DynamoDbConnectionForm mode="view" connection={connection} />
            )}
            edit={() => (
              <DynamoDbConnectionForm
                mode="edit"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
            clone={() => (
              <DynamoDbConnectionForm
                mode="clone"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
          />
        ),
      };
    }
    case 'mssqlConfig': {
      return {
        header: (
          <PageHeader
            header="Microsoft SQL Server"
            leftIcon={<ConnectionIcon connectionType="mssqlConfig" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <ModeView
            mode={mode}
            view={() => (
              <SqlServerConnectionForm mode="view" connection={connection} />
            )}
            edit={() => (
              <SqlServerConnectionForm
                mode="edit"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
            clone={() => (
              <SqlServerConnectionForm
                mode="clone"
                connection={connection}
                onSuccess={onSuccess}
              />
            )}
          />
        ),
      };
    }
    default:
      return {
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
