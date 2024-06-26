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
import { SchemaConstraintHandler } from './schema-constraint-handler';

interface Props {
  schema: ConnectionSchemaMap;
  virtualForeignKey?: VirtualForeignConstraintFormValues;
  selectedTables: Set<string>;
  addVirtualForeignKey: (vfk: VirtualForeignConstraintFormValues) => void;
  constraintHandler: SchemaConstraintHandler;
}

export function VirtualForeignKeyForm(props: Props): ReactElement {
  const {
    schema,
    selectedTables,
    virtualForeignKey,
    constraintHandler,
    addVirtualForeignKey,
  } = props;

  const srcTable = virtualForeignKey
    ? `${virtualForeignKey?.schema}-${virtualForeignKey?.table}`
    : undefined;
  const srcCols = virtualForeignKey?.columns || [''];
  const [sourceTable, setSourceTable] = useState<string | undefined>(srcTable);
  const [sourceColumns, setSourceColumns] = useState<string[]>(srcCols);

  const tarTable = virtualForeignKey
    ? `${virtualForeignKey?.foreignKey?.schema}-${virtualForeignKey?.foreignKey?.table}`
    : undefined;
  const tarCols = virtualForeignKey?.foreignKey.columns || [''];
  const [targetTable, setTargetTable] = useState<string | undefined>(tarTable);
  const [targetColumns, setTargetColumns] = useState<string[]>(tarCols);

  const addCompositeColumns = () => {
    setSourceColumns([...sourceColumns, '']);
    setTargetColumns([...targetColumns, '']);
  };

  const removeLastCompositeColumn = () => {
    if (sourceColumns.length == 1) {
      return;
    }
    const newSourceColumns = [...sourceColumns];
    newSourceColumns.pop();
    setSourceColumns(newSourceColumns);
    const newTargetColumns = [...targetColumns];
    newTargetColumns.pop();
    setTargetColumns(newTargetColumns);
  };

  const updateSourceColumn = (index: number, value: string) => {
    const newSourceColumns = [...sourceColumns];
    newSourceColumns[index] = value;
    setSourceColumns(newSourceColumns);
  };

  const updateTargetColumn = (index: number, value: string) => {
    const newTargetColumns = [...targetColumns];
    newTargetColumns[index] = value;
    setTargetColumns(newTargetColumns);
  };

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
            Select the source table and columns, as well as the target table and
            columns, to create a virtual foreign key. <br /> Add additional
            columns to create a composite virtual foreign key.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-2">
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
                  <div className="flex flex-col gap-3">
                    {sourceColumns.map((col, index) => (
                      <StringSelect
                        key={index}
                        value={col}
                        values={getSourceColumnOptions(
                          schema,
                          constraintHandler,
                          sourceTable
                        )}
                        badgeValueMap={getColumnDataTypeMap(
                          schema,
                          sourceTable
                        )}
                        setValue={(value) => updateSourceColumn(index, value)}
                        text="column"
                      />
                    ))}
                  </div>
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
                  <div className="flex flex-col gap-3">
                    {targetColumns.map((col, index) => (
                      <StringSelect
                        key={index}
                        value={col}
                        values={getTargetColumnOptions(schema, targetTable)}
                        badgeValueMap={getColumnDataTypeMap(
                          schema,
                          targetTable
                        )}
                        setValue={(value) => updateTargetColumn(index, value)}
                        text="column"
                      />
                    ))}
                  </div>
                  <div className="flex flex-row gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={addCompositeColumns}
                    >
                      +
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={removeLastCompositeColumn}
                    >
                      -
                    </Button>
                  </div>
                </div>
                <div className="flex justify-end">
                  <Button
                    type="button"
                    key="virtualforeignkey"
                    className="w-[90px]"
                    onClick={() => {
                      if (
                        !sourceTable ||
                        sourceColumns.includes('') ||
                        !targetTable ||
                        targetColumns.includes('')
                      ) {
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
                        schema: target.schema,
                        table: target.table,
                        columns: targetColumns,
                        foreignKey: {
                          schema: source.schema,
                          table: source.table,
                          columns: sourceColumns,
                        },
                      });
                      setSourceTable(undefined);
                      setSourceColumns(['']);
                      setTargetTable(undefined);
                      setTargetColumns(['']);
                    }}
                  >
                    Add
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function getColumnDataTypeMap(
  schema: ConnectionSchemaMap,
  table?: string
): Record<string, string> {
  const results: Record<string, string> = {};
  if (!table) {
    return results;
  }
  const columns = schema[table];
  columns.forEach((c) => {
    results[c.column] = c.dataType;
  });
  return results;
}

function getSourceColumnOptions(
  schema: ConnectionSchemaMap,
  constraintHandler: SchemaConstraintHandler,
  table?: string
): string[] {
  if (!table) {
    return [];
  }
  const columns = new Set<string>();
  const cols = schema[table];
  cols.forEach((c) => {
    const colkey = {
      schema: c.schema,
      table: c.table,
      column: c.column,
    };

    if (constraintHandler.getIsNullable(colkey)) {
      return;
    }

    const isUnique = constraintHandler.getIsUniqueConstraint(colkey);
    const isPrimary = constraintHandler.getIsPrimaryKey(colkey);

    if (isUnique || isPrimary) {
      columns.add(c.column);
    }
  });
  return Array.from(columns);
}

function getTargetColumnOptions(
  schema: ConnectionSchemaMap,
  table?: string
): string[] {
  if (!table) {
    return [];
  }
  const columns = new Set<string>();
  const cols = schema[table];
  cols.forEach((c) => {
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
