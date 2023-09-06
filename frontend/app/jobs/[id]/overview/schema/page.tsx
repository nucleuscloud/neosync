'use client';
import {
  SchemaTable,
  getConnectionSchema,
} from '@/app/jobs/components/SchemaTable/schema-table';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { useToast } from '@/components/ui/use-toast';
import {
  JobMapping,
  UpdateJobMappingsRequest,
  UpdateJobMappingsResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import { getJob } from '../util';

const JOB_MAPPING_SCHEMA = Yup.object({
  schema: Yup.string().required(),
  table: Yup.string().required(),
  column: Yup.string().required(),
  dataType: Yup.string().required(),
  transformer: Yup.string().required(),
  exclude: Yup.boolean(),
}).required();
type JobMappingFormValues = Yup.InferType<typeof JOB_MAPPING_SCHEMA>;

const SCHEMA_FORM_SCHEMA = Yup.object({
  mappings: Yup.array().of(JOB_MAPPING_SCHEMA).required(),
});
type SchemaFormValues = Yup.InferType<typeof SCHEMA_FORM_SCHEMA>;

interface SchemaMap {
  [schema: string]: {
    [table: string]: {
      [column: string]: {
        dataType: string;
      };
    };
  };
}

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { toast } = useToast();

  const form = useForm({
    resolver: yupResolver<SchemaFormValues>(SCHEMA_FORM_SCHEMA),
    defaultValues: async () => getMappings(id),
  });
  async function onSubmit(values: SchemaFormValues) {
    try {
      await updateJobMappings(id, values.mappings);
      toast({
        title: 'Successfully updated job mappings!',
        variant: 'default',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update job mappings',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

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

async function getMappings(jobId?: string): Promise<SchemaFormValues> {
  if (!jobId) {
    return { mappings: [] };
  }
  const jobRes = await getJob(jobId);
  if (!jobRes) {
    return { mappings: [] };
  }
  const job = jobRes?.job;

  const res = await getConnectionSchema(job?.connectionSourceId);
  if (!res) {
    return { mappings: [] };
  }

  const schemaMap: SchemaMap = {};
  res.schemas.forEach((c) => {
    if (!schemaMap[c.schema]) {
      schemaMap[c.schema] = {
        [c.table]: {
          [c.column]: {
            dataType: c.dataType,
          },
        },
      };
    } else if (!schemaMap[c.schema][c.table]) {
      schemaMap[c.schema][c.table] = {
        [c.column]: {
          dataType: c.dataType,
        },
      };
    } else {
      schemaMap[c.schema][c.table][c.column] = { dataType: c.dataType };
    }
  });

  const mappings = job?.mappings.map((r) => {
    const datatype = schemaMap[r.schema][r.table][r.column].dataType;
    return {
      ...r,
      transformer: r.transformer as unknown as string,
      dataType: datatype || '',
    };
  });
  return { mappings: mappings || [] };
}

async function updateJobMappings(
  jobId: string,
  mappings: JobMappingFormValues[]
): Promise<UpdateJobMappingsResponse> {
  const res = await fetch(`/api/jobs/${jobId}/mappings`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new UpdateJobMappingsRequest({
        id: jobId,
        mappings: mappings.map((m) => {
          return new JobMapping({
            schema: m.schema,
            table: m.table,
            column: m.column,
            transformer: m.transformer,
            exclude: m.exclude,
          });
        }),
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobMappingsResponse.fromJson(await res.json());
}
