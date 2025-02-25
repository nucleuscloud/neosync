import { buildConnectionConfigGcpCloudStorage } from '@/app/(mgmt)/[account]/connections/util';
import { GcpCloudStorageFormValues } from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';
import { ConnectionFormProps } from '../types';
import { useConnection } from '../useConnection';
import GcpCloudStorageForm from './GcpCloudStorageForm';

export default function GcpCloudStorageConnectionForm(
  props: ConnectionFormProps
): ReactElement {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigGcpCloudStorage,
    toFormValues,
  };

  const { isLoading, initialValues, handleSubmit, getValueWithSecrets } =
    useConnection<GcpCloudStorageFormValues>(connectionProps);

  if (isLoading) {
    return <SkeletonForm />;
  }

  return (
    <GcpCloudStorageForm
      mode={mode === 'clone' ? 'create' : mode}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      canViewSecrets={mode === 'view'}
      getValueWithSecrets={getValueWithSecrets}
    />
  );
}

function toFormValues(
  connection: Connection
): GcpCloudStorageFormValues | undefined {
  if (
    connection.connectionConfig?.config.case !== 'gcpCloudstorageConfig' ||
    !connection.connectionConfig?.config.value
  ) {
    return undefined;
  }

  return {
    connectionName: connection.name,
    gcp: {
      bucket: connection.connectionConfig.config.value.bucket,
      pathPrefix: connection.connectionConfig.config.value.pathPrefix,
    },
  };
}
