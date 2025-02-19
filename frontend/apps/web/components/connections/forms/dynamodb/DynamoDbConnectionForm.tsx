import { buildConnectionConfigDynamoDB } from '@/app/(mgmt)/[account]/connections/util';
import { DynamoDbFormValues } from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useConnection } from '../useConnection';
import DynamoDbForm from './DynamoDbForm';

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

export default function DynamoDbConnectionForm(props: Props): ReactElement {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigDynamoDB,
    toFormValues,
  };

  const { isLoading, initialValues, handleSubmit, getValueWithSecrets } =
    useConnection<DynamoDbFormValues>(connectionProps);

  if (isLoading) {
    return <SkeletonForm />;
  }

  return (
    <DynamoDbForm
      mode={mode === 'clone' ? 'create' : mode}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      canViewSecrets={mode === 'view'}
      getValueWithSecrets={getValueWithSecrets}
    />
  );
}

function toFormValues(connection: Connection): DynamoDbFormValues | undefined {
  if (
    connection.connectionConfig?.config.case !== 'dynamodbConfig' ||
    !connection.connectionConfig?.config.value
  ) {
    return undefined;
  }

  return {
    connectionName: connection.name,
    db: {
      region: connection.connectionConfig.config.value.region,
      endpoint: connection.connectionConfig.config.value.endpoint,
      credentials: connection.connectionConfig.config.value.credentials,
    },
  };
}
