import { Connection } from '@neosync/sdk';

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
  connection: Connection;
  onSuccess(conn: Connection): Promise<void>;
}

export type ConnectionFormProps =
  | CreateProps
  | EditProps
  | ViewProps
  | CloneProps;
