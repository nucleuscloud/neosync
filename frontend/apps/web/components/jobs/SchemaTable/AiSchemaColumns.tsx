import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ColumnDef } from '@tanstack/react-table';
import { SchemaColumnHeader } from './SchemaColumnHeader';
import { SchemaConstraintHandler } from './schema-constraint-handler';
import { handleDataTypeBadge, toColKey } from './util';

interface Props {
  constraintHandler: SchemaConstraintHandler;
}

interface Row {
  schema: string;
  table: string;
  column: string;
}

export function getSchemaColumns(props: Props): ColumnDef<Row>[] {
  const { constraintHandler } = props;

  return [
    {
      accessorKey: 'schema',
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'table',
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorFn: (row) => `${row.schema}.${row.table}`,
      id: 'schemaTable',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Table" />
      ),
      cell: ({ getValue }) => {
        return (
          <div className="flex flex-row gap-2 items-center">
            <span className="max-w-[500px] truncate font-medium">
              {getValue<string>()}
            </span>
          </div>
        );
      },
      maxSize: 500,
      size: 300,
    },
    {
      accessorKey: 'column',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Column" />
      ),
      cell: ({ row }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {row.getValue('column')}
          </span>
        );
      },
      maxSize: 500,
      size: 200,
    },
    {
      id: 'constraints',
      accessorKey: 'constraints',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Constraints" />
      ),
      accessorFn: (row) => {
        const key = toColKey(row.schema, row.table, row.column);
        const isPrimaryKey = constraintHandler.getIsPrimaryKey(key);
        const [isForeignKey, fkCols] = constraintHandler.getIsForeignKey(key);

        const pieces: string[] = [];
        if (isPrimaryKey) {
          pieces.push('Primary Key');
        }
        if (isForeignKey) {
          fkCols.forEach((col) => pieces.push(`Foreign Key: ${col}`));
        }
        return pieces.join('\n');
      },
      cell: ({ row }) => {
        const key = {
          schema: row.getValue<string>('schema'),
          table: row.getValue<string>('table'),
          column: row.getValue<string>('column'),
        };
        const isPrimaryKey = constraintHandler.getIsPrimaryKey(key);
        const [isForeignKey, fkCols] = constraintHandler.getIsForeignKey(key);
        const isUniqueConstraint = constraintHandler.getIsUniqueConstraint(key);
        return (
          <span className="max-w-[500px] truncate font-medium">
            <div className="flex flex-col lg:flex-row items-start gap-1">
              <div>
                {isPrimaryKey && (
                  <Badge
                    variant="outline"
                    className="text-xs bg-blue-100 text-gray-800 cursor-default dark:bg-blue-200 dark:text-gray-900"
                  >
                    Primary Key
                  </Badge>
                )}
              </div>
              <div>
                {isForeignKey && (
                  <TooltipProvider>
                    <Tooltip delayDuration={200}>
                      <TooltipTrigger>
                        <Badge
                          variant="outline"
                          className="text-xs bg-orange-100 text-gray-800 cursor-default dark:dark:text-gray-900 dark:bg-orange-300"
                        >
                          Foreign Key
                        </Badge>
                      </TooltipTrigger>
                      <TooltipContent>
                        {fkCols.map((col) => `Primary Key: ${col}`).join('\n')}
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                )}
              </div>
              <div>
                {isUniqueConstraint && (
                  <Badge
                    variant="outline"
                    className="text-xs bg-purple-100 text-gray-800 cursor-default dark:bg-purple-300 dark:text-gray-900"
                  >
                    Unique
                  </Badge>
                )}
              </div>
            </div>
          </span>
        );
      },
    },
    {
      accessorKey: 'isNullable',
      accessorFn: (row) => {
        const key = toColKey(row.schema, row.table, row.column);
        return constraintHandler.getIsNullable(key) ? 'Yes' : 'No';
      },
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Nullable" />
      ),
      cell: ({ row }) => {
        const key = {
          schema: row.getValue<string>('schema'),
          table: row.getValue<string>('table'),
          column: row.getValue<string>('column'),
        };
        const isNullable = constraintHandler.getIsNullable(key);
        const text = isNullable ? 'Yes' : 'No';
        return (
          <span className="max-w-[500px] truncate font-medium">
            <Badge variant="outline">{text}</Badge>
          </span>
        );
      },
    },
    {
      accessorKey: 'dataType',
      accessorFn: (row) => {
        const key = toColKey(row.schema, row.table, row.column);
        return handleDataTypeBadge(constraintHandler.getDataType(key));
      },
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Data Type" />
      ),
      cell: ({ row }) => {
        const key = {
          schema: row.getValue<string>('schema'),
          table: row.getValue<string>('table'),
          column: row.getValue<string>('column'),
        };
        const datatype = constraintHandler.getDataType(key);
        return (
          <TooltipProvider>
            <Tooltip delayDuration={200}>
              <TooltipTrigger asChild>
                <div>
                  <Badge variant="outline" className="max-w-[100px]">
                    <span className="truncate block overflow-hidden">
                      {handleDataTypeBadge(datatype)}
                    </span>
                  </Badge>
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <p>{datatype}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        );
      },
    },
  ];
}
