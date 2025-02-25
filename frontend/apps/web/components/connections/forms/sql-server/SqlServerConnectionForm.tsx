import { buildConnectionConfigMssql } from '@/app/(mgmt)/[account]/connections/util';
import { MssqlFormValues } from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Connection, MssqlConnectionConfig } from '@neosync/sdk';
import { ReactElement } from 'react';
import { ConnectionFormProps } from '../types';
import { useConnection } from '../useConnection';
import {
  getClientTlsFormValues,
  getSqlOptionsFormValues,
  getSshTunnelFormValues,
} from '../util';
import SqlServerForm from './SqlServerForm';

export default function SqlServerConnectionForm(
  props: ConnectionFormProps
): ReactElement<any> {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigMssql,
    toFormValues,
  };

  const {
    isLoading,
    initialValues,
    handleSubmit,
    getValueWithSecrets,
    connectionId,
  } = useConnection<MssqlFormValues>(connectionProps);

  if (isLoading) {
    return <SkeletonForm />;
  }

  return (
    <SqlServerForm
      mode={mode === 'clone' ? 'create' : mode}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      canViewSecrets={mode === 'view'}
      getValueWithSecrets={getValueWithSecrets}
      connectionId={connectionId}
    />
  );
}

function toFormValues(connection: Connection): MssqlFormValues | undefined {
  if (
    connection.connectionConfig?.config.case !== 'mssqlConfig' ||
    !connection.connectionConfig?.config.value
  ) {
    return undefined;
  }

  const connValues = getMssqlConnectionFormValues(
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
function getMssqlConnectionFormValues(
  connection: MssqlConnectionConfig
): Pick<MssqlFormValues, 'url' | 'envVar' | 'activeTab'> {
  switch (connection.connectionConfig.case) {
    case 'url':
      return {
        url: connection.connectionConfig.value,
        envVar: undefined,
        activeTab: 'url',
      };
    case 'urlFromEnv':
      return {
        url: undefined,
        envVar: connection.connectionConfig.value,
        activeTab: 'url-env',
      };
    default:
      return {
        url: '',
        envVar: '',
        activeTab: 'url',
      };
  }
}
