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
    <Table className="rounded-lg border border-gray-400 dark:border-gray-70 overflow-hidden">
      <TableHeader>
        <TableRow>
          {headers.map((header, headerIndex) => (
            <TableHead key={`header-${headerIndex}`}>{header}</TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {rowData.map((tableData, rowIndex) => (
          <TableRow key={`row-${rowIndex}-${tableData.data[0]}`}>
            {tableData.data.map((item, columnIndex) => (
              <TableCell
                key={`cell-${rowIndex}-${columnIndex}`}
                className="font-medium bg-[#FFFFFF] dark:bg-[#1c1c1c] "
              >
                {item}
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
