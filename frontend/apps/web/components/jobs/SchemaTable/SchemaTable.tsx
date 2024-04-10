'use client';
import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import DualListBox, {
  Action,
  Option,
} from '@/components/DualListBox/DualListBox';
import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { ConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { useGetTransformersHandler } from '@/libs/hooks/useGetTransformersHandler';
import {
  JobMappingFormValues,
  SchemaFormValues,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { JobMappingTransformer } from '@neosync/sdk';
import {
  CheckCircledIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  TableIcon,
} from '@radix-ui/react-icons';
import { ReactElement, useMemo, useState } from 'react';
import { FieldErrors, useFieldArray, useFormContext } from 'react-hook-form';
import { getSchemaColumns } from './SchemaColumns';
import SchemaPageTable from './SchemaPageTable';
import { JobType, SchemaConstraintHandler } from './schema-constraint-handler';

interface Props {
  data: JobMappingFormValues[];
  jobType: JobType;
  schema: ConnectionSchemaMap;
  isSchemaDataReloading: boolean;
  constraintHandler: SchemaConstraintHandler;
}

export function SchemaTable(props: Props): ReactElement {
  const { data, constraintHandler, jobType, schema } = props;

  const { account } = useAccount();
  const { handler, isLoading, isValidating } = useGetTransformersHandler(
    account?.id ?? ''
  );
  const [selectedItems, setSelectedItems] = useState<Set<string>>(
    new Set(data.map((d) => `${d.schema}.${d.table}`))
  );

  const columns = useMemo(() => {
    return getSchemaColumns({
      transformerHandler: handler,
      constraintHandler,
      jobType,
    });
  }, [handler, constraintHandler, jobType]);

  const form = useFormContext<SchemaFormValues | SingleTableSchemaFormValues>();
  const { append, remove, fields } = useFieldArray<
    SchemaFormValues | SingleTableSchemaFormValues
  >({
    control: form.control,
    name: 'mappings',
  });

  function toggleItem(items: Set<string>, _action: Action): void {
    if (items.size === 0) {
      const idxs = fields.map((_, idx) => idx);
      remove(idxs);
      setSelectedItems(new Set());
      return;
    }
    const [added, removed] = getDelta(items, selectedItems);

    const toRemove: number[] = [];
    const toAdd: any[] = [];

    fields.forEach((field, idx) => {
      if (removed.has(`${field.schema}.${field.table}`)) {
        toRemove.push(idx);
      }
    });
    added.forEach((item) => {
      const dbcols = schema[item];
      if (!dbcols) {
        return;
      }
      dbcols.forEach((dbcol) => {
        toAdd.push({
          schema: dbcol.schema,
          table: dbcol.table,
          column: dbcol.column,
          dataType: dbcol.dataType,
          transformer: convertJobMappingTransformerToForm(
            new JobMappingTransformer({})
          ),
        });
      });
    });
    if (toRemove.length > 0) {
      remove(toRemove);
    }
    if (toAdd.length > 0) {
      append(toAdd);
    }
    setSelectedItems(items);
  }

  if (isLoading || !data) {
    return <SkeletonTable />;
  }

  const extractedFormErrors = formErrorsToMessages(
    extractAllErrors(form.formState.errors, form.getValues('mappings'))
  );
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
              <div>{isValidating ? <Spinner /> : null}</div>
            </div>
            <CardDescription>
              Select the tables that you want to transform and move them from
              the source to destination table.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <DualListBox
              options={getDualListBoxOptions(schema, data)}
              selected={selectedItems}
              onChange={toggleItem}
              mode={jobType === 'generate' ? 'single' : 'many'}
            />
          </CardContent>
        </Card>
        <Card className="w-full flex flex-col">
          <CardHeader className="flex flex-col gap-2">
            <div className="flex flex-row items-center gap-2">
              {extractedFormErrors.length != 0 ? (
                <ExclamationTriangleIcon className="h-4 w-4 text-destructive" />
              ) : (
                <CheckCircledIcon className="w-4 h-4" />
              )}
              <CardTitle>Validations</CardTitle>

              {extractedFormErrors.length != 0 && (
                <Badge variant="destructive">
                  {extractedFormErrors.length} Errors
                </Badge>
              )}
            </div>
            <CardDescription>
              A list of schema validation errors to resolve before moving
              forward.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col flex-1">
            {extractedFormErrors.length === 0 ? (
              <div className="flex flex-col flex-1 items-center justify-center bg-green-100 dark:bg-green-900 text-green-900 dark:text-green-200 rounded-xl">
                <div className="text-sm flex flex-row items-center gap-2 px-1">
                  <div className="flex">
                    <CheckIcon />
                  </div>
                  <p>Everything looks good!</p>
                </div>
              </div>
            ) : (
              <ScrollArea className="max-h-[177px] overflow-auto">
                <div className="flex flex-col gap-2">
                  {extractedFormErrors.map((message, index) => (
                    <div
                      key={message + index}
                      className="text-xs bg-red-200 dark:bg-red-800/70 rounded-sm p-2 text-wrap"
                    >
                      {message}
                    </div>
                  ))}
                </div>
              </ScrollArea>
            )}
          </CardContent>
        </Card>
      </div>
      <SchemaPageTable
        columns={columns}
        data={data}
        transformerHandler={handler}
        constraintHandler={constraintHandler}
        jobType={jobType}
      />
    </div>
  );
}

function getDualListBoxOptions(
  schema: ConnectionSchemaMap,
  jobmappings: JobMappingFormValues[]
): Option[] {
  const tables = new Set(Object.keys(schema));
  jobmappings.forEach((jm) => tables.add(`${jm.schema}.${jm.table}`));
  return Array.from(tables).map((table): Option => ({ value: table }));
}

interface FormError {
  message: string;
  type?: string;
  path: string;
}

function extractAllErrors(
  errors: FieldErrors<SchemaFormValues | SingleTableSchemaFormValues>,
  values: JobMappingFormValues[],
  path = ''
): FormError[] {
  let messages: FormError[] = [];

  for (const key in errors) {
    let newPath = path;

    if (!isNaN(Number(key))) {
      const index = Number(key);
      if (index < values.length) {
        const value = values[index];
        const column = `${value.schema}.${value.table}.${value.column}`;
        newPath = path ? `${path}.${column}` : column;
      }
    }
    const error = (errors as any)[key as unknown as any] as any;

    if (!error) {
      continue;
    }
    if (error.message) {
      messages.push({
        path: newPath,
        message: error.message,
        type: error.type,
      });
    } else {
      messages = messages.concat(extractAllErrors(error, values, newPath));
    }
  }
  return messages;
}

function formErrorsToMessages(errors: FormError[]): string[] {
  const messages: string[] = [];
  errors.forEach((error) => {
    const pieces: string[] = [error.path];
    if (error.type) {
      pieces.push(`[${error.type}]`);
    }
    pieces.push(error.message);
    messages.push(pieces.join(' '));
  });

  return messages;
}

function getDelta(
  newSet: Set<string>,
  oldSet: Set<string>
): [Set<string>, Set<string>] {
  const added = new Set<string>();
  const removed = new Set<string>();

  oldSet.forEach((val) => {
    if (!newSet.has(val)) {
      removed.add(val);
    }
  });
  newSet.forEach((val) => {
    if (!oldSet.has(val)) {
      added.add(val);
    }
  });

  return [added, removed];
}
