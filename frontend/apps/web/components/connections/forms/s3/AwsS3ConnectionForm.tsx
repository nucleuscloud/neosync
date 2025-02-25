import { buildConnectionConfigAwsS3 } from '@/app/(mgmt)/[account]/connections/util';
import { AwsFormValues } from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';
import { ConnectionFormProps } from '../types';
import { useConnection } from '../useConnection';
import AwsS3Form from './AwsS3Form';

export default function AwsS3ConnectionForm(
  props: ConnectionFormProps
): ReactElement<any> {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigAwsS3,
    toFormValues,
  };

  const { isLoading, initialValues, handleSubmit, getValueWithSecrets } =
    useConnection<AwsFormValues>(connectionProps);

  if (isLoading) {
    return <SkeletonForm />;
  }

  return (
    <AwsS3Form
      mode={mode === 'clone' ? 'create' : mode}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      canViewSecrets={mode === 'view'}
      getValueWithSecrets={getValueWithSecrets}
    />
  );
}

function toFormValues(connection: Connection): AwsFormValues | undefined {
  if (
    connection.connectionConfig?.config.case !== 'awsS3Config' ||
    !connection.connectionConfig?.config.value
  ) {
    return undefined;
  }

  return {
    connectionName: connection.name,
    s3: {
      bucket: connection.connectionConfig.config.value.bucket,
      pathPrefix: connection.connectionConfig.config.value.pathPrefix,
    },
    advanced: {
      region: connection.connectionConfig.config.value.region,
      endpoint: connection.connectionConfig.config.value.endpoint,
    },
    credentials: connection.connectionConfig.config.value.credentials,
  };
}
