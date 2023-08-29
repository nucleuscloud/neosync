'use client';
import {
  SchemaTable,
  getConnectionSchema,
} from '@/app/jobs/components/SchemaForm/schema-form';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';

const JOB_MAPPING_SCHEMA = Yup.object({
  schema: Yup.string().required(),
  table: Yup.string().required(),
  column: Yup.string().required(),
  dataType: Yup.string().required(),
  transformer: Yup.string().required(),
}).required();
export type JobMappingFormValues = Yup.InferType<typeof JOB_MAPPING_SCHEMA>;

export const SCHEMA_FORM_SCHEMA = Yup.object({
  mappings: Yup.array().of(JOB_MAPPING_SCHEMA).required(),
});
export type SchemaFormValues = Yup.InferType<typeof SCHEMA_FORM_SCHEMA>;

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading } = useGetJob(id);

  const form = useForm({
    resolver: yupResolver<SchemaFormValues>(SCHEMA_FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getConnectionSchema(data?.job?.connectionSourceId);
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
  async function onSubmit(_values: SchemaFormValues) {}

  return (
    <div className="job-details-container">
      <PageHeader header="Schema" description="Manage job schema" />
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <SchemaTable data={form.getValues().mappings} />

          <div className="flex flex-row gap-1 justify-end">
            <Button key="submit" type="submit">
              Save
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
