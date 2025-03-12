import {
  EditPiiDetectionJobFormValues,
  FilterPatternTableIdentifier,
} from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import {
  DataSampling,
  SourceConnectionId,
  TableScanFilterMode,
  TableScanFilterPatterns,
  UserPrompt,
} from '@/app/(mgmt)/[account]/new/job/piidetect/schema/FormInputs';
import { useEditPiiDetectionSchemaStore } from '@/app/(mgmt)/[account]/new/job/piidetect/schema/stores';
import ButtonText from '@/components/ButtonText';
import { useAccount } from '@/components/providers/account-provider';
import Spinner from '@/components/Spinner';
import { Button } from '@/components/ui/button';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import {
  Connection,
  ConnectionDataService,
  ConnectionService,
  JobService,
} from '@neosync/sdk';
import {
  FormEvent,
  ReactElement,
  useCallback,
  useEffect,
  useMemo,
} from 'react';
import { toast } from 'sonner';
import { ValidationError } from 'yup';
import { toJobSource, toPiiDetectJobTypeConfig } from '../../../util';

interface Props {
  jobId: string;
}

export default function PiiDetectConnectionCard({
  jobId,
}: Props): ReactElement {
  const {
    data,
    refetch: mutate,
    isLoading: isJobDataLoading,
  } = useQuery(JobService.method.getJob, { id: jobId }, { enabled: !!jobId });

  const { mutateAsync: updateJobSourceConnection } = useMutation(
    JobService.method.updateJobSourceConnection
  );

  const {
    formData,
    setFormData,
    errors,
    setErrors,
    isSubmitting,
    setSubmitting,
    sourcedFromRemote,
    setFromRemoteJob: setFromRemote,
  } = useEditPiiDetectionSchemaStore();

  const {
    data: connectionSchemaDataResp,
    isPending,
    isFetching,
  } = useQuery(
    ConnectionDataService.method.getConnectionSchema,
    { connectionId: formData.sourceId },
    { enabled: !!formData.sourceId }
  );

  const availableSchemas = useMemo(() => {
    if (isPending || !connectionSchemaDataResp) {
      return [];
    }
    const uniqueSchemas = new Set<string>();
    connectionSchemaDataResp?.schemas?.forEach((schema) => {
      uniqueSchemas.add(schema.schema);
    });
    return Array.from(uniqueSchemas);
  }, [connectionSchemaDataResp, isPending, isFetching]);

  const availableTableIdentifiers = useMemo(() => {
    if (isPending || !connectionSchemaDataResp) {
      return [];
    }
    const uniqueTableIdentifiers = new Map<
      string,
      FilterPatternTableIdentifier
    >();
    connectionSchemaDataResp?.schemas?.forEach((schema) => {
      uniqueTableIdentifiers.set(`${schema.schema}.${schema.table}`, {
        schema: schema.schema,
        table: schema.table,
      });
    });
    return Array.from(uniqueTableIdentifiers.values());
  }, [connectionSchemaDataResp, isPending, isFetching]);

  useEffect(() => {
    if (sourcedFromRemote || isJobDataLoading || !data?.job) {
      return;
    }
    setFromRemote(data.job);
  }, [sourcedFromRemote, isJobDataLoading, data?.job, setFromRemote]);

  const getConnectionById = useGetConnectionById();

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    const job = data?.job;
    if (isSubmitting || !job) {
      return;
    }

    try {
      setSubmitting(true);
      setErrors({});

      const validatedData = await EditPiiDetectionJobFormValues.validate(
        formData,
        {
          abortEarly: false,
        }
      );

      await updateJobSourceConnection({
        id: job.id,
        mappings: [],
        virtualForeignKeys: [],
        source: toJobSource(
          {
            connect: {
              destinations: [],
              sourceId: validatedData.sourceId,
              sourceOptions: {},
            },
          },
          getConnectionById
        ),
        jobType: {
          jobType: {
            case: 'piiDetect',
            value: toPiiDetectJobTypeConfig(validatedData),
          },
        },
      });
      toast.success('Successfully updated source connection!');
      const updatedJobResp = await mutate();
      if (updatedJobResp.data?.job) {
        setFromRemote(updatedJobResp.data?.job);
      }
    } catch (err) {
      if (err instanceof ValidationError) {
        const validationErrors: Record<string, string> = {};
        err.inner.forEach((error) => {
          if (error.path) {
            validationErrors[error.path] = error.message;
          }
        });
        setErrors(validationErrors);
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={onSubmit} className="space-y-6">
      <SourceConnectionId
        value={formData.sourceId}
        onChange={(value) => setFormData({ ...formData, sourceId: value })}
        error={errors['sourceId']}
        isDisabled={true}
      />
      <UserPrompt
        value={formData.userPrompt ?? ''}
        onChange={(value) => setFormData({ ...formData, userPrompt: value })}
        error={errors['userPrompt']}
      />
      <DataSampling
        value={formData.dataSampling}
        onChange={(value) => setFormData({ ...formData, dataSampling: value })}
        errors={errors}
      />
      <TableScanFilterMode
        value={formData.tableScanFilter.mode}
        onChange={(value) =>
          setFormData({
            ...formData,
            tableScanFilter: { ...formData.tableScanFilter, mode: value },
          })
        }
        error={errors['tableScanFilter.mode']}
      />

      <TableScanFilterPatterns
        value={formData.tableScanFilter.patterns}
        onChange={(value) =>
          setFormData({
            ...formData,
            tableScanFilter: {
              ...formData.tableScanFilter,
              patterns: value,
            },
          })
        }
        availableSchemas={availableSchemas}
        availableTableIdentifiers={availableTableIdentifiers}
        errors={errors}
        mode={formData.tableScanFilter.mode}
      />
      <div className="flex flex-row gap-1 justify-end">
        <Button
          type="submit"
          disabled={isSubmitting}
          className="w-full sm:w-auto"
        >
          <ButtonText
            leftIcon={isSubmitting ? <Spinner /> : undefined}
            text="Update"
          />
        </Button>
      </div>
    </form>
  );
}

function useGetConnectionById(): (id: string) => Connection | undefined {
  const connectionsRecord = useGetConnectionsRecord();
  return useCallback(
    (id: string) => connectionsRecord[id],
    [connectionsRecord]
  );
}

function useGetConnectionsRecord(): Record<string, Connection | undefined> {
  const { account } = useAccount();
  const {
    data: connectionsData,
    isLoading,
    isPending,
  } = useQuery(
    ConnectionService.method.getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  return useMemo(() => {
    if (isLoading || isPending || !connectionsData) {
      return {};
    }
    const connections = connectionsData?.connections ?? [];
    return connections.reduce(
      (record, conn) => {
        record[conn.id] = conn;
        return record;
      },
      {} as Record<string, Connection | undefined>
    );
  }, [connectionsData, isLoading, isPending]);
}
