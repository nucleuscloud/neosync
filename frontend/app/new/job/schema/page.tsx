'use client';
import {
  SchemaTable,
  getConnectionSchema,
} from '@/app/jobs/components/SchemaTable/schema-table';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/contexts/account-context';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
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
  SCHEMA_FORM_SCHEMA,
  SchemaFormValues,
} from '../schema';

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
          transformer: JobMappingTransformer.UNSPECIFIED as unknown as string,
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
        router.push(`/jobs/${job.job.id}/overview`);
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
          <SchemaTable data={form.getValues().mappings} />

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

async function createNewJob(
  accountId: string,
  formData: FormValues
): Promise<CreateJobResponse> {
  const body = new CreateJobRequest({
    accountId,
    jobName: formData.define.jobName,
    cronSchedule: formData.define.cronSchedule,
    haltOnNewColumnAddition: formData.define.haltOnNewColumnAddition,
    mappings: formData.schema.mappings.map((m) => {
      return new JobMapping({
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer: m.transformer as unknown as JobMappingTransformer,
        exclude: m.exclude,
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
