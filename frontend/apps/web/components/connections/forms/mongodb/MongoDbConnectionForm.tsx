import { buildConnectionConfigMongo } from '@/app/(mgmt)/[account]/connections/util';
import { MongoDbFormValues } from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Connection, MongoConnectionConfig } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useConnection } from '../useConnection';
import { getClientTlsFormValues } from '../util';
import MongoDbForm from './MongoDbForm';

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

export default function MongoDbConnectionForm(props: Props): ReactElement {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigMongo,
    toFormValues,
  };

  const { isLoading, initialValues, handleSubmit, getValueWithSecrets } =
    useConnection<MongoDbFormValues>(connectionProps);

  if (isLoading) {
    return <SkeletonForm />;
  }

  return (
    <MongoDbForm
      mode={mode === 'clone' ? 'create' : mode}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      canViewSecrets={mode === 'view'}
      getValueWithSecrets={getValueWithSecrets}
    />
  );
}

function toFormValues(connection: Connection): MongoDbFormValues | undefined {
  if (
    connection.connectionConfig?.config.case !== 'mongoConfig' ||
    !connection.connectionConfig?.config.value
  ) {
    return undefined;
  }

  const connValues = getMongoDbConnectionFormValues(
    connection.connectionConfig.config.value
  );

  return {
    ...connValues,
    connectionName: connection.name,
    clientTls: getClientTlsFormValues(
      connection.connectionConfig.config.value.clientTls
    ),
  };
}

// extracts the connection config and returns the values for the form
function getMongoDbConnectionFormValues(
  connection: MongoConnectionConfig
): Pick<MongoDbFormValues, 'url'> {
  switch (connection.connectionConfig.case) {
    case 'url':
      return {
        url: connection.connectionConfig.value,
      };
    default:
      return {
        url: '',
      };
  }
}
