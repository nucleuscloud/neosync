import { buildConnectionConfigOpenAi } from '@/app/(mgmt)/[account]/connections/util';
import { OpenAiFormValues } from '@/yup-validations/connections';
import { create as createMessage } from '@bufbuild/protobuf';
import {
  CreateConnectionRequest,
  CreateConnectionRequestSchema,
  UpdateConnectionRequest,
  UpdateConnectionRequestSchema,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import OpenAiForm from './OpenAiForm';

interface BaseProps {
  initialValues?: OpenAiFormValues;
}

interface CreateProps extends BaseProps {
  mode: 'create';
  accountId: string;
  onSubmit(values: CreateConnectionRequest): Promise<void>;
}

interface EditProps extends BaseProps {
  mode: 'edit';
  connectionId: string;
  onSubmit(values: UpdateConnectionRequest): Promise<void>;
}

interface ViewProps extends BaseProps {
  mode: 'view';
}

type Props = CreateProps | EditProps | ViewProps;

export default function OpenAiConnectionForm(props: Props): ReactElement {
  const { mode, initialValues } = props;

  async function handleSubmit(values: OpenAiFormValues): Promise<void> {
    if (mode === 'view') {
      return;
    }

    if (mode === 'create') {
      await props.onSubmit(
        createMessage(CreateConnectionRequestSchema, {
          accountId: props.accountId,
          name: values.connectionName,
          connectionConfig: buildConnectionConfigOpenAi(values),
        })
      );
    } else {
      await props.onSubmit(
        createMessage(UpdateConnectionRequestSchema, {
          id: props.connectionId,
          name: values.connectionName,
          connectionConfig: buildConnectionConfigOpenAi(values),
        })
      );
    }
  }

  return (
    <OpenAiForm
      mode={mode}
      initialValues={initialValues}
      onSubmit={mode !== 'view' ? handleSubmit : undefined}
    />
  );
}
