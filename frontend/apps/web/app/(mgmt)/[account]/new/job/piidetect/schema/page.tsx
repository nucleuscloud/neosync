'use client';
import { getNewJobSessionKeys } from '@/app/(mgmt)/[account]/jobs/util';
import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import Spinner from '@/components/Spinner';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { getSingleOrUndefined } from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { ConnectionDataService } from '@neosync/sdk';
import { useRouter } from 'next/navigation';
import { FormEvent, ReactElement, use, useEffect, useMemo } from 'react';
import { useSessionStorage } from 'usehooks-ts';
import { ValidationError } from 'yup';
import {
  PiiDetectionConnectFormValues,
  PiiDetectionSchemaFormValues,
} from '../../job-form-validations';
import JobsProgressSteps, {
  getJobProgressSteps,
} from '../../JobsProgressSteps';
import { TableScanFilterMode, TableScanFilterPatterns } from './FormInputs';
import { usePiiDetectionSchemaStore } from './stores';

export default function Page(props: PageProps): ReactElement {
  const searchParams = use(props.searchParams);
  const { account } = useAccount();
  const router = useRouter();
  // const posthog = usePostHog();

  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = getSingleOrUndefined(searchParams?.sessionId) ?? '';
  const sessionKeys = getNewJobSessionKeys(sessionPrefix);

  // Used to complete the whole form
  // const defineFormKey = sessionKeys.global.define;
  // const [defineFormValues] = useSessionStorage<DefineFormValues>(
  //   defineFormKey,
  //   { jobName: '' }
  // );

  const connectFormKey = sessionKeys.piidetect.connect;
  const [connectFormValues] = useSessionStorage<PiiDetectionConnectFormValues>(
    connectFormKey,
    {
      sourceId: '',
    }
  );

  // const schemaFormKey = sessionKeys.piidetect.schema;
  // const [schemaFormData] = useSessionStorage<PiiDetectionSchemaFormValues>(
  //   schemaFormKey,
  //   {
  //     dataSampling: {
  //       isEnabled: true,
  //     },
  //     tableScanFilter: {
  //       mode: 'include_all',
  //       patterns: {
  //         schemas: [],
  //         tables: [],
  //       },
  //     },
  //     userPrompt: '',
  //   }
  // );

  // const { data: connectionData, isLoading: isConnectionLoading } = useQuery(
  //   ConnectionService.method.getConnection,
  //   { id: connectFormValues.sourceId },
  //   { enabled: !!connectFormValues.sourceId }
  // );

  const {
    data: connectionSchemaDataResp,
    isPending,
    isFetching,
  } = useQuery(
    ConnectionDataService.method.getConnectionSchema,
    { connectionId: connectFormValues.sourceId },
    { enabled: !!connectFormValues.sourceId }
  );
  // const { data: connectionsData } = useQuery(
  //   ConnectionService.method.getConnections,
  //   { accountId: account?.id },
  //   { enabled: !!account?.id }
  // );
  // const connections = connectionsData?.connections ?? [];
  // const connectionsRecord = connections.reduce(
  //   (record, conn) => {
  //     record[conn.id] = conn;
  //     return record;
  //   },
  //   {} as Record<string, Connection>
  // );

  // const { mutateAsync: createNewSyncJob } = useMutation(
  //   JobService.method.createJob
  // );

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

  const {
    formData,
    setFormData,
    errors,
    setErrors,
    isSubmitting,
    setSubmitting,
  } = usePiiDetectionSchemaStore();

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (isSubmitting) {
      return;
    }

    try {
      setSubmitting(true);
      setErrors({});

      const validatedData = await PiiDetectionSchemaFormValues.validate(
        formData,
        {
          abortEarly: false,
        }
      );
      console.log('TODO', validatedData);
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
    <div className="flex flex-col gap-5">
      <OverviewContainer
        Header={
          <PageHeader
            header="Schema"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps('pii-detection')}
                stepName={'schema'}
              />
            }
          />
        }
        containerClassName="connect-page"
      >
        <form onSubmit={onSubmit} className="space-y-6">
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
            errors={errors}
          />
          <div className="flex flex-row gap-1 justify-between">
            <Button key="back" type="button" onClick={() => router.back()}>
              Back
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting}
              className="w-full sm:w-auto"
            >
              <ButtonText
                leftIcon={isSubmitting ? <Spinner /> : undefined}
                text="Create"
              />
            </Button>
          </div>
        </form>
      </OverviewContainer>
    </div>
  );
}
