import { buildConnectionConfigOpenAi } from '@/app/(mgmt)/[account]/connections/util';
import { OpenAiFormValues } from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';
import { ConnectionFormProps } from '../types';
import { useConnection } from '../useConnection';
import OpenAiForm from './OpenAiForm';

export default function OpenAiConnectionForm(
  props: ConnectionFormProps
): ReactElement<any> {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigOpenAi,
    toFormValues: (connection: Connection): OpenAiFormValues | undefined => {
      if (
        connection.connectionConfig?.config.case !== 'openaiConfig' ||
        !connection.connectionConfig?.config.value
      ) {
        return undefined;
      }
      return {
        connectionName: connection.name,
        sdk: {
          url: connection.connectionConfig.config.value.apiUrl,
          apiKey: connection.connectionConfig.config.value.apiKey,
        },
      };
    },
  };

  const { isLoading, initialValues, handleSubmit, getValueWithSecrets } =
    useConnection<OpenAiFormValues>(connectionProps);

  if (isLoading) {
    return <SkeletonForm />;
  }

  return (
    <OpenAiForm
      mode={mode === 'clone' ? 'create' : mode}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      canViewSecrets={mode === 'view'}
      getValueWithSecrets={getValueWithSecrets}
    />
  );
}
