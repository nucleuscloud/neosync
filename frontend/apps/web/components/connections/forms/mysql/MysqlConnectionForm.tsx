import { buildConnectionConfigMysql } from '@/app/(mgmt)/[account]/connections/util';
import { MysqlFormValues } from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Connection, MysqlConnectionConfig } from '@neosync/sdk';
import { ReactElement } from 'react';
import { ConnectionFormProps } from '../types';
import { useConnection } from '../useConnection';
import {
  getClientTlsFormValues,
  getSqlOptionsFormValues,
  getSshTunnelFormValues,
} from '../util';
import MysqlForm from './MysqlForm';

export default function MysqlConnectionForm(
  props: ConnectionFormProps
): ReactElement<any> {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigMysql,
    toFormValues,
  };

  const {
    isLoading,
    initialValues,
    handleSubmit,
    getValueWithSecrets,
    connectionId,
  } = useConnection<MysqlFormValues>(connectionProps);

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
      connectionId={connectionId}
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

  const connValues = getMysqlConnectionFormValues(
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
