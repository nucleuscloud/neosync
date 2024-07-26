import { ColumnDef } from '@tanstack/react-table';
import { StringSelect } from '../../SchemaTable/StringSelect';
import { ColumnHeader } from './ColumnHeader';

interface Props {
  destinationDetailsRecord: Record<string, DestinationDetails>;
}

export interface DestinationDetails {
  destinationId: string;
  friendlyName: string;

  availableTableNames: string[];
}

export interface TableMappingRow {
  destinationId: string;
  sourceTable: string;
  destinationTable: string;
}

export function getTableMappingsColumns(
  props: Props
): ColumnDef<TableMappingRow>[] {
  const { destinationDetailsRecord } = props;

  return [
    {
      accessorKey: 'destinationId',
      header: ({ column }) => (
        <ColumnHeader column={column} title="Destination" />
      ),
      cell: ({ getValue }) => {
        const destId = getValue<string>();
        const details: DestinationDetails = destinationDetailsRecord[
          destId
        ] ?? {
          destinationId: destId,
          friendlyName: 'Unknown Name',
          availableTableNames: [],
        };
        return <span>{details.friendlyName}</span>;
      },
    },
    {
      accessorKey: 'sourceTable',
      header: ({ column }) => (
        <ColumnHeader column={column} title="Source Table" />
      ),
    },
    {
      accessorKey: 'destinationTable',
      header: ({ column }) => (
        <ColumnHeader column={column} title="Destination Table" />
      ),
      cell: ({ row, getValue }) => {
        const destTableName = getValue<string>();
        const destId = row.getValue<string>('destinationId');
        const details: DestinationDetails = destinationDetailsRecord[
          destId
        ] ?? {
          destinationId: destId,
          friendlyName: 'Unknown Name',
          availableTableNames: [destTableName],
        };

        return (
          <div className="flex flex-row gap-2 items-center">
            <StringSelect
              value={destTableName}
              values={details.availableTableNames}
              setValue={() => undefined}
              text="destination table"
            />
          </div>
        );
      },
    },
  ];
}
