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
  getTransformerFromField,
  getTransformerSelectButtonText,
  isInvalidTransformer,
} from '@/util/util';
import {
  convertJobMappingTransformerToForm,
  EditDestinationOptionsFormValues,
  JobMappingFormValues,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import { PartialMessage } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ConnectError,
  GetConnectionSchemaResponse,
  JobMapping,
  JobMappingTransformer,
  Passthrough,
  SystemTransformer,
  TransformerConfig,
  TransformerSource,
  ValidateUserJavascriptCodeRequest,
  ValidateUserJavascriptCodeResponse,
} from '@neosync/sdk';
import { validateUserJavascriptCode } from '@neosync/sdk/connectquery';
import { CheckIcon, Pencil1Icon, TableIcon } from '@radix-ui/react-icons';
import { UseMutateAsyncFunction } from '@tanstack/react-query';
import { Row } from '@tanstack/react-table';
import { nanoid } from 'nanoid';
import {
  HTMLProps,
  ReactElement,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import { JobMappingRow, NOSQL_COLUMNS } from '../JobMappingTable/Columns';
import JobMappingTable from '../JobMappingTable/JobMappingTable';
import FormErrorsCard, { FormError } from '../SchemaTable/FormErrorsCard';
import { ImportMappingsConfig } from '../SchemaTable/ImportJobMappingsButton';
import TransformerSelect from '../SchemaTable/TransformerSelect';
import {
  TransformerHandler,
  TransformerResult,
} from '../SchemaTable/transformer-handler';
import { useOnExportMappings } from '../SchemaTable/useOnExportMappings';
import {
  DestinationDetails,
  OnTableMappingUpdateRequest,
} from './TableMappings/Columns';
import TableMappingsCard from './TableMappings/TableMappingsCard';

interface Props {
  data: JobMappingFormValues[];
  schema: Record<string, GetConnectionSchemaResponse>;
  isSchemaDataReloading: boolean;
  isJobMappingsValidating?: boolean;

  onValidate?(): void;

  formErrors: FormError[];
  onAddMappings(values: AddNewNosqlRecordFormValues[]): void;
  onRemoveMappings(indices: number[]): void;
  onEditMappings(values: JobMappingFormValues, index: number): void;

  destinationOptions: EditDestinationOptionsFormValues[];
  destinationDetailsRecord: Record<string, DestinationDetails>;
  onDestinationTableMappingUpdate(req: OnTableMappingUpdateRequest): void;
  showDestinationTableMappings: boolean;
  onImportMappingsClick(
    jobmappings: JobMapping[],
    importConfig: ImportMappingsConfig
  ): void;
  getAvailableTransformers(index: number): TransformerResult;
  getTransformerFromField(index: number): Transformer;
  onApplyDefaultClick(override: boolean): void;
  getAvailableTransformersForBulk(
    rows: Row<JobMappingRow>[]
  ): TransformerResult;
  getTransformerFromFieldValue(value: JobMappingTransformerForm): Transformer;
  onTransformerBulkUpdate(
    indices: number[],
    config: JobMappingTransformerForm
  ): void;
}

export default function NosqlTable(props: Props): ReactElement {
  const {
    data,
    schema,
    formErrors,
    isJobMappingsValidating,
    onValidate,
    onAddMappings,
    onRemoveMappings,
    onEditMappings,
    destinationOptions,
    destinationDetailsRecord,
    onDestinationTableMappingUpdate,
    showDestinationTableMappings,
    onImportMappingsClick,
    getAvailableTransformers,
    getTransformerFromField,
    onApplyDefaultClick,
    getAvailableTransformersForBulk,
    getTransformerFromFieldValue,
    onTransformerBulkUpdate,
  } = props;
  const { account } = useAccount();
  const { handler, isLoading, isValidating } = useGetTransformersHandler(
    account?.id ?? ''
  );

  const collections = Array.from(Object.keys(schema));

  // useMemo ensures that we don't recreate the set unless the data changes
  const keySet = useMemo(() => {
    const set = new Set<string>();
    data.forEach((item: JobMappingFormValues) => {
      set.add(`${item.schema}.${item.table}.${item.column}`);
    });
    return set;
  }, [data]);

  // useCallback ensures that we only re-run the function if the keySet changes
  const isDuplicateKey = useCallback(
    (newValue: string, schema: string, table: string) => {
      const key = `${schema}.${table}.${newValue}`;
      return keySet.has(key);
    },
    [keySet]
  );

  // used to calculate the collections that can be updated based on a given key value
  // for ex. if a collection.key is "a.b.c" and we want to update it to "d.e.c" but "d.e.c" already exists, then we don't want to show the "d.e." collection as an uption for the update
  const filteredCollections = useCallback(
    (index: number) => {
      const currentColumn = data[index].column;
      const currentSchemaTable = `${data[index].schema}.${data[index].table}`;

      const conflictRows = data.filter(
        (obj) =>
          obj.column === currentColumn &&
          `${obj.schema}.${obj.table}` !== currentSchemaTable
      );

      return collections.filter(
        (item) =>
          !conflictRows.some((obj) => `${obj.schema}.${obj.table}` === item)
      );
    },
    [data, collections]
  );

  const tableData = useMemo(() => {
    return data.map((d): JobMappingRow => {
      return {
        schema: d.schema,
        table: d.table,
        column: d.column,
        dataType: '',
        attributes: '',
        constraints: '',
        isNullable: '',
        transformer: d.transformer,
      };
    });
  }, [data.length]);

  const { onClick: onExportMappingsClick } = useOnExportMappings({
    jobMappings: data,
  });

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
              isDuplicateKey={isDuplicateKey}
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
      <JobMappingTable<JobMappingRow, any>
        data={tableData}
        columns={NOSQL_COLUMNS}
        onTransformerUpdate={(idx, config) => {
          const row = data[idx];
          onEditMappings({ ...row, transformer: config }, idx);
        }}
        getAvailableTransformers={getAvailableTransformers}
        getTransformerFromField={getTransformerFromField}
        onExportMappingsClick={onExportMappingsClick}
        onImportMappingsClick={onImportMappingsClick}
        isApplyDefaultTransformerButtonDisabled={data.length === 0}
        getAvalableTransformersForBulk={getAvailableTransformersForBulk}
        getTransformerFromFieldValue={getTransformerFromFieldValue}
        onTransformerBulkUpdate={onTransformerBulkUpdate}
        onApplyDefaultClick={onApplyDefaultClick}
        onDeleteRow={(idx) => onRemoveMappings([idx])}
        onDuplicateRow={(idx) => {
          const row = data[idx];
          onAddMappings([
            {
              collection: `${row.schema}.${row.table}`,
              key: createDuplicateKey(row.column),
              transformer: row.transformer,
            },
          ]);
        }}
      />
    </div>
  );
}
interface AddNewRecordProps {
  collections: string[];
  onSubmit(values: AddNewNosqlRecordFormValues): void;
  transformerHandler: TransformerHandler;
  isDuplicateKey: (
    value: string,
    schema: string,
    table: string,
    currValue?: string
  ) => boolean;
}

const AddNewNosqlRecordFormValues = Yup.object({
  collection: Yup.string().required('The Collection is required.'),
  key: Yup.string()
    .required('The Key is required.')
    .test({
      name: 'uniqueMapping',
      message: 'This key already exists in the selected collection.',
      test: function (value, context) {
        const { collection } = this.parent;

        if (!collection || !value) {
          return true;
        }

        const lastDotIndex = collection.lastIndexOf('.');
        const schema = collection.substring(0, lastDotIndex);
        const table = collection.substring(lastDotIndex + 1);

        return (
          !context?.options?.context?.isDuplicateKey(value, schema, table) ||
          this.createError({
            message: 'This key already exists in this collection.',
          })
        );
      },
    }),
  transformer: JobMappingTransformerForm,
});

type AddNewNosqlRecordFormValues = Yup.InferType<
  typeof AddNewNosqlRecordFormValues
>;

interface AddNewNosqlRecordFormContext {
  accountId: string;
  isUserJavascriptCodeValid: UseMutateAsyncFunction<
    ValidateUserJavascriptCodeResponse,
    ConnectError,
    PartialMessage<ValidateUserJavascriptCodeRequest>,
    unknown
  >;
  isDuplicateKey: (value: string, schema: string, table: string) => boolean;
}

function AddNewRecord(props: AddNewRecordProps): ReactElement {
  const { collections, onSubmit, transformerHandler, isDuplicateKey } = props;

  const { account } = useAccount();
  const { mutateAsync: validateUserJsCodeAsync } = useMutation(
    validateUserJavascriptCode
  );
  const form = useForm<
    AddNewNosqlRecordFormValues,
    AddNewNosqlRecordFormContext
  >({
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
      accountId: account?.id ?? '',
      isUserJavascriptCodeValid: validateUserJsCodeAsync,
      isDuplicateKey: isDuplicateKey,
    },
  });

  return (
    <div className="flex flex-col w-full">
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
                <Select
                  onValueChange={(value) => {
                    field.onChange(value);
                    form.clearErrors('key');
                    const currentKey = form.getValues('key');
                    if (currentKey) {
                      const lastDotIndex = value.lastIndexOf('.');
                      const schema = value.substring(0, lastDotIndex);
                      const table = value.substring(lastDotIndex + 1);
                      if (isDuplicateKey(currentKey, schema, table)) {
                        form.setError('key', {
                          type: 'manual',
                          message:
                            'This key already exists in the selected collection.',
                        });
                      }
                    }
                  }}
                  value={field.value}
                >
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
                <Input
                  {...field}
                  placeholder="users.address.city"
                  disabled={!form.getValues('collection')}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="transformer"
          render={({ field }) => {
            const fv = field.value;
            const transformer = getTransformerFromField(transformerHandler, fv);
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
                        buttonText={getTransformerSelectButtonText(transformer)}
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
                      disabled={isInvalidTransformer(transformer)}
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
            disabled={!form.formState.isValid}
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

// interface GetColumnsProps {
//   onDelete(row: Row): void;
//   onDuplicate(row: Row): void;
//   transformerHandler: TransformerHandler;
//   onEdit(row: Row, index: number): void;
//   isDuplicateKey: (
//     newValue: string,
//     schema: string,
//     table: string,
//     currValue?: string
//   ) => boolean;
//   filteredCollections(ind: number): string[];
// }

// function getColumns(props: GetColumnsProps): ColumnDef<Row>[] {
//   const {
//     onDelete,
//     transformerHandler,
//     onEdit,
//     onDuplicate,
//     isDuplicateKey,
//     filteredCollections,
//   } = props;
//   return [
//     {
//       accessorKey: 'isSelected',
//       header: ({ table }) => (
//         <IndeterminateCheckbox
//           {...{
//             checked: table.getIsAllRowsSelected(),
//             indeterminate: table.getIsSomeRowsSelected(),
//             onChange: table.getToggleAllRowsSelectedHandler(),
//           }}
//         />
//       ),
//       cell: ({ row }) => (
//         <div>
//           <IndeterminateCheckbox
//             {...{
//               checked: row.getIsSelected(),
//               indeterminate: row.getIsSomeSelected(),
//               onChange: row.getToggleSelectedHandler(),
//             }}
//           />
//         </div>
//       ),
//       enableSorting: false,
//       enableHiding: false,
//       size: 30,
//     },
//     {
//       accessorKey: 'schema',
//       header: ({ column }) => (
//         <SchemaColumnHeader column={column} title="Schema" />
//       ),
//     },
//     {
//       accessorKey: 'table',
//       header: ({ column }) => (
//         <SchemaColumnHeader column={column} title="Table" />
//       ),
//     },
//     {
//       accessorFn: (row) => {
//         if (row.schema && row.table) {
//           return `${row.schema}.${row.table}`;
//         }
//         if (row.schema) {
//           return row.schema;
//         }
//         return row.table;
//       },
//       id: 'schemaTable',
//       footer: (props) => props.column.id,
//       header: ({ column }) => (
//         <SchemaColumnHeader column={column} title="Collection" />
//       ),
//       cell: ({ getValue, row }) => {
//         return (
//           <EditCollection
//             text={getValue<string>()}
//             collections={filteredCollections(row.index)}
//             onEdit={(updatedObject) => {
//               const lastDotIndex = updatedObject.collection.lastIndexOf('.');
//               onEdit(
//                 {
//                   schema: updatedObject.collection.substring(0, lastDotIndex),
//                   table: updatedObject.collection.substring(lastDotIndex + 1),
//                   column: row.getValue('column'),
//                   transformer: row.getValue('transformer'),
//                 },
//                 row.index
//               );
//             }}
//           />
//         );
//       },
//       maxSize: 500,
//       size: 300,
//     },
//     {
//       accessorKey: 'column',
//       header: ({ column }) => (
//         <SchemaColumnHeader column={column} title="Document Key" />
//       ),
//       cell: ({ row }) => {
//         const text = row.getValue<string>('column');
//         return (
//           <EditDocumentKey
//             text={text}
//             isDuplicate={(newValue: string, currValue?: string) =>
//               currValue !== newValue &&
//               isDuplicateKey(
//                 newValue,
//                 row.getValue('schema'),
//                 row.getValue('table'),
//                 currValue
//               )
//             }
//             onEdit={(updatedObject) => {
//               onEdit(
//                 {
//                   schema: row.getValue('schema'),
//                   table: row.getValue('table'),
//                   column: updatedObject.column,
//                   transformer: row.getValue('transformer'),
//                 },
//                 row.index
//               );
//             }}
//           />
//         );
//       },
//       maxSize: 500,
//       size: 200,
//     },
//     {
//       id: 'transformer',
//       accessorKey: 'transformer',
//       header: ({ column }) => (
//         <SchemaColumnHeader column={column} title="Transformer" />
//       ),
//       cell: ({ row }) => {
//         // row.getValue doesn't work here due to a tanstack bug where the transformer value is out of sync with getValue
//         // row.original works here. There must be a caching bug with the transformer prop being an object.
//         // This may be related: https://github.com/TanStack/table/issues/5363
//         const fv = row.original.transformer;
//         const transformer = getTransformerFromField(transformerHandler, fv);
//         return (
//           <span className="max-w-[500px] truncate font-medium">
//             <div className="flex flex-row gap-2">
//               <div>
//                 <TransformerSelect
//                   getTransformers={() => transformerHandler.getTransformers()}
//                   buttonText={getTransformerSelectButtonText(transformer)}
//                   value={fv}
//                   onSelect={(updatedTransformer) =>
//                     onEdit(
//                       {
//                         schema: row.getValue('schema'),
//                         table: row.getValue('table'),
//                         column: row.getValue('column'),
//                         transformer: updatedTransformer,
//                       },
//                       row.index
//                     )
//                   }
//                   side={'left'}
//                   disabled={false}
//                   buttonClassName="w-[175px]"
//                 />
//               </div>
//               <EditTransformerOptions
//                 transformer={transformer}
//                 value={fv}
//                 onSubmit={(updatedTransformer) => {
//                   onEdit(
//                     {
//                       schema: row.getValue('schema'),
//                       table: row.getValue('table'),
//                       column: row.getValue('column'),
//                       transformer: updatedTransformer,
//                     },
//                     row.index
//                   );
//                 }}
//                 disabled={isInvalidTransformer(transformer)}
//               />
//             </div>
//           </span>
//         );
//       },
//     },
//     {
//       id: 'actions',
//       header: () => <p>Actions</p>,
//       cell: ({ row }) => {
//         return (
//           <DataTableRowActions
//             row={row}
//             onDuplicate={() => onDuplicate(row.original)}
//             onDelete={() =>
//               onDelete({
//                 schema: row.getValue('schema'),
//                 table: row.getValue('table'),
//                 column: row.getValue('column'),
//                 transformer: row.getValue('transformer'),
//               })
//             }
//           />
//         );
//       },
//     },
//   ];
// }

function createDuplicateKey(key: string): string {
  const uniqueSuffix = nanoid(6);
  return `${key}_${uniqueSuffix}`;
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

interface EditCollectionProps {
  collections: string[];
  text: string;
  onEdit: (updatedObject: { collection: string }) => void;
}

function EditCollection(props: EditCollectionProps): ReactElement {
  const { text, collections, onEdit } = props;

  const [isEditingMapping, setIsEditingMapping] = useState<boolean>(false);
  const [isSelectedCollection, setSelectedCollection] = useState<string>(text);

  const handleSave = () => {
    onEdit({ collection: isSelectedCollection });
    setIsEditingMapping(false);
  };

  return (
    <div className="w-full flex flex-row items-center gap-1">
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
