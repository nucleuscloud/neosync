import Link from '@docusaurus/Link';
import { ArrowRightIcon } from '@radix-ui/react-icons';
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
  title: string;
  type: string;
  description: string;
  link: string;
}

export function TransformersTable(props: TableData): ReactElement {
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
          <TableRow key={tableData.title} className="h2">
            <TableCell className="font-medium bg-[#FFFFFF]">
              <Link
                href={`/transformers/system/reference#${tableData.link}`}
                className=""
              >
                <div className="flex flex-row gap-2">
                  <div>{tableData.title}</div>
                  <div className="flex justify-end transition-transform duration-300 transform group-hover:translate-x-[4px]">
                    <ArrowRightIcon />
                  </div>
                </div>
              </Link>
            </TableCell>
            <TableCell
              key={tableData.type}
              className="font-medium bg-[#FFFFFF]"
            >
              {tableData.type}
            </TableCell>
            <TableCell
              key={tableData.description}
              className="font-medium bg-[#FFFFFF]"
            >
              {tableData.description}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
