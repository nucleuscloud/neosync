'use client';
import DualListBox, {
  Action,
  Option,
} from '@/components/DualListBox/DualListBox';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { GetConnectionSchemaResponse } from '@neosync/sdk';
import { TableIcon } from '@radix-ui/react-icons';
import { ReactElement, useMemo } from 'react';
import { getSchemaColumns } from './AiSchemaColumns';
import AiSchemaPageTable from './AiSchemaPageTable';
import { SchemaConstraintHandler } from './schema-constraint-handler';

export interface AiSchemaTableRecord {
  schema: string;
  table: string;
  column: string;
}

interface Props {
  data: AiSchemaTableRecord[];
  schema: Record<string, GetConnectionSchemaResponse>;
  isSchemaDataReloading: boolean;
  constraintHandler: SchemaConstraintHandler;

  selectedTables: Set<string>;
  onSelectedTableToggle(items: Set<string>, action: Action): void;
}

export function AiSchemaTable(props: Props): ReactElement<any> {
  const {
    data,
    constraintHandler,
    schema,
    selectedTables,
    onSelectedTableToggle,
  } = props;

  const columns = useMemo(() => {
    return getSchemaColumns({
      constraintHandler,
    });
  }, [constraintHandler]);

  if (!data) {
    return <SkeletonTable />;
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col md:flex-row gap-3">
        <Card className="w-full">
          <CardHeader className="flex flex-col gap-2">
            <div className="flex flex-row items-center gap-2">
              <div className="flex">
                <TableIcon className="h-4 w-4" />
              </div>
              <CardTitle>Table Selection</CardTitle>
            </div>
            <CardDescription>
              Select the table to be used for data generation.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <DualListBox
              options={getDualListBoxOptions(
                new Set(Object.keys(schema)),
                data
              )}
              selected={selectedTables}
              onChange={onSelectedTableToggle}
              mode={'single'}
            />
          </CardContent>
        </Card>
      </div>
      <AiSchemaPageTable columns={columns} data={data} />
    </div>
  );
}

function getDualListBoxOptions(
  tables: Set<string>,
  formValues: { schema: string; table: string; column: string }[]
): Option[] {
  formValues.forEach((jm) => tables.add(`${jm.schema}.${jm.table}`));
  return Array.from(tables).map((table): Option => ({ value: table }));
}
