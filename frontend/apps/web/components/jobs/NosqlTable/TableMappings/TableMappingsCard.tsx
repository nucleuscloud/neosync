import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { EditDestinationOptionsFormValues } from '@/yup-validations/jobs';
import { TableIcon } from '@radix-ui/react-icons';
import { ReactElement, useMemo } from 'react';
import {
  DestinationDetails,
  getTableMappingsColumns,
  OnTableMappingUpdateRequest,
  TableMappingRow,
} from './Columns';
import TableMappingsTable from './TableMappingsTable';

export interface Props {
  mappings: EditDestinationOptionsFormValues[];
  onUpdate(req: OnTableMappingUpdateRequest): void;
  destinationDetailsRecord: Record<string, DestinationDetails>;
}

export default function TableMappingsCard(props: Props): ReactElement {
  const { mappings, onUpdate, destinationDetailsRecord } = props;
  const columns = useMemo(
    () => getTableMappingsColumns({ destinationDetailsRecord, onUpdate }),
    [destinationDetailsRecord, onUpdate]
  );
  return (
    <Card className="w-full">
      <CardHeader className="flex flex-col gap-2">
        <div className="flex flex-row items-center gap-2">
          <div className="flex">
            <TableIcon className="h-4 w-4" />
          </div>
          <CardTitle>DynamoDB Table Mappings</CardTitle>
        </div>
        <CardDescription className="max-w-2xl">
          Map a table from source to destination. As tables are added in the
          form above, they will dynamically be added to this section. A mapping
          is required to denote which table each source table should be synced
          to for each corresponding DynamoDB destination.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <TableMappingsTable
          data={toTableMappingRows(mappings)}
          columns={columns}
        />
      </CardContent>
    </Card>
  );
}

function toTableMappingRows(
  mappings: EditDestinationOptionsFormValues[]
): TableMappingRow[] {
  return mappings.flatMap((mapping) => {
    return (
      mapping.dynamodb?.tableMappings.map((tm): TableMappingRow => {
        return {
          destinationId: mapping.destinationId,
          sourceTable: tm.sourceTable,
          destinationTable: tm.destinationTable,
        };
      }) ?? []
    );
  });
}
