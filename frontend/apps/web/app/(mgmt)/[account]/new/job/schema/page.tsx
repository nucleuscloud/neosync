'use client';

import FormPersist from '@/app/(mgmt)/FormPersist';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import NosqlTable from '@/components/jobs/NosqlTable/NosqlTable';
import { OnTableMappingUpdateRequest } from '@/components/jobs/NosqlTable/TableMappings/Columns';
import {
  getAllFormErrors,
  SchemaTable,
} from '@/components/jobs/SchemaTable/SchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
import { TransformerResult } from '@/components/jobs/SchemaTable/transformer-handler';
import { getTransformerFilter } from '@/components/jobs/SchemaTable/util';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { useGetTransformersHandler } from '@/libs/hooks/useGetTransformersHandler';
import { getSingleOrUndefined } from '@/libs/utils';
import { getErrorMessage, getTransformerFromField } from '@/util/util';
import {
  convertJobMappingTransformerToForm,
  JobMappingTransformerForm,
  SchemaFormValues,
  SchemaFormValuesDestinationOptions,
  VirtualForeignConstraintFormValues,
} from '@/yup-validations/jobs';
import { PartialMessage } from '@bufbuild/protobuf';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  DatabaseColumn,
  ForeignConstraintTables,
  GenerateDefault,
  GetConnectionSchemaMapRequest,
  GetConnectionSchemaMapsResponse,
  JobMappingTransformer,
  Passthrough,
  PrimaryConstraint,
  TransformerConfig,
  ValidateJobMappingsResponse,
  VirtualForeignConstraint,
  VirtualForeignKey,
} from '@neosync/sdk';
import {
  createJob,
  getConnection,
  getConnections,
  getConnectionSchemaMap,
  getConnectionSchemaMaps,
  getConnectionTableConstraints,
  validateJobMappings,
} from '@neosync/sdk/connectquery';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useCallback, useEffect, useMemo, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { useSessionStorage } from 'usehooks-ts';
import { useOnImportMappings } from '../../../jobs/[id]/source/components/useOnImportMappings';
import {
  getDestinationDetailsRecord,
  getFilteredTransformersForBulkSet,
  getOnSelectedTableToggle,
  isConnectionSubsettable,
  isDynamoDBConnection,
  isNosqlSource,
  shouldShowDestinationTableMappings,
} from '../../../jobs/[id]/source/components/util';
import {
  clearNewJobSession,
  getCreateNewSyncJobRequest,
  getNewJobSessionKeys,
  validateJobMapping,
} from '../../../jobs/util';
import JobsProgressSteps, { getJobProgressSteps } from '../JobsProgressSteps';
import { ConnectFormValues, DefineFormValues } from '../job-form-validations';

export interface ColumnMetadata {
  pk: { [key: string]: PrimaryConstraint };
  fk: { [key: string]: ForeignConstraintTables };
  isNullable: DatabaseColumn[];
}

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const posthog = usePostHog();
  const [validateMappingsResponse, setValidateMappingsResponse] = useState<
    ValidateJobMappingsResponse | undefined
  >();

  const [isValidatingMappings, setIsValidatingMappings] = useState(false);

  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = getSingleOrUndefined(searchParams?.sessionId) ?? '';
  const sessionKeys = getNewJobSessionKeys(sessionPrefix);

  // Used to complete the whole form
  const defineFormKey = sessionKeys.global.define;
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    defineFormKey,
    { jobName: '' }
  );

  const connectFormKey = sessionKeys.dataSync.connect;
  const [connectFormValues] = useSessionStorage<ConnectFormValues>(
    connectFormKey,
    {
      sourceId: '',
      sourceOptions: {},
      destinations: [{ connectionId: '', destinationOptions: {} }],
    }
  );

  const schemaFormKey = sessionKeys.dataSync.schema;
  const [schemaFormData] = useSessionStorage<SchemaFormValues>(schemaFormKey, {
    mappings: [],
    connectionId: '', // hack to track if source id changes
    destinationOptions: [],
  });

  const { data: connectionData, isLoading: isConnectionLoading } = useQuery(
    getConnection,
    { id: connectFormValues.sourceId },
    { enabled: !!connectFormValues.sourceId }
  );

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaMapLoading,
    isFetching: isSchemaMapValidating,
  } = useQuery(
    getConnectionSchemaMap,
    { connectionId: connectFormValues.sourceId },
    { enabled: !!connectFormValues.sourceId }
  );
  const { data: connectionsData } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];
  const connectionsRecord = connections.reduce(
    (record, conn) => {
      record[conn.id] = conn;
      return record;
    },
    {} as Record<string, Connection>
  );

  const { data: tableConstraints, isFetching: isTableConstraintsValidating } =
    useQuery(
      getConnectionTableConstraints,
      { connectionId: connectFormValues.sourceId },
      { enabled: !!connectFormValues.sourceId }
    );

  const { data: destinationConnectionSchemaMapsResp } = useQuery(
    getConnectionSchemaMaps,
    {
      requests: connectFormValues.destinations.map(
        (dest): PartialMessage<GetConnectionSchemaMapRequest> => ({
          connectionId: dest.connectionId,
        })
      ),
    },
    {
      enabled:
        (connectFormValues.destinations.length ?? 0) > 0 &&
        connectionData?.connection?.connectionConfig?.config?.case ===
          'dynamodbConfig',
    }
  );

  const { mutateAsync: createNewSyncJob } = useMutation(createJob);

  const form = useForm<SchemaFormValues>({
    resolver: yupResolver<SchemaFormValues>(SchemaFormValues),
    values: getFormValues(connectFormValues.sourceId, schemaFormData),
    context: { accountId: account?.id },
  });

  const { mutateAsync: validateJobMappingsAsync } =
    useMutation(validateJobMappings);

  async function onSubmit(values: SchemaFormValues) {
    if (!account || !source) {
      return;
    }
    if (!isConnectionSubsettable(source)) {
      try {
        const connMap = new Map(connections.map((c) => [c.id, c]));
        const job = await createNewSyncJob(
          getCreateNewSyncJobRequest(
            {
              define: defineFormValues,
              connect: connectFormValues,
              schema: values,
              // subset: {},
            },
            account.id,
            (id) => connMap.get(id)
          )
        );
        posthog.capture('New Job Flow Complete', {
          jobType: 'data-sync',
        });
        toast.success('Successfully created job!');

        clearNewJobSession(window.sessionStorage, sessionPrefix);

        if (job.job?.id) {
          router.push(`/${account?.name}/jobs/${job.job.id}`);
        } else {
          router.push(`/${account?.name}/jobs`);
        }
      } catch (err) {
        console.error(err);
        toast.error('Unable to create job', {
          description: getErrorMessage(err),
        });
      }
      return;
    }
    posthog.capture('New Job Flow Schema Complete', { jobType: 'data-sync' });
    router.push(`/${account?.name}/new/job/subset?sessionId=${sessionPrefix}`);
  }

  const formMappings = form.watch('mappings');
  const formVirtualForeignKeys = form.watch('virtualForeignKeys');
  async function validateMappings() {
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        connectFormValues.sourceId,
        formMappings,
        account?.id || '',
        [],
        validateJobMappingsAsync
      );
      setValidateMappingsResponse(res);
    } catch (error) {
      console.error('Failed to validate job mappings:', error);
      toast.error('Unable to validate job mappings', {
        description: getErrorMessage(error),
      });
    } finally {
      setIsValidatingMappings(false);
    }
  }

  async function validateVirtualForeignKeys(
    vfks: VirtualForeignConstraintFormValues[]
  ) {
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        connectFormValues.sourceId,
        formMappings,
        account?.id || '',
        vfks,
        validateJobMappingsAsync
      );
      setValidateMappingsResponse(res);
    } catch (error) {
      console.error('Failed to validate virtual foreign keys:', error);
      toast.error('Unable to validate virtual foreign keys', {
        description: getErrorMessage(error),
      });
    } finally {
      setIsValidatingMappings(false);
    }
  }

  const schemaConstraintHandler = useMemo(() => {
    const virtualForeignKeys = formVirtualForeignKeys?.map((v) => {
      return new VirtualForeignConstraint({
        schema: v.schema,
        table: v.table,
        columns: v.columns,
        foreignKey: new VirtualForeignKey({
          schema: v.foreignKey.schema,
          table: v.foreignKey.table,
          columns: v.foreignKey.columns,
        }),
      });
    });
    return getSchemaConstraintHandler(
      connectionSchemaDataMap?.schemaMap ?? {},
      tableConstraints?.primaryKeyConstraints ?? {},
      tableConstraints?.foreignKeyConstraints ?? {},
      tableConstraints?.uniqueConstraints ?? {},
      virtualForeignKeys ?? []
    );
  }, [
    isSchemaMapValidating,
    isTableConstraintsValidating,
    formVirtualForeignKeys,
  ]);
  const [selectedTables, setSelectedTables] = useState<Set<string>>(new Set());

  const { append, remove, update } = useFieldArray<SchemaFormValues>({
    control: form.control,
    name: 'mappings',
  });

  const { append: appendVfk, remove: removeVfk } =
    useFieldArray<SchemaFormValues>({
      control: form.control,
      name: 'virtualForeignKeys',
    });

  const onSelectedTableToggle = getOnSelectedTableToggle(
    connectionSchemaDataMap?.schemaMap ?? {},
    selectedTables,
    setSelectedTables,
    formMappings,
    remove,
    append
  );

  useEffect(() => {
    if (!connectFormValues.sourceId || !account?.id) {
      return;
    }
    const validateJobMappings = async () => {
      await validateMappings();
    };
    validateJobMappings();
  }, [selectedTables, connectFormValues.sourceId, account?.id]);

  useEffect(() => {
    if (
      isSchemaMapLoading ||
      selectedTables.size > 0 ||
      !connectFormValues.sourceId
    ) {
      return;
    }
    const js = getFormValues(connectFormValues.sourceId, schemaFormData);
    setSelectedTables(
      new Set(
        js.mappings.map((mapping) => `${mapping.schema}.${mapping.table}`)
      )
    );
  }, [isSchemaMapLoading, connectFormValues.sourceId]);

  async function addVirtualForeignKey(vfk: VirtualForeignConstraintFormValues) {
    appendVfk(vfk);
    const vfks = [vfk, ...(formVirtualForeignKeys || [])];
    await validateVirtualForeignKeys(vfks);
  }

  async function removeVirtualForeignKey(index: number) {
    const newVfks: VirtualForeignConstraintFormValues[] = [];
    formVirtualForeignKeys?.forEach((vfk, idx) => {
      if (idx != index) {
        newVfks.push(vfk);
      }
    });
    removeVfk(index);
    await validateVirtualForeignKeys(newVfks);
  }

  const onDestinationTableMappingUpdate = useCallback(
    (req: OnTableMappingUpdateRequest) => {
      const destOpts = form.getValues('destinationOptions');
      const destOpt = destOpts.find(
        (d) => d.destinationId === req.destinationId
      );
      const tm = destOpt?.dynamodb?.tableMappings.find(
        (tm) => tm.sourceTable === req.souceName
      );
      if (tm) {
        tm.destinationTable = req.tableName;
        form.setValue('destinationOptions', destOpts);
      }
      return;
    },
    []
  );

  const { onClick: onImportMappingsClick } = useOnImportMappings({
    setMappings(mappings) {
      form.setValue('mappings', mappings, {
        shouldDirty: true,
        shouldTouch: true,
        shouldValidate: false,
      });
    },
    getMappings() {
      return form.getValues('mappings');
    },
    appendNewMappings(mappings) {
      append(mappings);
    },
    setTransformer(idx, transformer) {
      form.setValue(`mappings.${idx}.transformer`, transformer, {
        shouldDirty: true,
        shouldTouch: true,
        shouldValidate: false,
      });
    },
    triggerUpdate() {
      form.trigger('mappings');
    },
    setSelectedTables: setSelectedTables,
  });

  const { handler, isLoading: isGetTransformersLoading } =
    useGetTransformersHandler(account?.id ?? '');

  if (isConnectionLoading || isSchemaMapLoading || isGetTransformersLoading) {
    return <SkeletonForm />;
  }

  const source = connectionData?.connection;

  const dynamoDbDestinationConnections =
    source && isDynamoDBConnection(source)
      ? connectFormValues.destinations
          .map((d) => connectionsRecord[d.connectionId])
          .filter((c) => !!c && isDynamoDBConnection(c))
      : [];

  const dynamoDbDestinations =
    source && isDynamoDBConnection(source)
      ? connectFormValues.destinations.map((d) => ({
          id: d.connectionId,
          connectionId: d.connectionId,
        }))
      : [];

  function getAvailableTransformers(idx: number): TransformerResult {
    const row = formMappings[idx];
    return handler.getFilteredTransformers(
      getTransformerFilter(
        schemaConstraintHandler,
        {
          schema: row.schema,
          table: row.table,
          column: row.column,
        },
        'sync'
      )
    );
  }

  function onTransformerUpdate(
    index: number,
    transformer: JobMappingTransformerForm
  ): void {
    const val = form.getValues(`mappings.${index}`);
    update(index, {
      schema: val.schema,
      table: val.table,
      column: val.column,
      transformer,
    });
  }

  function onApplyDefaultClick(override: boolean): void {
    const formMappings = form.getValues('mappings');
    formMappings.forEach((fm, idx) => {
      // skips setting the default transformer if the user has already set the transformer
      if (fm.transformer.config.case && !override) {
        return;
      } else {
        const colkey = {
          schema: fm.schema,
          table: fm.table,
          column: fm.column,
        };
        const isGenerated = schemaConstraintHandler.getIsGenerated(colkey);
        const identityType = schemaConstraintHandler.getIdentityType(colkey);
        const newJm =
          isGenerated && !identityType
            ? new JobMappingTransformer({
                config: new TransformerConfig({
                  config: {
                    case: 'generateDefaultConfig',
                    value: new GenerateDefault(),
                  },
                }),
              })
            : new JobMappingTransformer({
                config: new TransformerConfig({
                  config: {
                    case: 'passthroughConfig',
                    value: new Passthrough(),
                  },
                }),
              });
        onTransformerUpdate(idx, convertJobMappingTransformerToForm(newJm));
      }
    });
    setTimeout(() => {
      form.trigger('mappings');
    }, 0);
  }

  function onTransformerBulkUpdate(
    indices: number[],
    config: JobMappingTransformerForm
  ): void {
    indices.forEach((idx) => {
      onTransformerUpdate(idx, config);
    });
    setTimeout(() => {
      form.trigger('mappings');
    }, 0);
  }

  return (
    <div className="flex flex-col gap-5">
      <FormPersist formKey={schemaFormKey} form={form} />
      <OverviewContainer
        Header={
          <PageHeader
            header="Schema"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps(
                  'data-sync',
                  isConnectionSubsettable(source ?? new Connection())
                )}
                stepName={'schema'}
              />
            }
          />
        }
        containerClassName="connect-page"
      >
        <div />
      </OverviewContainer>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          {isNosqlSource(source ?? new Connection({})) && (
            <NosqlTable
              data={formMappings}
              schema={connectionSchemaDataMap?.schemaMap ?? {}}
              isSchemaDataReloading={isSchemaMapValidating}
              isJobMappingsValidating={isValidatingMappings}
              formErrors={getAllFormErrors(
                form.formState.errors,
                formMappings,
                validateMappingsResponse
              )}
              onValidate={validateMappings}
              onRemoveMappings={(indices) => {
                const indexSet = new Set(indices);
                const remainingTables = formMappings
                  .filter((_, idx) => !indexSet.has(idx))
                  .map((fm) => fm.table);

                if (indices.length > 0) {
                  remove(indices);
                }

                if (!source || isDynamoDBConnection(source)) {
                  return;
                }

                // Check and update destinationOptions if needed
                const destOpts = form.getValues('destinationOptions');
                const updatedDestOpts = destOpts
                  .map((opt) => {
                    if (opt.dynamodb) {
                      const updatedTableMappings =
                        opt.dynamodb.tableMappings.filter((tm) => {
                          // Check if any columns remain for the table
                          const tableColumnsExist = remainingTables.some(
                            (table) => table === tm.sourceTable
                          );
                          return tableColumnsExist;
                        });

                      return {
                        ...opt,
                        dynamoDb: {
                          ...opt.dynamodb,
                          tableMappings: updatedTableMappings,
                        },
                      };
                    }
                    return opt;
                  })
                  .filter(
                    (opt) => (opt.dynamodb?.tableMappings.length ?? 0) > 0
                  );
                form.setValue('destinationOptions', updatedDestOpts);
              }}
              onEditMappings={(values, index) => {
                if (index >= 0 && index < formMappings.length) {
                  update(index, values);
                }
              }}
              onAddMappings={(values) => {
                append(
                  values.map((v) => {
                    const [schema, table] = v.collection.split('.');
                    return {
                      schema,
                      table,
                      column: v.key,
                      transformer: v.transformer,
                    };
                  })
                );
                const uniqueCollections = Array.from(
                  new Set(values.map((v) => v.collection))
                );

                const destOpts = form.getValues('destinationOptions');
                const existing = new Map(
                  destOpts.map((d) => [d.destinationId, d])
                );
                const updated = dynamoDbDestinations.map(
                  (dest): SchemaFormValuesDestinationOptions => {
                    const opt = existing.get(dest.id);
                    if (opt) {
                      const sourceSet = new Set(
                        opt.dynamodb?.tableMappings.map(
                          (mapping) => mapping.sourceTable
                        ) ?? []
                      );

                      // Add missing uniqueCollections to the existing tableMappings
                      const updatedTableMappings = [
                        ...(opt.dynamodb?.tableMappings ?? []),
                        ...uniqueCollections
                          .map((c) => {
                            const [, table] = c.split('.');
                            return {
                              sourceTable: table,
                              destinationTable: '',
                            };
                          })
                          .filter(
                            (mapping) => !sourceSet.has(mapping.sourceTable)
                          ),
                      ];

                      return {
                        ...opt,
                        dynamodb: {
                          ...opt.dynamodb,
                          tableMappings: updatedTableMappings,
                        },
                      };
                    }

                    return {
                      destinationId: dest.id,
                      dynamodb: {
                        tableMappings: uniqueCollections.map((c) => {
                          const [, table] = c.split('.');
                          return {
                            sourceTable: table,
                            destinationTable: '',
                          };
                        }),
                      },
                    };
                  }
                );
                form.setValue('destinationOptions', updated);
              }}
              destinationDetailsRecord={getDestinationDetailsRecord(
                dynamoDbDestinations,
                connectionsRecord,
                destinationConnectionSchemaMapsResp ??
                  new GetConnectionSchemaMapsResponse()
              )}
              onDestinationTableMappingUpdate={onDestinationTableMappingUpdate}
              showDestinationTableMappings={shouldShowDestinationTableMappings(
                source ?? new Connection(),
                dynamoDbDestinationConnections.length > 0
              )}
              destinationOptions={form.watch('destinationOptions')}
              onImportMappingsClick={onImportMappingsClick}
              getAvailableTransformers={getAvailableTransformers}
              getTransformerFromField={(idx) => {
                const row = formMappings[idx];
                return getTransformerFromField(handler, row.transformer);
              }}
              getAvailableTransformersForBulk={(rows) => {
                return getFilteredTransformersForBulkSet(
                  rows,
                  handler,
                  schemaConstraintHandler,
                  'sync',
                  'nosql'
                );
              }}
              getTransformerFromFieldValue={(fvalue) => {
                return getTransformerFromField(handler, fvalue);
              }}
              onApplyDefaultClick={onApplyDefaultClick}
              onTransformerBulkUpdate={onTransformerBulkUpdate}
            />
          )}

          {!isNosqlSource(source ?? new Connection({})) && (
            <SchemaTable
              data={formMappings}
              jobType="sync"
              constraintHandler={schemaConstraintHandler}
              schema={connectionSchemaDataMap?.schemaMap ?? {}}
              isSchemaDataReloading={isSchemaMapValidating}
              isJobMappingsValidating={isValidatingMappings}
              selectedTables={selectedTables}
              onSelectedTableToggle={onSelectedTableToggle}
              formErrors={getAllFormErrors(
                form.formState.errors,
                formMappings,
                validateMappingsResponse
              )}
              onValidate={validateMappings}
              virtualForeignKeys={formVirtualForeignKeys}
              addVirtualForeignKey={addVirtualForeignKey}
              removeVirtualForeignKey={removeVirtualForeignKey}
              onImportMappingsClick={onImportMappingsClick}
              onTransformerUpdate={(idx, cfg) => {
                onTransformerUpdate(idx, cfg);
              }}
              getAvailableTransformers={getAvailableTransformers}
              getTransformerFromField={(idx) => {
                const row = formMappings[idx];
                return getTransformerFromField(handler, row.transformer);
              }}
              getAvailableTransformersForBulk={(rows) => {
                return getFilteredTransformersForBulkSet(
                  rows,
                  handler,
                  schemaConstraintHandler,
                  'sync',
                  'relational'
                );
              }}
              getTransformerFromFieldValue={(fvalue) => {
                return getTransformerFromField(handler, fvalue);
              }}
              onApplyDefaultClick={onApplyDefaultClick}
              onTransformerBulkUpdate={onTransformerBulkUpdate}
            />
          )}
          <div className="flex flex-row gap-1 justify-between">
            <Button key="back" type="button" onClick={() => router.back()}>
              Back
            </Button>
            <Button key="submit" type="submit">
              {isConnectionSubsettable(source ?? new Connection())
                ? 'Next'
                : 'Submit'}
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}

function getFormValues(
  connectionId: string,
  existingData: SchemaFormValues | undefined
): SchemaFormValues {
  const existingMappings = existingData?.mappings ?? [];
  if (
    existingData &&
    existingMappings.length > 0 &&
    existingData.connectionId === connectionId
  ) {
    return existingData;
  }

  return {
    mappings: [],
    virtualForeignKeys: [],
    connectionId,
    destinationOptions: [],
  };
}
