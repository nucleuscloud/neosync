import {
  FormDescription,
  FormField,
  FormItem,
  FormLabel
} from '@/components/ui/form';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { ReactElement } from 'react';
import { SampleRecord } from '../types';

interface Props {
  columns: ColumnDef<SampleRecord>[];
  records: SampleRecord[];
}
export default function SampleTable(props: Props): ReactElement {
  const { records, columns } = props;

  const table = useReactTable({
    data: records,
    columns: columns,
    enableRowSelection: false,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <FormField
      name="syntheticDataSample"
      render={() => (
        <FormItem>
          <FormLabel>Synthetic Data Sample</FormLabel>
          <FormDescription>
            A sample of synthetic data given the inputs above. Returns at most
             10 records.
          </FormDescription>
          <div className="rounded-md border overflow-auto dark:border-gray-700">
            <Table>
              <TableHeader className="bg-gray-100 dark:bg-gray-800">
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead key={header.id} className="pl-2">
                        {header.isPlaceholder
                          ? null
                          : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      data-state={row.getIsSelected() && 'selected'}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id}>
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext()
                          )}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell
                      colSpan={columns.length}
                      className="h-24 text-center"
                    >
                      Click Sample to generate a preview of synthetic data.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </FormItem>
      )}
    />
  );
}
