'use client';
import ConnectionIcon from '@/components/connections/ConnectionIcon';
import PageHeader from '@/components/headers/PageHeader';
import {
  Connection,
  PostgresConnection,
  SSHAuthentication,
  UpdateConnectionResponse,
} from '@neosync/sdk';
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
                <ConnectionIcon name="neon" />
              ) : (
                <ConnectionIcon name="postgres" />
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
              tunnel: {
                host: value.tunnel?.host ?? '',
                port: value.tunnel?.port ?? 22,
                knownHostPublicKey: value.tunnel?.knownHostPublicKey ?? '',
                user: value.tunnel?.user ?? '',
                passphrase:
                  value.tunnel && value.tunnel.authentication
                    ? getPassphraseFromSshAuthentication(
                        value.tunnel.authentication
                      ) ?? ''
                    : '',
                privateKey:
                  value.tunnel && value.tunnel.authentication
                    ? getPrivateKeyFromSshAuthentication(
                        value.tunnel.authentication
                      ) ?? ''
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
                leftIcon={<ConnectionIcon name="mysql" />}
                extraHeading={extraPageHeading}
                subHeadings={subHeading}
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
                  tunnel: {
                    host: mysqlValue.tunnel?.host ?? '',
                    port: mysqlValue.tunnel?.port ?? 22,
                    knownHostPublicKey:
                      mysqlValue.tunnel?.knownHostPublicKey ?? '',
                    user: mysqlValue.tunnel?.user ?? '',
                    passphrase:
                      mysqlValue.tunnel && mysqlValue.tunnel.authentication
                        ? getPassphraseFromSshAuthentication(
                            mysqlValue.tunnel.authentication
                          ) ?? ''
                        : '',
                    privateKey:
                      mysqlValue.tunnel && mysqlValue.tunnel.authentication
                        ? getPrivateKeyFromSshAuthentication(
                            mysqlValue.tunnel.authentication
                          ) ?? ''
                        : '',
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
            leftIcon={<ConnectionIcon name="aws S3" />}
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
