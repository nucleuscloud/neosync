import { buildConnectionConfigDynamoDB } from '@/app/(mgmt)/[account]/connections/util';
import { DynamoDbFormValues } from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';
import { ConnectionFormProps } from '../types';
import { useConnection } from '../useConnection';
import DynamoDbForm from './DynamoDbForm';

export default function DynamoDbConnectionForm(
  props: ConnectionFormProps
): ReactElement {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigDynamoDB,
    toFormValues,
  };

  const {
    isLoading,
    initialValues,
    handleSubmit,
    getValueWithSecrets,
    connectionId,
  } = useConnection<DynamoDbFormValues>(connectionProps);

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
      connectionId={connectionId}
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
    advanced: {
      region: connection.connectionConfig.config.value.region,
      endpoint: connection.connectionConfig.config.value.endpoint,
    },
    credentials: connection.connectionConfig.config.value.credentials,
  };
}
