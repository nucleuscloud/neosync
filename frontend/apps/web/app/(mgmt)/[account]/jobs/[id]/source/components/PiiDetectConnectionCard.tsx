import { FilterPatternTableIdentifier } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import {
  DataSampling,
  TableScanFilterMode,
  TableScanFilterPatterns,
  UserPrompt,
} from '@/app/(mgmt)/[account]/new/job/piidetect/schema/FormInputs';
import { usePiiDetectionSchemaStore } from '@/app/(mgmt)/[account]/new/job/piidetect/schema/stores';
import ButtonText from '@/components/ButtonText';
import { useAccount } from '@/components/providers/account-provider';
import Spinner from '@/components/Spinner';
import { Button } from '@/components/ui/button';
import { useQuery } from '@connectrpc/connect-query';
import {
  Connection,
  ConnectionDataService,
  ConnectionService,
  JobService,
} from '@neosync/sdk';
import { FormEvent, ReactElement, useEffect, useMemo } from 'react';
import { getConnectionIdFromSource } from './util';

interface Props {
  jobId: string;
}

export default function PiiDetectConnectionCard({
  jobId,
}: Props): ReactElement {
  const { account } = useAccount();
  const {
    data,
    refetch: mutate,
    isLoading: isJobDataLoading,
  } = useQuery(JobService.method.getJob, { id: jobId }, { enabled: !!jobId });
  const sourceConnectionId = getConnectionIdFromSource(data?.job?.source);

  const { data: connectionsData } = useQuery(
    ConnectionService.method.getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];
  const connectionsRecord = connections.reduce(
    (record, conn) => {
      record[conn.id] = conn;
      return record;
    },
    {} as Record<string, Connection | undefined>
  );

  const {
    data: connectionSchemaDataResp,
    isPending,
    isFetching,
  } = useQuery(
    ConnectionDataService.method.getConnectionSchema,
    { connectionId: sourceConnectionId },
    { enabled: !!sourceConnectionId }
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

  const {
    formData,
    setFormData,
    errors,
    setErrors,
    isSubmitting,
    setSubmitting,
    sourcedFromRemote,
    setFromRemoteJob: setFromRemote,
  } = usePiiDetectionSchemaStore();

  useEffect(() => {
    if (sourcedFromRemote || isJobDataLoading || !data?.job) {
      return;
    }
    setFromRemote(data.job);
  }, [sourcedFromRemote, isJobDataLoading, data?.job, setFromRemote]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (isSubmitting) {
      return;
    }

    // todo
  }

  return (
    <form onSubmit={onSubmit} className="space-y-6">
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
