'use client';
import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import DualListBox, {
  Action,
  Option,
} from '@/components/DualListBox/DualListBox';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { ConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { useGetMergedTransformers } from '@/libs/hooks/useGetMergedTransformers';
import { cn } from '@/libs/utils';
import {
  JobMappingFormValues,
  SchemaFormValues,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import {
  GetConnectionSchemaResponse,
  JobMappingTransformer,
} from '@neosync/sdk';
import { ReactElement, useMemo, useState } from 'react';
import { FieldErrors, useFieldArray, useFormContext } from 'react-hook-form';
import { SchemaConstraintHandler, getSchemaColumns } from './SchemaColumns';
import SchemaPageTable from './SchemaPageTable';

interface Props {
  data: JobMappingFormValues[];
  excludeInputReqTransformers?: boolean; // will result in only generators (functions with no data input)
  jobType: string; // todo: update to be named type
  schema: ConnectionSchemaMap;
  constraintHandler: SchemaConstraintHandler;
}

export function SchemaTable(props: Props): ReactElement {
  const {
    data,
    excludeInputReqTransformers,
    constraintHandler,
    jobType,
    schema,
  } = props;

  const { account } = useAccount();
  const {
    systemTransformers,
    userDefinedTransformers,
    userDefinedMap,
    systemMap,
    isLoading,
    isValidating,
  } = useGetMergedTransformers(
    excludeInputReqTransformers ?? false,
    account?.id ?? ''
  );
  const [selectedItems, setSelectedItems] = useState<Set<string>>(
    new Set(data.map((d) => `${d.schema}.${d.table}`))
  );

  const columns = useMemo(() => {
    return getSchemaColumns({
      systemTransformers,
      userDefinedTransformers,
      systemMap,
      userDefinedMap,
      constraintHandler,
    });
  }, [isValidating, constraintHandler]);

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
    extractAllErrors(form.formState.errors)
  );

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col md:flex-row gap-3 items-start">
        <div className="flex">
          <Card className="p-0">
            <CardHeader className="p-3 pb-0">
              <CardTitle>Select the tables this job should map</CardTitle>
              <CardDescription className="max-w-[350px]">
                Once selected, choose the desired transformations for each
                column below.
              </CardDescription>
            </CardHeader>
            <CardContent className="p-3">
              <DualListBox
                options={getDualListBoxOptions(schema, data)}
                selected={selectedItems}
                onChange={toggleItem}
                title="Table"
              />
            </CardContent>
          </Card>
        </div>
        <div className="flex">
          <Card
            className={cn(
              'p-0',
              extractedFormErrors.length === 0 ? 'hidden' : ''
            )}
          >
            <CardHeader className="p-3 pb-0">
              <CardTitle>Validation Errors</CardTitle>
              <CardDescription className="max-w-[350px]">
                Validation errors must be fixed prior to submitting the form
              </CardDescription>
            </CardHeader>
            <CardContent className="p-3">
              <ScrollArea className="h-72 max-w-80 overflow-y-visible">
                {extractedFormErrors.map((message, idx) => (
                  <div key={message}>
                    <p className="text-sm text-wrap">{message}</p>
                    <Separator className="my-2" />
                  </div>
                ))}
              </ScrollArea>
            </CardContent>
          </Card>
        </div>
      </div>
      <SchemaPageTable
        columns={columns}
        data={data}
        userDefinedTransformerMap={userDefinedMap}
        userDefinedTransformers={userDefinedTransformers}
        systemTransformerMap={systemMap}
        systemTransformers={systemTransformers}
        jobType={jobType}
        constraintHandler={constraintHandler}
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
  path = ''
): FormError[] {
  let messages: any[] = [];

  for (const key in errors) {
    const newPath = path ? `${path}.${key}` : key;
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
      messages = messages.concat(extractAllErrors(error, newPath));
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

export async function getConnectionSchema(
  accountId: string,
  connectionId?: string
): Promise<GetConnectionSchemaResponse | undefined> {
  if (!accountId || !connectionId) {
    return;
  }
  const res = await fetch(
    `/api/accounts/${accountId}/connections/${connectionId}/schema`,
    {
      method: 'GET',
      headers: {
        'content-type': 'application/json',
      },
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return GetConnectionSchemaResponse.fromJson(await res.json());
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
