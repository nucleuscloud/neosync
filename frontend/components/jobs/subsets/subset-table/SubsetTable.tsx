import { ReactElement } from 'react';
import { TableRow, getColumns } from './column';
import { DataTable } from './data-table';

interface Props {
  data: TableRow[];
  onEdit(schema: string, table: string): void;
}

export default function SubsetTable(props: Props): ReactElement {
  const { data, onEdit } = props;

  const columns = getColumns({ onEdit });

  return <DataTable columns={columns} data={data} />;
}
