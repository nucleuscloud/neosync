import { buildConnectionConfigPostgres } from '@/app/(mgmt)/[account]/connections/util';
import { PostgresFormValues } from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Connection, PostgresConnectionConfig } from '@neosync/sdk';
import { ReactElement } from 'react';
import { ConnectionFormProps } from '../types';
import { useConnection } from '../useConnection';
import {
  getClientTlsFormValues,
  getSqlOptionsFormValues,
  getSshTunnelFormValues,
} from '../util';
import PostgresForm from './PostgresForm';

export default function PostgresConnectionForm(
  props: ConnectionFormProps
): ReactElement {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigPostgres,
    toFormValues,
  };

  const { isLoading, initialValues, handleSubmit, getValueWithSecrets } =
    useConnection<PostgresFormValues>(connectionProps);

  if (isLoading) {
    return <SkeletonForm />;
  }

  return (
    <PostgresForm
      mode={mode === 'clone' ? 'create' : mode}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      canViewSecrets={mode === 'view'}
      getValueWithSecrets={getValueWithSecrets}
    />
  );
}

function toFormValues(connection: Connection): PostgresFormValues | undefined {
  if (
    connection.connectionConfig?.config.case !== 'pgConfig' ||
    !connection.connectionConfig?.config.value
  ) {
    return undefined;
  }

  const connValues = getPgConnectionFormValues(
    connection.connectionConfig.config.value
  );

  return {
    ...connValues,
    connectionName: connection.name,
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
function getPgConnectionFormValues(
  connection: PostgresConnectionConfig
): Pick<PostgresFormValues, 'db' | 'url' | 'envVar' | 'activeTab'> {
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
