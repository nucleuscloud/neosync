import React, { ReactElement } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../components/CustomTable';

interface TableData {
  headers: string[];
  rowData: RowData[];
}

interface RowData {
  data: string[];
}

export function DocsTable(props: TableData): ReactElement {
  const { headers, rowData } = props;

  return (
    <Table className="rounded-lg overflow-hidden border border-gray-400">
      <TableHeader>
        <TableRow>
          {headers.map((header) => (
            <TableHead key={header}>{header}</TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {rowData.map((tableData) => (
          <TableRow key={tableData.data[0]}>
            {tableData.data.map((item) => (
              <TableCell key={item} className="font-medium bg-[#FFFFFF]">
                {item}
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
