import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useGetTransformersHandler } from '@/libs/hooks/useGetTransformersHandler';
import { Transformer } from '@/shared/transformers';
import {
  JobMappingFormValues,
  JobMappingTransformerForm,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  GetConnectionSchemaResponse,
  JobMappingTransformer,
  Passthrough,
  SystemTransformer,
  TransformerConfig,
  TransformerSource,
} from '@neosync/sdk';
import { TableIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import { HTMLProps, ReactElement, useEffect, useMemo, useRef } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import FormErrorsCard, { FormError } from '../SchemaTable/FormErrorsCard';
import { SchemaColumnHeader } from '../SchemaTable/SchemaColumnHeader';
import SchemaPageTable, { Row } from '../SchemaTable/SchemaPageTable';
import TransformerSelect from '../SchemaTable/TransformerSelect';
import { SchemaConstraintHandler } from '../SchemaTable/schema-constraint-handler';
import { TransformerHandler } from '../SchemaTable/transformer-handler';
import { DataTableRowActions } from './data-table-row-actions';

interface Props {
  data: JobMappingFormValues[];
  schema: Record<string, GetConnectionSchemaResponse>;
  isSchemaDataReloading: boolean;
  constraintHandler: SchemaConstraintHandler;
  isJobMappingsValidating?: boolean;

  onValidate?(): void;

  formErrors: FormError[];
  onAddMappings(values: AddNewNosqlRecordFormValues[]): void;
  onRemoveMappings(values: JobMappingFormValues[]): void;
  onEditMappings(values: JobMappingFormValues[]): void;
}

export default function NosqlTable(props: Props): ReactElement {
  const {
    data,
    schema,
    formErrors,
    isJobMappingsValidating,
    constraintHandler,
    onValidate,
    onAddMappings,
    onRemoveMappings,
    onEditMappings,
  } = props;
  const { account } = useAccount();
  const { handler, isLoading, isValidating } = useGetTransformersHandler(
    account?.id ?? ''
  );
  const columns = useMemo(
    () =>
      getColumns({
        onDelete(row) {
          onRemoveMappings([row]);
        },
        onEdit(row) {
          onEditMappings([row]);
        },
        transformerHandler: handler,
      }),
    [onRemoveMappings, onEditMappings, handler, isLoading]
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
              <CardTitle>Add new mapping</CardTitle>
              <div>{isValidating ? <Spinner /> : null}</div>
            </div>
            <CardDescription>
              Select a collection and input a document key to transform, along
              with specifying the relevant transformer.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AddNewRecord
              collections={Array.from(Object.keys(schema))}
              onSubmit={(values) => {
                onAddMappings([values]);
              }}
              transformerHandler={handler}
            />
          </CardContent>
        </Card>
        <FormErrorsCard
          formErrors={formErrors}
          isValidating={isJobMappingsValidating}
          onValidate={onValidate}
        />
      </div>
      <SchemaPageTable
        columns={columns}
        data={data}
        transformerHandler={handler}
        jobType="sync"
        constraintHandler={constraintHandler}
      />
    </div>
  );
}

interface AddNewRecordProps {
  collections: string[];
  onSubmit(values: AddNewNosqlRecordFormValues): void;
  transformerHandler: TransformerHandler;
}

const AddNewNosqlRecordFormValues = Yup.object({
  collection: Yup.string().required(),
  key: Yup.string().required(),
  transformer: JobMappingTransformerForm,
});
type AddNewNosqlRecordFormValues = Yup.InferType<
  typeof AddNewNosqlRecordFormValues
>;

function AddNewRecord(props: AddNewRecordProps): ReactElement {
  const { collections, onSubmit, transformerHandler } = props;

  const form = useForm<AddNewNosqlRecordFormValues>({
    resolver: yupResolver(AddNewNosqlRecordFormValues),
    defaultValues: {
      collection: '',
      key: '',
      transformer: convertJobMappingTransformerToForm(
        new JobMappingTransformer({
          source: TransformerSource.PASSTHROUGH,
          config: new TransformerConfig({
            config: {
              case: 'passthroughConfig',
              value: new Passthrough(),
            },
          }),
        })
      ),
    },
  });
  return (
    <div className="flex flex-col w-full space-y-4">
      <Form {...form}>
        <FormField
          control={form.control}
          name="collection"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Collection</FormLabel>
              <FormDescription>
                The collection to associate a mapping with.
              </FormDescription>
              <FormControl>
                <Select onValueChange={field.onChange} value={field.value}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {collections.map((collection) => (
                      <SelectItem value={collection} key={collection}>
                        {collection}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="key"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Document Key</FormLabel>
              <FormDescription>
                The key within the document to add a mapping with.
              </FormDescription>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="transformer"
          render={({ field }) => {
            let transformer: Transformer | undefined;
            const fv = field.value;
            if (
              fv.source === TransformerSource.USER_DEFINED &&
              fv.config.case === 'userDefinedTransformerConfig'
            ) {
              transformer = transformerHandler.getUserDefinedTransformerById(
                fv.config.value.id
              );
            } else {
              transformer = transformerHandler.getSystemTransformerBySource(
                fv.source
              );
            }
            const buttonText = transformer
              ? transformer.name
              : 'Select Transformer';
            return (
              <FormItem>
                <FormLabel>Transformer</FormLabel>
                <FormDescription>Select a transformer to map</FormDescription>
                <FormControl>
                  <div className="flex flex-row gap-2">
                    <div>
                      <TransformerSelect
                        getTransformers={() =>
                          transformerHandler.getTransformers()
                        }
                        buttonText={buttonText}
                        value={fv}
                        onSelect={field.onChange}
                        side={'left'}
                        disabled={false}
                        buttonClassName="w-[175px]"
                      />
                    </div>
                    <EditTransformerOptions
                      transformer={transformer ?? new SystemTransformer()}
                      value={fv}
                      onSubmit={(newvalue) => {
                        field.onChange(newvalue);
                      }}
                      disabled={!transformer}
                    />
                  </div>
                </FormControl>
                <FormMessage />
              </FormItem>
            );
          }}
        />
        <div className="flex justify-end">
          <Button
            type="button"
            onClick={(e) =>
              form.handleSubmit((values) => {
                onSubmit(values);
                form.resetField('key');
                form.resetField('transformer');
              })(e)
            }
          >
            Add
          </Button>
        </div>
      </Form>
    </div>
  );
}

interface GetColumnsProps {
  onDelete(row: Row): void;
  transformerHandler: TransformerHandler;
  onEdit(row: Row): void;
}

function getColumns(props: GetColumnsProps): ColumnDef<Row>[] {
  const { onDelete, onEdit, transformerHandler } = props;
  return [
    {
      accessorKey: 'isSelected',
      header: ({ table }) => (
        <IndeterminateCheckbox
          {...{
            checked: table.getIsAllRowsSelected(),
            indeterminate: table.getIsSomeRowsSelected(),
            onChange: table.getToggleAllRowsSelectedHandler(),
          }}
        />
      ),
      cell: ({ row }) => (
        <div>
          <IndeterminateCheckbox
            {...{
              checked: row.getIsSelected(),
              indeterminate: row.getIsSomeSelected(),
              onChange: row.getToggleSelectedHandler(),
            }}
          />
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
      size: 30,
    },
    {
      accessorKey: 'schema',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Schema" />
      ),
    },
    {
      accessorKey: 'table',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Table" />
      ),
    },
    {
      accessorFn: (row) => `${row.schema}.${row.table}`,
      id: 'schemaTable',
      footer: (props) => props.column.id,
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Collection" />
      ),
      cell: ({ getValue }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {getValue<string>()}
          </span>
        );
      },
      maxSize: 500,
      size: 300,
    },
    {
      accessorKey: 'column',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Document Key" />
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
      id: 'transformer',
      accessorKey: 'transformer',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Transformer" />
      ),
      cell: ({ row }) => {
        // row.getValue doesn't work here due to a tanstack bug where the transformer value is out of sync with getValue
        // row.original works here. There must be a caching bug with the transformer prop being an object.
        // This may be related: https://github.com/TanStack/table/issues/5363
        const fv = row.original.transformer;
        let transformer: Transformer | undefined;
        if (
          fv.source === TransformerSource.USER_DEFINED &&
          fv.config.case === 'userDefinedTransformerConfig'
        ) {
          transformer = transformerHandler.getUserDefinedTransformerById(
            fv.config.value.id
          );
        } else {
          transformer = transformerHandler.getSystemTransformerBySource(
            fv.source
          );
        }
        const buttonText = transformer
          ? transformer.name
          : 'Select Transformer';
        return (
          <span className="max-w-[500px] truncate font-medium">
            <div className="flex flex-row gap-2">
              <div>
                <TransformerSelect
                  getTransformers={() => transformerHandler.getTransformers()}
                  buttonText={buttonText}
                  value={fv}
                  onSelect={(updatedTransformer) =>
                    onEdit({
                      schema: row.getValue('schema'),
                      table: row.getValue('table'),
                      column: row.getValue('column'),
                      transformer: updatedTransformer,
                    })
                  }
                  side={'left'}
                  disabled={false}
                  buttonClassName="w-[175px]"
                />
              </div>
              <EditTransformerOptions
                transformer={transformer ?? new SystemTransformer()}
                value={fv}
                onSubmit={(updatedTransformer) => {
                  onEdit({
                    schema: row.getValue('schema'),
                    table: row.getValue('table'),
                    column: row.getValue('column'),
                    transformer: updatedTransformer,
                  });
                }}
                disabled={!transformer}
              />
            </div>
          </span>
        );
      },
    },
    {
      id: 'actions',
      header: () => <p>Actions</p>,
      cell: ({ row }) => {
        return (
          <DataTableRowActions
            row={row}
            onDelete={() =>
              onDelete({
                schema: row.getValue('schema'),
                table: row.getValue('table'),
                column: row.getValue('column'),
                transformer: row.getValue('transformer'),
              })
            }
          />
        );
      },
    },
  ];
}

function IndeterminateCheckbox({
  indeterminate,
  className = 'w-4 h-4 flex',
  ...rest
}: { indeterminate?: boolean } & HTMLProps<HTMLInputElement>) {
  const ref = useRef<HTMLInputElement>(null!);

  useEffect(() => {
    if (typeof indeterminate === 'boolean') {
      ref.current.indeterminate = !rest.checked && indeterminate;
    }
  }, [ref, indeterminate, rest.checked]);

  return (
    <input
      type="checkbox"
      ref={ref}
      className={className + ' cursor-pointer '}
      {...rest}
    />
  );
}
