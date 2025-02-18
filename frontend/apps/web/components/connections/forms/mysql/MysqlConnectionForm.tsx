import { buildConnectionConfigMysql } from '@/app/(mgmt)/[account]/connections/util';
import {
  ClientTlsFormValues,
  MysqlFormValues,
  SqlOptionsFormValues,
  SshTunnelFormValues,
} from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import {
  ClientTlsConfig,
  Connection,
  MysqlConnectionConfig,
  SSHAuthentication,
  SSHTunnel,
  SqlConnectionOptions,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { useConnection } from '../useConnection';
import MysqlForm from './MysqlForm';

interface CreateProps {
  mode: 'create';
  onSuccess(conn: Connection): Promise<void>;
}

interface EditProps {
  mode: 'edit';
  connection: Connection;
  onSuccess(conn: Connection): Promise<void>;
}

interface ViewProps {
  mode: 'view';
  connection: Connection;
}

interface CloneProps {
  mode: 'clone';
  connectionId: string;
  onSuccess(conn: Connection): Promise<void>;
}

type Props = CreateProps | EditProps | ViewProps | CloneProps;

export default function MysqlConnectionForm(props: Props): ReactElement {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigMysql,
    toFormValues,
  };

  const { isLoading, initialValues, handleSubmit, getValueWithSecrets } =
    useConnection<MysqlFormValues>(connectionProps);

  if (isLoading) {
    return <SkeletonForm />;
  }

  return (
    <MysqlForm
      mode={mode === 'clone' ? 'create' : mode}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      canViewSecrets={mode === 'view'}
      getValueWithSecrets={getValueWithSecrets}
    />
  );
}

function toFormValues(connection: Connection): MysqlFormValues | undefined {
  if (
    connection.connectionConfig?.config.case !== 'mysqlConfig' ||
    !connection.connectionConfig?.config.value
  ) {
    return undefined;
  }

  const { db, url, envVar } = getMysqlConnectionFormValues(
    connection.connectionConfig.config.value
  );

  return {
    connectionName: connection.name,
    db,
    url,
    envVar,
    options: getSqlOptionsFormValues(
      connection.connectionConfig.config.value.connectionOptions
    ),
    clientTls: getClientTlsFormValues(
      connection.connectionConfig.config.value.clientTls
    ),
    tunnel: getSshTunnelFormValues(
      connection.connectionConfig.config.value.tunnel
    ),
  };
}

// extracts the connection config and returns the values for the form
function getMysqlConnectionFormValues(
  connection: MysqlConnectionConfig
): Pick<MysqlFormValues, 'db' | 'url' | 'envVar' | 'activeTab'> {
  switch (connection.connectionConfig.case) {
    case 'connection':
      return {
        db: connection.connectionConfig.value,
        url: undefined,
        envVar: undefined,
        activeTab: 'host',
      };
    case 'url':
      return {
        db: {},
        url: connection.connectionConfig.value,
        envVar: undefined,
        activeTab: 'url',
      };
    case 'urlFromEnv':
      return {
        db: {},
        url: undefined,
        envVar: connection.connectionConfig.value,
        activeTab: 'url-env',
      };
    default:
      return {
        db: {},
        url: '',
        envVar: '',
        activeTab: 'url',
      };
  }
}

function getSqlOptionsFormValues(
  input: SqlConnectionOptions | undefined
): SqlOptionsFormValues {
  return {
    maxConnectionLimit: input?.maxConnectionLimit ?? 20,
    maxIdleDuration: input?.maxIdleDuration ?? '',
    maxIdleLimit: input?.maxIdleConnections ?? 2,
    maxOpenDuration: input?.maxOpenDuration ?? '',
  };
}

function getClientTlsFormValues(
  input: ClientTlsConfig | undefined
): ClientTlsFormValues {
  return {
    rootCert: input?.rootCert ?? '',
    clientCert: input?.clientCert ?? '',
    clientKey: input?.clientKey ?? '',
    serverName: input?.serverName ?? '',
  };
}

function getSshTunnelFormValues(
  input: SSHTunnel | undefined
): SshTunnelFormValues {
  return {
    host: input?.host ?? '',
    port: input?.port ?? 22,
    user: input?.user ?? '',
    privateKey: input?.authentication
      ? (getPrivateKeyFromSshAuthentication(input.authentication) ?? '')
      : '',
    passphrase: input?.authentication
      ? (getPassphraseFromSshAuthentication(input.authentication) ?? '')
      : '',
    knownHostPublicKey: input?.knownHostPublicKey ?? '',
  };
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
