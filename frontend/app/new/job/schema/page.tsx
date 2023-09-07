'use client';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import {
  SchemaTable,
  getConnectionSchema,
} from '@/components/jobs/SchemaTable/schema-table';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import {
  CreateJobRequest,
  CreateJobResponse,
  JobMapping,
  JobSourceOptions,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { SCHEMA_FORM_SCHEMA, SchemaFormValues } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import { DefineFormValues, FlowFormValues, FormValues } from '../schema';

export default function Page({ searchParams }: PageProps): ReactElement {
  const router = useRouter();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';

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
          transformer: '',
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
    try {
      const job = await createNewJob({
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

async function createNewJob(formData: FormValues): Promise<CreateJobResponse> {
  const body = new CreateJobRequest({
    jobName: formData.define.jobName,
    cronSchedule: formData.define.cronSchedule,
    sourceOptions: new JobSourceOptions({
      haltOnNewColumnAddition: false,
    }),
    mappings: formData.schema.mappings.map((m) => {
      return new JobMapping({
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer: m.transformer,
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
