'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/contexts/account-context';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetTransformers } from '@/libs/hooks/useGetTransformers';
import { GetConnectionSchemaResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  CreateJobRequest,
  CreateJobResponse,
  JobMapping,
  JobMappingTransformer,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import {
  DefineFormValues,
  FlowFormValues,
  FormValues,
  JobMappingFormValues,
  SCHEMA_FORM_SCHEMA,
  SchemaFormValues,
} from '../schema';
import { getColumns } from './components/SchemaTable/column';
import { DataTable } from './components/SchemaTable/data-table';

export default function Page({ searchParams }: PageProps): ReactElement {
  const router = useRouter();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';

  const account = useAccount();
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    `${sessionPrefix}-new-job-define`,
    { jobName: '' }
  );

  const [flowFormValues] = useSessionStorage<FlowFormValues>(
    `${sessionPrefix}-new-job-flow`,
    { sourceId: '', destinationId: '' }
  );

  useSessionStorage<SchemaFormValues>(`${sessionPrefix}-new-job-schema`, {
    mappings: [],
  });

  const form = useForm({
    resolver: yupResolver<SchemaFormValues>(SCHEMA_FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getConnectionSchema(flowFormValues.sourceId);
      if (!res) {
        return { mappings: [] };
      }
      const mappings = res.schemas.map((r) => {
        return {
          ...r,
          transformer: 'JOB_MAPPING_TRANSFORMER_UNSPECIFIED',
        };
      });
      return { mappings };
    },
  });
  useFormPersist(`${sessionPrefix}-new-job-schema`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: window.sessionStorage,
  });

  async function onSubmit(values: SchemaFormValues) {
    if (!account?.id) {
      return;
    }
    try {
      const job = await createNewJob(account.id, {
        define: defineFormValues,
        flow: flowFormValues,
        schema: values,
      });
      if (job.job?.id) {
        router.push(`/jobs/${job.job.id}`);
      } else {
        router.push(`/jobs`);
      }
    } catch (err) {
      console.error(err);
    }
  }

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Schemas"
          description="Define source to destination mappings for your data"
        />
      }
    >
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <JobTable data={form.getValues().mappings} />

          <div className="flex flex-row gap-1 justify-between">
            <Button key="back" type="button" onClick={() => router.back()}>
              Back
            </Button>
            <Button key="submit" type="submit">
              Save
            </Button>
          </div>
        </form>
      </Form>
    </OverviewContainer>
  );
}

interface JobTableProps {
  data: JobMappingFormValues[];
}

function JobTable(props: JobTableProps): ReactElement {
  const { data } = props;
  const { data: transformers, isLoading: transformersIsLoading } =
    useGetTransformers();

  if (transformersIsLoading) {
    return <Skeleton />;
  }

  const columns = getColumns({ transformers: transformers?.transformers });

  return (
    <div>
      <DataTable columns={columns} data={data} />
    </div>
  );
}

async function createNewJob(
  accountId: string,
  formData: FormValues
): Promise<CreateJobResponse> {
  const body = new CreateJobRequest({
    accountId,
    jobName: formData.define.jobName,
    cronSchedule: formData.define.cronSchedule,
    haltOnNewColumnAddition: false,
    mappings: formData.schema.mappings.map((m) => {
      return new JobMapping({
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer: m.transformer as unknown as JobMappingTransformer,
      });
    }),
    connectionSourceId: formData.flow.sourceId,
    connectionDestinationIds: [formData.flow.destinationId],
  });
  const res = await fetch(`/api/jobs`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateJobResponse.fromJson(await res.json());
}

async function getConnectionSchema(
  connectionId?: string
): Promise<GetConnectionSchemaResponse | undefined> {
  if (!connectionId) {
    return;
  }
  const res = await fetch(`/api/connections/${connectionId}/schema`, {
    method: 'GET',
    headers: {
      'content-type': 'application/json',
    },
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return GetConnectionSchemaResponse.fromJson(await res.json());
}
