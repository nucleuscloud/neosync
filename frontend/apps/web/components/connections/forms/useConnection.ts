import { useAccount } from '@/components/providers/account-provider';
import { create as createMessage } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  Connection,
  ConnectionConfig,
  ConnectionService,
  CreateConnectionRequestSchema,
  UpdateConnectionRequestSchema,
} from '@neosync/sdk';
import { useEffect, useState } from 'react';

export interface CreateProps<T> {
  mode: 'create';
  buildConnectionConfig(formValues: T): ConnectionConfig;
  onSuccess(conn: Connection): Promise<void> | void;
}

export interface EditProps<T> {
  mode: 'edit';
  connection: Connection;
  buildConnectionConfig(formValues: T): ConnectionConfig;
  toFormValues(connection: Connection): T | undefined;
  onSuccess(conn: Connection): Promise<void> | void;
}

export interface ViewProps<T> {
  mode: 'view';
  connection: Connection;

  canViewSecrets?: boolean;
  toFormValues(connection: Connection): T | undefined;
}

export interface CloneProps<T> {
  mode: 'clone';
  connectionId: string;
  toFormValues(connection: Connection): T | undefined;
  buildConnectionConfig(formValues: T): ConnectionConfig;
  onSuccess(conn: Connection): Promise<void> | void;
}

interface UseConnectionResult<T> {
  isLoading: boolean;
  initialValues: T | undefined;
  handleSubmit(values: T): Promise<void>;
  getValueWithSecrets(): Promise<T | undefined>;
}

type Props<T> = CreateProps<T> | EditProps<T> | ViewProps<T> | CloneProps<T>;
export function useConnection<T extends { connectionName: string }>(
  props: Props<T>
): UseConnectionResult<T> {
  const { mode } = props;

  const { account } = useAccount();

  const { mutateAsync: getConnection } = useMutation(
    ConnectionService.method.getConnection
  );
  const { mutateAsync: createConnection } = useMutation(
    ConnectionService.method.createConnection
  );
  const { mutateAsync: updateConnection } = useMutation(
    ConnectionService.method.updateConnection
  );
  const [isLoading, setIsLoading] = useState(true);
  const [initialValues, setInitialValues] = useState<T | undefined>(undefined);

  useEffect(() => {
    async function loadValues(): Promise<void> {
      setIsLoading(true);
      try {
        if (mode === 'create') {
          setInitialValues(undefined);
          return;
        }

        if (mode === 'clone') {
          const connectionResp = await getConnection({
            id: props.connectionId,
            excludeSensitive: false,
          });
          if (connectionResp.connection) {
            setInitialValues(props.toFormValues(connectionResp.connection));
          }
          return;
        }

        // For view/edit modes
        const values = props.toFormValues(props.connection);
        if (mode === 'view' || initialValues === undefined) {
          setInitialValues(values);
        }
      } finally {
        setIsLoading(false);
      }
    }
    loadValues();
  }, [
    mode,
    mode === 'view' ? props.connection : undefined,
    mode === 'clone' ? props.connectionId : undefined,
  ]);

  async function handleSubmit(values: T): Promise<void> {
    if (mode === 'view' || !account?.id) {
      return;
    }

    if (mode === 'create' || mode === 'clone') {
      const newConnResp = await createConnection(
        createMessage(CreateConnectionRequestSchema, {
          accountId: account.id,
          name: values.connectionName,
          connectionConfig: props.buildConnectionConfig(values),
        })
      );
      if (newConnResp.connection) {
        props.onSuccess(newConnResp.connection);
      }
    } else if (mode === 'edit') {
      const updatedConnResp = await updateConnection(
        createMessage(UpdateConnectionRequestSchema, {
          id: props.connection.id,
          name: values.connectionName,
          connectionConfig: props.buildConnectionConfig(values),
        })
      );
      if (updatedConnResp.connection) {
        props.onSuccess(updatedConnResp.connection);
      }
    }
  }

  async function getValueWithSecrets(): Promise<T | undefined> {
    if (mode !== 'view') {
      return undefined;
    }

    // todo: handle errors?
    const connectionResp = await getConnection({
      id: props.connection.id,
      excludeSensitive: false,
    });
    if (connectionResp.connection) {
      const formValues = props.toFormValues(connectionResp.connection);
      if (formValues) {
        return formValues;
      }
    }
    return undefined;
  }

  return {
    isLoading,
    initialValues,
    handleSubmit,
    getValueWithSecrets,
  };
}
