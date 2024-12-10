'use client';
import ConnectionIcon from '@/components/connections/ConnectionIcon';
import PageHeader from '@/components/headers/PageHeader';
import {
  Connection,
  MysqlConnection,
  PostgresConnection,
  SSHAuthentication,
  UpdateConnectionResponse,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import AwsS3Form from './AwsS3Form';
import DynamoDBForm from './DynamoDBForm';
import GcpCloudStorageForm from './GcpCloudStorageForm';
import MongoDbForm from './MongoDbForm';
import MssqlForm from './MssqlForm';
import MysqlForm from './MysqlForm';
import OpenAiForm from './OpenAiForm';
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
  subHeading?: ReactElement;
}

export function getConnectionComponentDetails(
  props: GetConnectionComponentDetailsProps
): ConnectionComponent {
  const { connection, onSaved, extraPageHeading, onSaveFailed, subHeading } =
    props;

  switch (connection?.connectionConfig?.config?.case) {
    case 'pgConfig':
      const value = connection.connectionConfig.config.value;
      let pgConfig: PostgresConnection | string | undefined;

      let dbConfig = {
        host: '',
        name: '',
        user: '',
        pass: '',
        port: 5432,
        sslMode: 'disable',
      };

      // define header type for postgres, either generic postgres or neon
      let headerType = 'generic';

      switch (value.connectionConfig.case) {
        case 'connection':
          pgConfig = value.connectionConfig.value;
          dbConfig = {
            host: pgConfig.host ?? '',
            name: pgConfig.name ?? '',
            user: pgConfig.user ?? '',
            pass: pgConfig.pass ?? '',
            port: pgConfig.port ?? 5432,
            sslMode: pgConfig.sslMode ?? 'disable',
          };
          if (pgConfig.host.includes('neon')) {
            headerType = 'neon';
          } else {
            headerType = 'generic';
          }
          break;
        case 'url':
          pgConfig = value.connectionConfig.value;
          if (pgConfig.includes('neon')) {
            headerType = 'neon';
          } else {
            headerType = 'generic';
          }
          break;
        default:
          pgConfig = value.connectionConfig.value;
          dbConfig = dbConfig;
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
          <PostgresForm
            connectionId={connection.id}
            defaultValues={{
              connectionName: connection.name,
              db: dbConfig,
              url: typeof pgConfig === 'string' ? pgConfig : '',
              options: {
                maxConnectionLimit: value.connectionOptions?.maxConnectionLimit,
                maxIdleDuration: value.connectionOptions?.maxIdleDuration,
                maxIdleLimit: value.connectionOptions?.maxIdleConnections,
                maxOpenDuration: value.connectionOptions?.maxOpenDuration,
              },
              clientTls: {
                rootCert: value.clientTls?.rootCert
                  ? value.clientTls.rootCert
                  : '',
                clientCert: value.clientTls?.clientCert
                  ? value.clientTls.clientCert
                  : '',
                clientKey: value.clientTls?.clientKey
                  ? value.clientTls.clientKey
                  : '',
                serverName: value.clientTls?.serverName
                  ? value.clientTls.serverName
                  : '',
              },
              tunnel: {
                host: value.tunnel?.host ?? '',
                port: value.tunnel?.port ?? 22,
                knownHostPublicKey: value.tunnel?.knownHostPublicKey ?? '',
                user: value.tunnel?.user ?? '',
                passphrase:
                  value.tunnel && value.tunnel.authentication
                    ? (getPassphraseFromSshAuthentication(
                        value.tunnel.authentication
                      ) ?? '')
                    : '',
                privateKey:
                  value.tunnel && value.tunnel.authentication
                    ? (getPrivateKeyFromSshAuthentication(
                        value.tunnel.authentication
                      ) ?? '')
                    : '',
              },
            }}
            onSaved={(resp) => onSaved(resp)}
            onSaveFailed={onSaveFailed}
          />
        ),
      };

    case 'mysqlConfig':
      const mysqlValue = connection.connectionConfig.config.value;

      let mysqlConfig: MysqlConnection | string | undefined;

      let mysqldbConfig = {
        host: '',
        name: '',
        user: '',
        pass: '',
        port: 5432,
        protocol: '',
      };

      switch (mysqlValue.connectionConfig.case) {
        case 'connection':
          mysqlConfig = mysqlValue.connectionConfig.value;
          mysqldbConfig = {
            host: mysqlConfig.host,
            name: mysqlConfig.name,
            user: mysqlConfig.user,
            pass: mysqlConfig.pass,
            port: mysqlConfig.port,
            protocol: mysqlConfig.protocol,
          };
          break;
        case 'url':
          mysqlConfig = mysqlValue.connectionConfig.value;
          break;
        default:
          mysqlConfig = mysqlValue.connectionConfig.value;
          mysqldbConfig = mysqldbConfig;
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
            header="Mysql"
            leftIcon={<ConnectionIcon connectionType="mysqlConfig" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <MysqlForm
            connectionId={connection.id}
            defaultValues={{
              connectionName: connection.name,
              db: mysqldbConfig,
              url: typeof mysqlConfig === 'string' ? mysqlConfig : '',
              options: {
                maxConnectionLimit:
                  mysqlValue.connectionOptions?.maxConnectionLimit,
                maxIdleDuration: mysqlValue.connectionOptions?.maxIdleDuration,
                maxIdleLimit: mysqlValue.connectionOptions?.maxIdleConnections,
                maxOpenDuration: mysqlValue.connectionOptions?.maxOpenDuration,
              },
              tunnel: {
                host: mysqlValue.tunnel?.host ?? '',
                port: mysqlValue.tunnel?.port ?? 22,
                knownHostPublicKey: mysqlValue.tunnel?.knownHostPublicKey ?? '',
                user: mysqlValue.tunnel?.user ?? '',
                passphrase:
                  mysqlValue.tunnel && mysqlValue.tunnel.authentication
                    ? (getPassphraseFromSshAuthentication(
                        mysqlValue.tunnel.authentication
                      ) ?? '')
                    : '',
                privateKey:
                  mysqlValue.tunnel && mysqlValue.tunnel.authentication
                    ? (getPrivateKeyFromSshAuthentication(
                        mysqlValue.tunnel.authentication
                      ) ?? '')
                    : '',
              },
              clientTls: {
                rootCert: mysqlValue.clientTls?.rootCert
                  ? mysqlValue.clientTls.rootCert
                  : '',
                clientCert: mysqlValue.clientTls?.clientCert
                  ? mysqlValue.clientTls.clientCert
                  : '',
                clientKey: mysqlValue.clientTls?.clientKey
                  ? mysqlValue.clientTls.clientKey
                  : '',
                serverName: mysqlValue.clientTls?.serverName
                  ? mysqlValue.clientTls.serverName
                  : '',
              },
            }}
            onSaved={(resp) => onSaved(resp)}
            onSaveFailed={onSaveFailed}
          />
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
            onSaved={(resp) => onSaved(resp)}
            onSaveFailed={onSaveFailed}
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
          <OpenAiForm
            connectionId={connection.id}
            defaultValues={{
              connectionName: connection.name,
              sdk: {
                url: connection.connectionConfig.config.value.apiUrl,
                apiKey: connection.connectionConfig.config.value.apiKey,
              },
            }}
            onSaved={(resp) => onSaved(resp)}
            onSaveFailed={onSaveFailed}
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
          <MongoDbForm
            connectionId={connection.id}
            defaultValues={{
              connectionName: connection.name,
              url: connection.connectionConfig.config.value.connectionConfig
                .value,

              clientTls: {
                rootCert: connection.connectionConfig.config.value.clientTls
                  ?.rootCert
                  ? connection.connectionConfig.config.value.clientTls.rootCert
                  : '',
                clientCert: connection.connectionConfig.config.value.clientTls
                  ?.clientCert
                  ? connection.connectionConfig.config.value.clientTls
                      .clientCert
                  : '',
                clientKey: connection.connectionConfig.config.value.clientTls
                  ?.clientKey
                  ? connection.connectionConfig.config.value.clientTls.clientKey
                  : '',
                serverName: connection.connectionConfig.config.value.clientTls
                  ?.serverName
                  ? connection.connectionConfig.config.value.clientTls
                      .serverName
                  : '',
              },
            }}
            onSaved={(resp) => onSaved(resp)}
            onSaveFailed={onSaveFailed}
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
            onSaved={(resp) => onSaved(resp)}
            onSaveFailed={onSaveFailed}
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
          <DynamoDBForm
            connectionId={connection.id}
            defaultValues={{
              connectionName: connection.name,
              db: {
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
            onSaved={(resp) => onSaved(resp)}
            onSaveFailed={onSaveFailed}
          />
        ),
      };
    }
    case 'mssqlConfig': {
      const mssqlValue = connection.connectionConfig.config.value;

      let mssqlConfig: string | undefined;

      switch (mssqlValue.connectionConfig.case) {
        case 'url':
          mssqlConfig = mssqlValue.connectionConfig.value;
          break;
        default:
          mssqlConfig = mssqlValue.connectionConfig.value;
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
            header="Microsoft SQL Server"
            leftIcon={<ConnectionIcon connectionType="mssqlConfig" />}
            extraHeading={extraPageHeading}
            subHeadings={subHeading}
          />
        ),
        body: (
          <MssqlForm
            connectionId={connection.id}
            defaultValues={{
              connectionName: connection.name,
              db: {
                url: typeof mssqlConfig === 'string' ? mssqlConfig : '',
              },
              options: {
                maxConnectionLimit:
                  mssqlValue.connectionOptions?.maxConnectionLimit,
                maxIdleDuration: mssqlValue.connectionOptions?.maxIdleDuration,
                maxIdleLimit: mssqlValue.connectionOptions?.maxIdleConnections,
                maxOpenDuration: mssqlValue.connectionOptions?.maxOpenDuration,
              },
              tunnel: {
                host: mssqlValue.tunnel?.host ?? '',
                port: mssqlValue.tunnel?.port ?? 22,
                knownHostPublicKey: mssqlValue.tunnel?.knownHostPublicKey ?? '',
                user: mssqlValue.tunnel?.user ?? '',
                passphrase:
                  mssqlValue.tunnel && mssqlValue.tunnel.authentication
                    ? (getPassphraseFromSshAuthentication(
                        mssqlValue.tunnel.authentication
                      ) ?? '')
                    : '',
                privateKey:
                  mssqlValue.tunnel && mssqlValue.tunnel.authentication
                    ? (getPrivateKeyFromSshAuthentication(
                        mssqlValue.tunnel.authentication
                      ) ?? '')
                    : '',
              },
              clientTls: {
                rootCert: mssqlValue.clientTls?.rootCert
                  ? mssqlValue.clientTls.rootCert
                  : '',
                clientCert: mssqlValue.clientTls?.clientCert
                  ? mssqlValue.clientTls.clientCert
                  : '',
                clientKey: mssqlValue.clientTls?.clientKey
                  ? mssqlValue.clientTls.clientKey
                  : '',
                serverName: mssqlValue.clientTls?.serverName
                  ? mssqlValue.clientTls.serverName
                  : '',
              },
            }}
            onSaved={(resp) => onSaved(resp)}
            onSaveFailed={onSaveFailed}
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

function getPassphraseFromSshAuthentication(
  sshauth: SSHAuthentication
): string | undefined {
  switch (sshauth.authConfig.case) {
    case 'passphrase':
      return sshauth.authConfig.value.value;
    case 'privateKey':
      return sshauth.authConfig.value.passphrase;
    default:
      return undefined;
  }
}

function getPrivateKeyFromSshAuthentication(
  sshauth: SSHAuthentication
): string | undefined {
  switch (sshauth.authConfig.case) {
    case 'privateKey':
      return sshauth.authConfig.value.value;
    default:
      return undefined;
  }
}
