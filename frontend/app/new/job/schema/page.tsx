'use client';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import {
  SchemaTable,
  getConnectionSchema,
} from '@/components/jobs/SchemaTable/schema-table';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { useGetTransformers } from '@/libs/hooks/useGetTransformers';
import { Connection } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  CreateJobRequest,
  CreateJobResponse,
  JobDestination,
  JobMapping,
  JobSource,
  JobSourceOptions,
  SqlSourceConnectionOptions,
  Transformer,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import {
  SCHEMA_FORM_SCHEMA,
  SchemaFormValues,
  toJobDestinationOptions,
  toTransformerConfigOptions,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import { DefineFormValues, FlowFormValues, FormValues } from '../schema';

export default function Page({ searchParams }: PageProps): ReactElement {
  const account = useAccount();
  const router = useRouter();
  const { toast } = useToast();
  const { data: connectionsData } = useGetConnections(account?.id ?? '');

  const connections = connectionsData?.connections ?? [];
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';

  const { data: transformersData } = useGetTransformers();

  const transformers = transformersData?.transformers ?? [];

  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    `${sessionPrefix}-new-job-define`,
    { jobName: '' }
  );

  const [flowFormValues] = useSessionStorage<FlowFormValues>(
    `${sessionPrefix}-new-job-flow`,
    {
      sourceId: '',
      sourceOptions: {},
      destinations: [{ connectionId: '', destinationOptions: {} }],
    }
  );

  useSessionStorage<SchemaFormValues>(`${sessionPrefix}-new-job-schema`, {
    mappings: [],
  });

  async function getSchema() {
    try {
      const res = await getConnectionSchema(flowFormValues.sourceId);
      if (!res) {
        return { mappings: [] };
      }
      const mappings = res.schemas.map((r) => {
        return {
          ...r,
          transformer: {
            value: '',
            config: {},
          },
        };
      });
      return { mappings };
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to get connection schema',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
      return { mappings: [] };
    }
  }

  const form = useForm({
    resolver: yupResolver<SchemaFormValues>(SCHEMA_FORM_SCHEMA),
    defaultValues: async () => {
      return getSchema();
    },
  });
  const isBrowser = () => typeof window !== 'undefined';

  useFormPersist(`${sessionPrefix}-new-job-schema`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });

  function onSubmit(transformers: Transformer[]) {
    return async function (values: SchemaFormValues) {
      if (!account) {
        return;
      }
      try {
        const job = await createNewJob(
          {
            define: defineFormValues,
            flow: flowFormValues,
            schema: values,
          },
          account.id,
          connections,
          transformers
        );
        if (job.job?.id) {
          router.push(`/jobs/${job.job.id}`);
        } else {
          router.push(`/jobs`);
        }
      } catch (err) {
        console.error(err);
        toast({
          title: 'Unable to create job',
          description: getErrorMessage(err),
          variant: 'destructive',
        });
      }
    };
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
          <SchemaTable data={form.getValues().mappings || []} />

        <form
          onSubmit={form.handleSubmit(onSubmit(transformers))}
          className="space-y-8"
        >
          <SchemaTable data={form.getValues().mappings || []} />
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
  formData: FormValues,
  accountId: string,
  connections: Connection[],
  transformers: Transformer[]
): Promise<CreateJobResponse> {
  const body = new CreateJobRequest({
    accountId,
    jobName: formData.define.jobName,
    cronSchedule: formData.define.cronSchedule,
    mappings: formData.schema.mappings.map((m) => {
      return new JobMapping({
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer: toTransformerConfigOptions(m.transformer, transformers),
        exclude: m.exclude,
      });
    }),
    source: new JobSource({
      connectionId: formData.flow.sourceId,
      options: new JobSourceOptions({
        config: {
          case: 'sqlOptions',
          value: new SqlSourceConnectionOptions({
            haltOnNewColumnAddition:
              formData.flow.sourceOptions.haltOnNewColumnAddition,
          }),
        },
      }),
    }),
    destinations: formData.flow.destinations.map((d) => {
      return new JobDestination({
        connectionId: d.connectionId,
        options: toJobDestinationOptions(
          d,
          connections.find((c) => c.id == d.connectionId)
        ),
      });
    }),
  });

  console.log('body', body);
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
