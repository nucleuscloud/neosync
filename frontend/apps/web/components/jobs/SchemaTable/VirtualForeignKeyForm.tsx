import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { toast } from '@/components/ui/use-toast';
import { ConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { VirtualForeignConstraintFormValues } from '@/yup-validations/jobs';
import { ReactElement, useState } from 'react';
import { GoWorkflow } from 'react-icons/go';
import { StringSelect } from './StringSelect';

interface Props {
  schema: ConnectionSchemaMap;
  virtualForeignKey?: VirtualForeignConstraintFormValues;
  selectedTables: Set<string>;
  addVirtualForeignKey: (vfk: VirtualForeignConstraintFormValues) => void;
}

export function VirtualForeignKeyForm(props: Props): ReactElement {
  const { schema, selectedTables, virtualForeignKey, addVirtualForeignKey } =
    props;

  const srcTable = virtualForeignKey
    ? `${virtualForeignKey?.schema}-${virtualForeignKey?.table}`
    : undefined;
  const srcCol = virtualForeignKey?.columns && virtualForeignKey?.columns[0];
  const [sourceTable, setSourceTable] = useState<string | undefined>(srcTable);
  const [sourceColumn, setSourceColumn] = useState<string | undefined>(srcCol);

  const tarTable = virtualForeignKey
    ? `${virtualForeignKey?.foreignKey?.schema}-${virtualForeignKey?.foreignKey?.table}`
    : undefined;
  const tarCol =
    virtualForeignKey?.foreignKey.columns &&
    virtualForeignKey?.foreignKey.columns[0];
  const [targetTable, setTargetTable] = useState<string | undefined>(tarTable);
  const [targetColumn, setTargetColumn] = useState<string | undefined>(tarCol);

  return (
    <div className="flex flex-col md:flex-row gap-3">
      <Card className="w-full">
        <CardHeader className="flex flex-col gap-2">
          <div className="flex flex-row items-center gap-2">
            <div className="flex">
              <GoWorkflow className="h-4 w-4" />
            </div>
            <CardTitle>Add Virtual Foreign Key</CardTitle>
          </div>
          <CardDescription>
            Select the tables that you want to transform and move them from the
            source to destination table.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col md:flex-row gap-6">
            <div className="flex flex-col gap-2">
              <FormLabel>Source</FormLabel>
              <FormDescription>The primary key</FormDescription>
              <div className="flex flex-col md:flex-row gap-3">
                <StringSelect
                  value={sourceTable}
                  values={Array.from(selectedTables)}
                  setValue={setSourceTable}
                  text="table"
                />
                <StringSelect
                  value={sourceColumn}
                  values={getColumnOptions(schema, sourceTable)}
                  setValue={setSourceColumn}
                  text="column"
                />
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <FormLabel>Target</FormLabel>
              <FormDescription>The foreign key</FormDescription>
              <div className="flex flex-col md:flex-row gap-3">
                <StringSelect
                  value={targetTable}
                  values={Array.from(selectedTables)}
                  setValue={setTargetTable}
                  text="table"
                />
                <StringSelect
                  value={targetColumn}
                  values={getColumnOptions(schema, targetTable)}
                  setValue={setTargetColumn}
                  text="column"
                />
                <Button
                  type="button"
                  key="virtualforeignkey"
                  onClick={() => {
                    if (
                      !sourceTable ||
                      !sourceColumn ||
                      !targetTable ||
                      !targetColumn
                    ) {
                      // add alert toast. missing required values
                      toast({
                        title: 'Unable to add virtual foreign key',
                        description: 'Missing required field',
                        variant: 'destructive',
                      });
                      return;
                    }
                    const source = splitSchemaTable(sourceTable);
                    const target = splitSchemaTable(targetTable);
                    addVirtualForeignKey({
                      schema: source.schema,
                      table: source.table,
                      columns: [sourceColumn],
                      foreignKey: {
                        schema: target.schema,
                        table: target.table,
                        columns: [targetColumn],
                      },
                    });
                  }}
                >
                  Add
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function getColumnOptions(
  schema: ConnectionSchemaMap,
  table?: string
): string[] {
  if (!table) {
    return [];
  }
  const columns = new Set<string>();
  const cols = schema[table];
  cols.forEach((c) => {
    // need to filter out any non unique cols
    columns.add(c.column);
  });
  return Array.from(columns);
}

function splitSchemaTable(input: string): { schema: string; table: string } {
  const [schema, table] = input.split('.');
  if (!schema || !table) {
    throw new Error("Input must be in the form 'schema.table'");
  }
  return { schema, table };
}
