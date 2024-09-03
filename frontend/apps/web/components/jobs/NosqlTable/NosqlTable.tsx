import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import Spinner from '@/components/Spinner';
import TruncatedText from '@/components/TruncatedText';
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
import { cn } from '@/libs/utils';
import { Transformer } from '@/shared/transformers';
import {
  convertJobMappingTransformerToForm,
  EditDestinationOptionsFormValues,
  JobMappingFormValues,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  GetConnectionSchemaResponse,
  JobMappingTransformer,
  Passthrough,
  SystemTransformer,
  TransformerConfig,
  TransformerSource,
} from '@neosync/sdk';
import { validateUserJavascriptCode } from '@neosync/sdk/connectquery';
import { CheckIcon, Pencil1Icon, TableIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import {
  HTMLProps,
  ReactElement,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import FormErrorsCard, { FormError } from '../SchemaTable/FormErrorsCard';
import { SchemaColumnHeader } from '../SchemaTable/SchemaColumnHeader';
import SchemaPageTable, { Row } from '../SchemaTable/SchemaPageTable';
import TransformerSelect from '../SchemaTable/TransformerSelect';
import { SchemaConstraintHandler } from '../SchemaTable/schema-constraint-handler';
import { TransformerHandler } from '../SchemaTable/transformer-handler';
import {
  DestinationDetails,
  OnTableMappingUpdateRequest,
} from './TableMappings/Columns';
import TableMappingsCard from './TableMappings/TableMappingsCard';
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
  onEditMappings(values: JobMappingFormValues, index: number): void;

  destinationOptions: EditDestinationOptionsFormValues[];
  destinationDetailsRecord: Record<string, DestinationDetails>;
  onDestinationTableMappingUpdate(req: OnTableMappingUpdateRequest): void;
  showDestinationTableMappings: boolean;
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
    destinationOptions,
    destinationDetailsRecord,
    onDestinationTableMappingUpdate,
    showDestinationTableMappings,
  } = props;
  const { account } = useAccount();
  const { handler, isLoading, isValidating } = useGetTransformersHandler(
    account?.id ?? ''
  );

  const collections = Array.from(Object.keys(schema));

  const columns = useMemo(
    () =>
      getColumns({
        onDelete(row) {
          onRemoveMappings([row]);
        },
        onEdit(row, index) {
          onEditMappings(row, index);
        },
        onDuplicate(row) {
          onAddMappings(row);
        },
        transformerHandler: handler,
        collections: collections,
        data: data,
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
              collections={collections}
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
      {showDestinationTableMappings && (
        <div>
          <TableMappingsCard
            mappings={destinationOptions}
            onUpdate={onDestinationTableMappingUpdate}
            destinationDetailsRecord={destinationDetailsRecord}
          />
        </div>
      )}
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

  const { account } = useAccount();
  const { mutateAsync: validateUserJsCodeAsync } = useMutation(
    validateUserJavascriptCode
  );

  const form = useForm<AddNewNosqlRecordFormValues>({
    resolver: yupResolver(AddNewNosqlRecordFormValues),
    mode: 'onChange',
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
    context: {
      accountId: account?.id,
      isUserJavascriptCodeValid: validateUserJsCodeAsync,
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
                The collection that you want to map.
              </FormDescription>
              <FormControl>
                <Select onValueChange={field.onChange} value={field.value}>
                  <SelectTrigger
                    className={cn(
                      field.value ? undefined : 'text-muted-foreground'
                    )}
                  >
                    <SelectValue placeholder="Select a collection" />
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
                Use dot notation to select a key for the mapping.
              </FormDescription>
              <FormControl>
                <Input {...field} placeholder="users.address.city" />
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
  onDuplicate(row: AddNewNosqlRecordFormValues[]): void;
  transformerHandler: TransformerHandler;
  onEdit(row: Row, index: number): void;
  collections: string[];
  data: JobMappingFormValues[];
}

function getColumns(props: GetColumnsProps): ColumnDef<Row>[] {
  const {
    onDelete,
    transformerHandler,
    onEdit,
    collections,
    onDuplicate,
    data,
  } = props;
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
      accessorFn: (row) => {
        if (row.schema && row.table) {
          return `${row.schema}.${row.table}`;
        }
        if (row.schema) {
          return row.schema;
        }
        return row.table;
      },
      id: 'schemaTable',
      footer: (props) => props.column.id,
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Collection" />
      ),
      cell: ({ getValue, row }) => {
        return (
          <EditCollection
            collections={collections}
            text={getValue<string>()}
            onEdit={(updatedObject) => {
              onEdit(
                {
                  schema: updatedObject.collection.split('.')[0],
                  table: updatedObject.collection.split('.')[1],
                  column: row.getValue('column'),
                  transformer: row.getValue('transformer'),
                },
                row.index
              );
            }}
          />
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
        const text = row.getValue<string>('column');
        return (
          <EditDocumentKey
            text={text}
            onEdit={(updatedObject) => {
              onEdit(
                {
                  schema: row.getValue('schema'),
                  table: row.getValue('table'),
                  column: updatedObject.column,
                  transformer: row.getValue('transformer'),
                },
                row.index
              );
            }}
          />
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
                    onEdit(
                      {
                        schema: row.getValue('schema'),
                        table: row.getValue('table'),
                        column: row.getValue('column'),
                        transformer: updatedTransformer,
                      },
                      row.index
                    )
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
                  onEdit(
                    {
                      schema: row.getValue('schema'),
                      table: row.getValue('table'),
                      column: row.getValue('column'),
                      transformer: updatedTransformer,
                    },
                    row.index
                  );
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
            onDuplicate={() => {
              console.log('data', data);
              onDuplicate([
                {
                  collection: `${row.getValue('schema')}.${row.getValue('table')}`,
                  // key: row.getValue('column') + 'copy', // need a way to check that we're not able to create multiple rows with the same keys
                  key: CreateDuplicatingMapping(
                    row.getValue('schema'),
                    row.getValue('table'),
                    row.getValue('column'),
                    data
                  ),
                  transformer: row.getValue('transformer'),
                },
              ]);
            }}
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

// searches through the table and creates a unique row copy based on the schema, table and column
function CreateDuplicatingMapping(
  schema: string,
  table: string,
  key: string,
  data: JobMappingFormValues[]
): string {
  let maxSuffix = 0;

  data.forEach((item) => {
    if (item.schema === schema && item.table === table) {
      const match = item.column.match(new RegExp(`^${key}_(\\d+)$`));
      if (match) {
        const suffix = parseInt(match[1], 10);
        if (suffix > maxSuffix) {
          maxSuffix = suffix;
        }
      }
    }
  });

  const newSuffix = maxSuffix + 1;
  const newKey = `${key}_${newSuffix}`;
  return newKey;
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
interface EditDocumentKeyProps {
  text: string;
  onEdit: (updatedObject: { column: string }) => void;
}

function EditDocumentKey({ text, onEdit }: EditDocumentKeyProps): ReactElement {
  const [isEditingMapping, setIsEditingMapping] = useState<boolean>(false);
  const [inputValue, setInputValue] = useState<string>(text);

  const handleSave = () => {
    onEdit({ column: inputValue });
    setIsEditingMapping(false);
  };

  return (
    <div className="w-full flex flex-row items-center gap-4">
      {isEditingMapping ? (
        <Input
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
        />
      ) : (
        <TruncatedText text={inputValue} />
      )}
      <Button
        variant="outline"
        size="sm"
        className="hidden h-[36px] lg:flex"
        type="button"
        onClick={() => {
          if (isEditingMapping) {
            handleSave();
          } else {
            setIsEditingMapping(true);
          }
        }}
      >
        {isEditingMapping ? <CheckIcon /> : <Pencil1Icon />}
      </Button>
    </div>
  );
}

interface EditCollectionProps {
  collections: string[];
  text: string;
  onEdit: (updatedObject: { collection: string }) => void;
}

function EditCollection(props: EditCollectionProps): ReactElement {
  const { collections, text, onEdit } = props;

  const [isEditingMapping, setIsEditingMapping] = useState<boolean>(false);
  const [isSelectedCollection, setSelectedCollection] = useState<string>(text);

  const handleSave = () => {
    onEdit({ collection: isSelectedCollection });
    setIsEditingMapping(false);
  };

  return (
    <div className="w-full flex flex-row items-center gap-4">
      {isEditingMapping ? (
        <Select
          onValueChange={(val) => setSelectedCollection(val)}
          value={isSelectedCollection}
        >
          <SelectTrigger>
            <SelectValue
              placeholder="Select a collection"
              className="placeholder:text-muted-foreground/70"
            />
          </SelectTrigger>
          <SelectContent>
            {collections.map((collection) => (
              <SelectItem value={collection} key={collection}>
                {collection}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      ) : (
        <TruncatedText text={isSelectedCollection} />
      )}
      <Button
        variant="outline"
        size="sm"
        className="hidden h-[36px] lg:flex"
        type="button"
        onClick={() => {
          if (isEditingMapping) {
            handleSave();
          } else {
            setIsEditingMapping(true);
          }
        }}
      >
        {isEditingMapping ? <CheckIcon /> : <Pencil1Icon />}
      </Button>
    </div>
  );
}
