'use client';
import { ReactElement } from 'react';
import { TableRow, getColumns } from './column';
import { DataTable } from './data-table';

interface Props {
  data: TableRow[];
  onEdit(schema: string, table: string): void;
  hasLocalChange(schema: string, table: string): boolean;
  onReset(schema: string, table: string): void;
}

export default function SubsetTable(props: Props): ReactElement {
  const { data, onEdit, hasLocalChange, onReset } = props;

  const columns = getColumns({ onEdit, hasLocalChange, onReset });

  return <DataTable columns={columns} data={data} />;
}
