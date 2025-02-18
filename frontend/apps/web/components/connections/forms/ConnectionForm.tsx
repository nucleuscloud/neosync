import { create as createMessage } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  Connection,
  ConnectionConfig,
  ConnectionService,
  CreateConnectionRequest,
  CreateConnectionRequestSchema,
  UpdateConnectionRequest,
  UpdateConnectionRequestSchema,
} from '@neosync/sdk';
import { ComponentType, ReactElement } from 'react';

export interface BaseProps<T> {
  initialValues?: T;
  Form: ComponentType<{
    mode: 'create' | 'edit' | 'view';
    initialValues?: T;
    onSubmit?(values: T): Promise<void>;
    canViewSecrets?: boolean;
    getValueWithSecrets?(): Promise<T>;
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
  connectionId: string;

  canViewSecrets?: boolean;
  getValueWithSecrets(): Promise<T>;
  toFormValues(connection: Connection): T | undefined;
}

type Props<T> = CreateProps<T> | EditProps<T> | ViewProps<T>;

export default function ConnectionForm<T extends { connectionName: string }>(
  props: Props<T>
): ReactElement {
  const { mode, initialValues, Form: ConnectionForm } = props;

  const { mutateAsync: getConnection } = useMutation(
    ConnectionService.method.getConnection
  );

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

  async function getValueWithSecrets(): Promise<T> {
    if (mode !== 'view') {
      return initialValues as T;
    }

    // todo: handle errors?
    const connectionResp = await getConnection({
      id: props.connectionId,
      excludeSensitive: false,
    });
    if (connectionResp.connection) {
      const formValues = props.toFormValues(connectionResp.connection);
      if (formValues) {
        return formValues;
      }
    }
    return initialValues as T;
  }

  if (mode === 'view') {
    return (
      <ConnectionForm
        mode={mode}
        initialValues={initialValues}
        canViewSecrets={props.canViewSecrets}
        getValueWithSecrets={getValueWithSecrets}
      />
    );
  }

  return (
    <ConnectionForm
      mode={mode}
      initialValues={initialValues}
      onSubmit={handleSubmit}
    />
  );
}
