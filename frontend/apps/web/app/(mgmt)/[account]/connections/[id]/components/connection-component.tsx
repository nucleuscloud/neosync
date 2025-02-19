'use client';
import ConnectionIcon from '@/components/connections/ConnectionIcon';
import DynamoDbConnectionForm from '@/components/connections/forms/dynamodb/DynamoDbConnectionForm';
import MongoDbConnectionForm from '@/components/connections/forms/mongodb/MongoDbConnectionForm';
import MysqlConnectionForm from '@/components/connections/forms/mysql/MysqlConnectionForm';
import OpenAiConnectionForm from '@/components/connections/forms/openai/OpenAiConnectionForm';
import PostgresConnectionForm from '@/components/connections/forms/postgres/PostgresConnectionForm';
import SqlServerConnectionForm from '@/components/connections/forms/sql-server/SqlServerConnectionForm';
import PageHeader from '@/components/headers/PageHeader';
import { Connection, PostgresConnectionConfig } from '@neosync/sdk';
import { ReactElement } from 'react';
import AwsS3Form from './AwsS3Form';
import GcpCloudStorageForm from './GcpCloudStorageForm';

interface ConnectionComponent {
  name: string;
  summary?: ReactElement;
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
  onSaveFailed?: never;
};

type EditModeProps = BaseConnectionComponentDetailsProps & {
  mode: 'edit';
  onSaved: (connection: Connection) => void;
  onSaveFailed: (err: unknown) => void;
};

// type CloneModeProps = BaseConnectionComponentDetailsProps & {
//   mode: 'clone';
//   connection?: never;

//   connectionId: string;
//   onSuccess: (connection: Connection) => void;
// };

type GetConnectionComponentDetailsProps = ViewModeProps | EditModeProps;

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

interface ModeViewProps {
  mode: 'view' | 'edit';
  view(): ReactElement;
  edit(): ReactElement;
}
function ModeView(props: ModeViewProps): ReactElement {
  const { mode, view, edit } = props;

  if (mode === 'view') {
    return view();
  }

  return edit();
}

export function useGetConnectionComponentDetails(
  props: GetConnectionComponentDetailsProps
): ConnectionComponent {
  const {
    connection,

    onSaved,
    extraPageHeading,
    onSaveFailed,
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
        name: connection.name,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
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
          />
          // <PostgresForm
          //   connectionId={connection.id}
          //   defaultValues={{
          //     connectionName: connection.name,
          //     db,
          //     url,
          //     envVar,
          //     options: {
          //       maxConnectionLimit: value.connectionOptions?.maxConnectionLimit,
          //       maxIdleDuration: value.connectionOptions?.maxIdleDuration,
          //       maxIdleLimit: value.connectionOptions?.maxIdleConnections,
          //       maxOpenDuration: value.connectionOptions?.maxOpenDuration,
          //     },
          //     clientTls: {
          //       rootCert: value.clientTls?.rootCert
          //         ? value.clientTls.rootCert
          //         : '',
          //       clientCert: value.clientTls?.clientCert
          //         ? value.clientTls.clientCert
          //         : '',
          //       clientKey: value.clientTls?.clientKey
          //         ? value.clientTls.clientKey
          //         : '',
          //       serverName: value.clientTls?.serverName
          //         ? value.clientTls.serverName
          //         : '',
          //     },
          //     tunnel: {
          //       host: value.tunnel?.host ?? '',
          //       port: value.tunnel?.port ?? 22,
          //       knownHostPublicKey: value.tunnel?.knownHostPublicKey ?? '',
          //       user: value.tunnel?.user ?? '',
          //       passphrase:
          //         value.tunnel && value.tunnel.authentication
          //           ? (getPassphraseFromSshAuthentication(
          //               value.tunnel.authentication
          //             ) ?? '')
          //           : '',
          //       privateKey:
          //         value.tunnel && value.tunnel.authentication
          //           ? (getPrivateKeyFromSshAuthentication(
          //               value.tunnel.authentication
          //             ) ?? '')
          //           : '',
          //     },
          //   }}
          //   onSaved={(resp) => onSaved?.(resp?.connection ?? connection)}
          //   onSaveFailed={(err) => onSaveFailed?.(err)}
          //   // onSaveFailed={onSaveFailed}
          // />
        ),
      };
    }

    case 'mysqlConfig': {
      // const mysqlValue = connection.connectionConfig.config.value;
      // const { db, url, envVar } = getMysqlConnectionFormValues(
      //   connection.connectionConfig.config.value
      // );

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
          />
          // <MysqlForm
          //   connectionId={connection.id}
          //   defaultValues={{
          //     connectionName: connection.name,
          //     db,
          //     url,
          //     envVar,
          //     options: {
          //       maxConnectionLimit:
          //         mysqlValue.connectionOptions?.maxConnectionLimit,
          //       maxIdleDuration: mysqlValue.connectionOptions?.maxIdleDuration,
          //       maxIdleLimit: mysqlValue.connectionOptions?.maxIdleConnections,
          //       maxOpenDuration: mysqlValue.connectionOptions?.maxOpenDuration,
          //     },
          //     tunnel: {
          //       host: mysqlValue.tunnel?.host ?? '',
          //       port: mysqlValue.tunnel?.port ?? 22,
          //       knownHostPublicKey: mysqlValue.tunnel?.knownHostPublicKey ?? '',
          //       user: mysqlValue.tunnel?.user ?? '',
          //       passphrase:
          //         mysqlValue.tunnel && mysqlValue.tunnel.authentication
          //           ? (getPassphraseFromSshAuthentication(
          //               mysqlValue.tunnel.authentication
          //             ) ?? '')
          //           : '',
          //       privateKey:
          //         mysqlValue.tunnel && mysqlValue.tunnel.authentication
          //           ? (getPrivateKeyFromSshAuthentication(
          //               mysqlValue.tunnel.authentication
          //             ) ?? '')
          //           : '',
          //     },
          //     clientTls: {
          //       rootCert: mysqlValue.clientTls?.rootCert
          //         ? mysqlValue.clientTls.rootCert
          //         : '',
          //       clientCert: mysqlValue.clientTls?.clientCert
          //         ? mysqlValue.clientTls.clientCert
          //         : '',
          //       clientKey: mysqlValue.clientTls?.clientKey
          //         ? mysqlValue.clientTls.clientKey
          //         : '',
          //       serverName: mysqlValue.clientTls?.serverName
          //         ? mysqlValue.clientTls.serverName
          //         : '',
          //     },
          //   }}
          //   onSaved={(resp) => onSaved?.(resp?.connection ?? connection)}
          //   onSaveFailed={(err) => onSaveFailed?.(err)}
          // />
        ),
      };
    }

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
            leftIcon={<ConnectionIcon connectionType="awsS3Config" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <AwsS3Form
            connectionId={connection.id}
            defaultValues={{
              connectionName: connection.name,
              s3: {
                bucket: connection.connectionConfig.config.value.bucket,
                pathPrefix: connection.connectionConfig.config.value.pathPrefix,
                credentials: {
                  accessKeyId:
                    connection.connectionConfig.config.value.credentials
                      ?.accessKeyId,
                  secretAccessKey:
                    connection.connectionConfig.config.value.credentials
                      ?.secretAccessKey,
                  sessionToken:
                    connection.connectionConfig.config.value.credentials
                      ?.sessionToken,
                  fromEc2Role:
                    connection.connectionConfig.config.value.credentials
                      ?.fromEc2Role,
                  roleArn:
                    connection.connectionConfig.config.value.credentials
                      ?.roleArn,
                  roleExternalId:
                    connection.connectionConfig.config.value.credentials
                      ?.roleExternalId,
                  profile:
                    connection.connectionConfig.config.value.credentials
                      ?.profile,
                },
                endpoint: connection.connectionConfig.config.value.endpoint,
                region: connection.connectionConfig.config.value.region,
              },
            }}
            onSaved={(resp) => onSaved?.(resp?.connection ?? connection)}
            onSaveFailed={(err) => onSaveFailed?.(err)}
          />
        ),
      };
    case 'openaiConfig':
      return {
        name: connection.name,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
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
          />
        ),
      };
    case 'mongoConfig':
      if (
        connection.connectionConfig.config.value.connectionConfig.case !== 'url'
      ) {
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
      return {
        name: connection.name,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
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
          />
        ),
      };
    case 'gcpCloudstorageConfig': {
      return {
        name: connection.name,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
        header: (
          <PageHeader
            header="GCP Cloud Storage"
            leftIcon={<ConnectionIcon connectionType="gcpCloudstorageConfig" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <GcpCloudStorageForm
            connectionId={connection.id}
            defaultValues={{
              connectionName: connection.name,
              gcp: {
                bucket: connection.connectionConfig.config.value.bucket,
                pathPrefix: connection.connectionConfig.config.value.pathPrefix,
              },
            }}
            onSaved={(resp) => onSaved?.(resp?.connection ?? connection)}
            onSaveFailed={(err) => onSaveFailed?.(err)}
          />
        ),
      };
    }
    case 'dynamodbConfig': {
      return {
        name: connection.name,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
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
          />
        ),
      };
    }
    case 'mssqlConfig': {
      return {
        name: connection.name,
        summary: (
          <div>
            <p>No summary found.</p>
          </div>
        ),
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
          />
        ),
      };
    }
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
