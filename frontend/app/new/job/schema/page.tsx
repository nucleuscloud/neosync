'use client';

import {
  SchemaTable,
  getConnectionSchema,
} from '@/components/jobs/SchemaTable/schema-table';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { SCHEMA_FORM_SCHEMA, SchemaFormValues } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { usePathname, useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps from '../JobsProgressSteps';
import { ConnectFormValues } from '../schema';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const { toast } = useToast();

  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';

  const [connectFormValues] = useSessionStorage<ConnectFormValues>(
    `${sessionPrefix}-new-job-connect`,
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
      const res = await getConnectionSchema(connectFormValues.sourceId);
      if (!res) {
        return { mappings: [] };
      }

      const mappings = res.schemas.map((r) => {
        return {
          ...r,
          transformer: {
            value: 'passthrough',
            config: { config: { case: '', value: {} } },
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

  async function onSubmit(_values: SchemaFormValues) {
    if (!account) {
      return;
    }
    router.push(`/new/job/subset?sessionId=${sessionPrefix}`);
  }

  const params = usePathname();
  const [stepName, _] = useState<string>(params.split('/').pop() ?? '');

  return (
    <div className="flex flex-col gap-20">
      <div className="mt-10">
        <JobsProgressSteps stepName={stepName} />
      </div>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <SchemaTable data={form.getValues().mappings} />
          <div className="flex flex-row gap-1 justify-between">
            <Button key="back" type="button" onClick={() => router.back()}>
              Back
            </Button>
            <Button key="submit" type="submit">
              Next
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
