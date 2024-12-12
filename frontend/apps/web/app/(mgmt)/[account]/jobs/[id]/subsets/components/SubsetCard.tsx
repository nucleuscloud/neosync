import {
  ConnectionConfigCase,
  getConnectionType,
} from '@/app/(mgmt)/[account]/connections/util';
import { SubsetFormValues } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import SubsetOptionsForm from '@/components/jobs/Form/SubsetOptionsForm';
import EditItem from '@/components/jobs/subsets/EditItem';
import SubsetTable from '@/components/jobs/subsets/subset-table/SubsetTable';
import { TableRow } from '@/components/jobs/subsets/subset-table/column';
import {
  GetColumnsForSqlAutocomplete,
  buildRowKey,
  buildTableRowData,
  isValidSubsetType,
} from '@/components/jobs/subsets/utils';
import LearnMoreLink from '@/components/labels/LearnMoreLink';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Form } from '@/components/ui/form';
import { Separator } from '@/components/ui/separator';
import { getErrorMessage } from '@/util/util';
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ConnectionConfig,
  GetJobResponse,
  JobSourceOptions,
} from '@neosync/sdk';
import {
  getConnection,
  getConnectionTableConstraints,
  getJob,
  setJobSourceSqlConnectionSubsets,
} from '@neosync/sdk/connectquery';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useQueryClient } from '@tanstack/react-query';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { toJobSourceSqlSubsetSchemas } from '../../../util';
import { getConnectionIdFromSource } from '../../source/components/util';
import SubsetSkeleton from './SubsetSkeleton';

interface Props {
  jobId: string;
}

export default function SubsetCard(props: Props): ReactElement {
  const { jobId } = props;
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const { data, isLoading: isJobLoading } = useQuery(
    getJob,
    { id: jobId },
    { enabled: !!jobId }
  );
  const queryclient = useQueryClient();
  const sourceConnectionId = getConnectionIdFromSource(data?.job?.source);
  const { data: tableConstraints, isFetching: isTableConstraintsValidating } =
    useQuery(
      getConnectionTableConstraints,
      { connectionId: sourceConnectionId },
      { enabled: !!sourceConnectionId }
    );
  const { mutateAsync: setJobSubsets } = useMutation(
    setJobSourceSqlConnectionSubsets
  );
  const { data: sourceConnectionData } = useQuery(
    getConnection,
    { id: sourceConnectionId },
    { enabled: !!sourceConnectionId }
  );

  const fkConstraints = tableConstraints?.foreignKeyConstraints;

  const [rootTables, setRootTables] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (!isTableConstraintsValidating && fkConstraints) {
      data?.job?.mappings.forEach((m) => {
        const tn = `${m.schema}.${m.table}`;
        if (!fkConstraints[tn]) {
          rootTables.add(tn);
          setRootTables(rootTables);
        }
      });
    }
  }, [fkConstraints, isTableConstraintsValidating]);

  const formValues = getFormValues(data?.job?.source?.options);
  const form = useForm({
    resolver: yupResolver<SubsetFormValues>(SubsetFormValues),
    defaultValues: { subsets: [] },
    values: formValues,
  });

  const tableRowData = buildTableRowData(
    data?.job?.mappings ?? [],
    rootTables,
    form.watch().subsets // ensures that all form changes cause a re-render since stuff happens outside of the form that depends on the form values
  );
  const [itemToEdit, setItemToEdit] = useState<TableRow | undefined>();

  const formValuesMap = new Map(
    formValues.subsets.map((ss) => [buildRowKey(ss.schema, ss.table), ss])
  );

  if (isJobLoading) {
    return (
      <div className="space-y-10">
        <SubsetSkeleton />
      </div>
    );
  }

  const connectionType = getConnectionType(
    sourceConnectionData?.connection?.connectionConfig ?? new ConnectionConfig()
  );

  if (!isValidSubsetType(connectionType)) {
    return (
      <Alert variant="warning">
        <ExclamationTriangleIcon className="h-4 w-4" />
        <AlertTitle>Heads up!</AlertTitle>
        <AlertDescription>
          The source connection configured does not currently support
          subsettings
        </AlertDescription>
      </Alert>
    );
  }

  async function onSubmit(values: SubsetFormValues): Promise<void> {
    if (!isValidSubsetType(connectionType)) {
      return;
    }

    try {
      const updatedJobRes = await setJobSubsets({
        id: jobId,
        subsetByForeignKeyConstraints:
          values.subsetOptions.subsetByForeignKeyConstraints,
        schemas: toJobSourceSqlSubsetSchemas(values, connectionType),
      });
      toast.success('Successfully updated database subsets');
      queryclient.setQueryData(
        createConnectQueryKey(getJob, { id: updatedJobRes.job?.id }),
        new GetJobResponse({ job: updatedJobRes.job })
      );
    } catch (err) {
      console.error(err);
      toast.error('Unable to update database subsets', {
        description: getErrorMessage(err),
      });
    }
  }

  function hasLocalChange(schema: string, table: string): boolean {
    const key = buildRowKey(schema, table);
    const trData = tableRowData[key];

    const svrData = formValuesMap.get(key);

    if (!svrData && !!trData.where) {
      return true;
    }

    return trData.where !== svrData?.whereClause;
  }

  function onLocalRowReset(schema: string, table: string): void {
    const key = buildRowKey(schema, table);
    const idx = form
      .getValues()
      .subsets.findIndex(
        (item) => buildRowKey(item.schema, item.table) === key
      );
    if (idx >= 0) {
      const svrData = formValuesMap.get(key);
      form.setValue(`subsets.${idx}`, {
        schema: schema,
        table: table,
        whereClause: svrData?.whereClause ?? undefined,
      });
    }
  }

  return (
    <div>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="flex flex-col">
          {showSubsetOptions(connectionType) && (
            <SubsetOptionsForm maxColNum={2} />
          )}
          <div className="flex flex-col gap-2">
            <div>
              <SubsetTable
                data={Object.values(tableRowData)}
                onEdit={(schema, table) => {
                  setIsDialogOpen(true);
                  const key = buildRowKey(schema, table);
                  if (tableRowData[key]) {
                    setItemToEdit({
                      ...tableRowData[key],
                    });
                  }
                }}
                hasLocalChange={hasLocalChange}
                onReset={onLocalRowReset}
              />
            </div>
            <div
              // this prevents the tooltips inside of the dialog from automatically opening when the dialog opens since it stops the focus from being set in the dialog component
              onFocusCapture={(e) => {
                e.stopPropagation();
              }}
            >
              <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent
                  className="max-w-5xl"
                  onPointerDownOutside={(e) => e.preventDefault()}
                  onEscapeKeyDown={(e) => e.preventDefault()}
                >
                  <DialogHeader>
                    <div className="flex flex-row w-full">
                      <div className="flex flex-col space-y-2 w-full">
                        <div className="flex flex-row justify-between items-center">
                          <div className="flex flex-row gap-4">
                            <DialogTitle className="text-xl">
                              Subset Query
                            </DialogTitle>
                          </div>
                        </div>
                        <div className="flex flex-row items-center gap-2">
                          <DialogDescription>
                            Subset your data using SQL expressions.{' '}
                            <LearnMoreLink href="https://docs.neosync.dev/table-constraints/subsetting" />
                          </DialogDescription>
                        </div>
                      </div>
                    </div>
                    <Separator />
                  </DialogHeader>
                  <div className="pt-4">
                    <EditItem
                      connectionId={sourceConnectionId ?? ''}
                      item={itemToEdit}
                      onItem={setItemToEdit}
                      onCancel={() => {
                        setItemToEdit(undefined);
                        setIsDialogOpen(false);
                      }}
                      columns={GetColumnsForSqlAutocomplete(
                        data?.job?.mappings ?? [],
                        itemToEdit
                      )}
                      onSave={() => {
                        if (!itemToEdit) {
                          return;
                        }
                        const key = buildRowKey(
                          itemToEdit.schema,
                          itemToEdit.table
                        );
                        const idx = form
                          .getValues()
                          .subsets.findIndex(
                            (item) =>
                              buildRowKey(item.schema, item.table) === key
                          );
                        if (idx >= 0) {
                          form.setValue(`subsets.${idx}`, {
                            schema: itemToEdit.schema,
                            table: itemToEdit.table,
                            whereClause: itemToEdit.where,
                          });
                        } else {
                          form.setValue(
                            `subsets`,
                            form.getValues().subsets.concat({
                              schema: itemToEdit.schema,
                              table: itemToEdit.table,
                              whereClause: itemToEdit.where,
                            })
                          );
                        }
                        setItemToEdit(undefined);
                        setIsDialogOpen(false);
                      }}
                      connectionType={connectionType}
                    />
                  </div>
                </DialogContent>
              </Dialog>
            </div>
            <div className="flex flex-row pt-10 justify-end">
              <Button key="submit" type="submit">
                Save
              </Button>
            </div>
          </div>
        </form>
      </Form>
    </div>
  );
}

// Determines if the subset options should be wholesale shown
// May need to change this in the future if a subset of the options apply to specific source connections.
// Currently there is only one and it only applies to pg/mysql
export function showSubsetOptions(
  connType: ConnectionConfigCase | null
): boolean {
  return (
    connType === 'pgConfig' ||
    connType === 'mysqlConfig' ||
    connType === 'mssqlConfig'
  );
}

function getFormValues(sourceOpts?: JobSourceOptions): SubsetFormValues {
  switch (sourceOpts?.config.case) {
    case 'postgres':
    case 'mysql':
    case 'mssql': {
      const schemas = sourceOpts.config.value.schemas;
      const subsets: SubsetFormValues['subsets'] = schemas.flatMap((schema) => {
        return schema.tables.map((table) => {
          return {
            schema: schema.schema,
            table: table.table,
            whereClause: table.whereClause,
          };
        });
      });
      return {
        subsets,
        subsetOptions: {
          subsetByForeignKeyConstraints:
            sourceOpts.config.value.subsetByForeignKeyConstraints,
        },
      };
    }
    case 'dynamodb': {
      const tables = sourceOpts.config.value.tables;
      const subsets: SubsetFormValues['subsets'] = tables.map((tableOpt) => {
        return {
          schema: 'dynamodb',
          table: tableOpt.table,
          whereClause: tableOpt.whereClause,
        };
      });
      return {
        subsets,
        subsetOptions: {
          subsetByForeignKeyConstraints: false,
        },
      };
    }
    default: {
      return {
        subsets: [],
        subsetOptions: { subsetByForeignKeyConstraints: false },
      };
    }
  }
}
