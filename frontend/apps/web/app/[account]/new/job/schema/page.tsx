'use client';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { SchemaTable } from '@/components/jobs/SchemaTable/schema-table';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { useGetConnectionSchema } from '@/libs/hooks/useGetConnectionSchema';
import {
  SCHEMA_FORM_SCHEMA,
  SchemaFormValues,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  DatabaseColumn,
  JobMappingTransformer,
  Passthrough,
  TransformerConfig,
} from '@neosync/sdk';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, { DATA_SYNC_STEPS } from '../JobsProgressSteps';
import { ConnectFormValues } from '../schema';

const isBrowser = () => typeof window !== 'undefined';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();

  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
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

  const [schemaFormData] = useSessionStorage<SchemaFormValues>(
    `${sessionPrefix}-new-job-schema`,
    {
      mappings: [],
    }
  );

  const { data: connectionSchemaData } = useGetConnectionSchema(
    account?.id ?? '',
    connectFormValues.sourceId
  );

  const form = useForm<SchemaFormValues>({
    resolver: yupResolver<SchemaFormValues>(SCHEMA_FORM_SCHEMA),
    values: getFormValues(connectionSchemaData?.schemas ?? [], schemaFormData),
  });

  useFormPersist(`${sessionPrefix}-new-job-schema`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });

  async function onSubmit(_values: SchemaFormValues) {
    if (!account) {
      return;
    }
    router.push(`/${account?.name}/new/job/subset?sessionId=${sessionPrefix}`);
  }

  return (
    <div className="flex flex-col gap-5">
      <OverviewContainer
        Header={
          <PageHeader
            header="Schema"
            progressSteps={
              <JobsProgressSteps steps={DATA_SYNC_STEPS} stepName={'schema'} />
            }
          />
        }
        containerClassName="connect-page"
      >
        <div />
      </OverviewContainer>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <SchemaTable data={form.watch().mappings} />
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

function getFormValues(
  dbCols: DatabaseColumn[],
  existingData: SchemaFormValues | undefined
): SchemaFormValues {
  const existingMappings = existingData?.mappings ?? [];
  if (existingData && existingMappings.length > 0) {
    return existingData;
  }

  return {
    mappings: dbCols.map((r) => {
      return {
        ...r,
        transformer: convertJobMappingTransformerToForm(
          new JobMappingTransformer({
            source: 'passthrough',
            config: new TransformerConfig({
              config: {
                case: 'passthroughConfig',
                value: new Passthrough({}),
              },
            }),
          })
        ),
      };
    }),
  };
}
