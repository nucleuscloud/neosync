import { buildConnectionConfigOpenAi } from '@/app/(mgmt)/[account]/connections/util';
import { OpenAiFormValues } from '@/yup-validations/connections';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Connection } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useConnection } from '../useConnection';
import OpenAiForm from './OpenAiForm';

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

export default function OpenAiConnectionForm(props: Props): ReactElement {
  const { mode } = props;

  const connectionProps = {
    ...props,
    buildConnectionConfig: buildConnectionConfigOpenAi,
    toFormValues: (connection: Connection) => {
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
