import { create as createMessage } from '@bufbuild/protobuf';
import {
  ConnectionConfig,
  CreateConnectionRequest,
  CreateConnectionRequestSchema,
  UpdateConnectionRequest,
  UpdateConnectionRequestSchema,
} from '@neosync/sdk';
import { ComponentType, ReactElement } from 'react';

export interface BaseProps<T> {
  initialValues?: T;
  ConnectionForm: ComponentType<{
    mode: 'create' | 'edit' | 'view';
    initialValues?: T;
    onSubmit?(values: T): Promise<void>;
  }>;
}

export interface CreateProps<T> extends BaseProps<T> {
  mode: 'create';
  accountId: string;
  onSubmit(values: CreateConnectionRequest): Promise<void>;
  buildConnectionConfig(formValues: T): ConnectionConfig;
}

export interface EditProps<T> extends BaseProps<T> {
  mode: 'edit';
  connectionId: string;
  onSubmit(values: UpdateConnectionRequest): Promise<void>;
  buildConnectionConfig(formValues: T): ConnectionConfig;
}

export interface ViewProps<T> extends BaseProps<T> {
  mode: 'view';
}

type Props<T> = CreateProps<T> | EditProps<T> | ViewProps<T>;

export default function ConnectionForm<T extends { connectionName: string }>(
  props: Props<T>
): ReactElement {
  const { mode, initialValues, ConnectionForm } = props;

  async function handleSubmit(values: T): Promise<void> {
    if (mode === 'view') {
      return;
    }

    if (mode === 'create') {
      await props.onSubmit(
        createMessage(CreateConnectionRequestSchema, {
          accountId: props.accountId,
          name: values.connectionName,
          connectionConfig: props.buildConnectionConfig(values),
        })
      );
    } else {
      await props.onSubmit(
        createMessage(UpdateConnectionRequestSchema, {
          id: props.connectionId,
          name: values.connectionName,
          connectionConfig: props.buildConnectionConfig(values),
        })
      );
    }
  }

  return (
    <ConnectionForm
      mode={mode}
      initialValues={initialValues}
      onSubmit={mode !== 'view' ? handleSubmit : undefined}
    />
  );
}
