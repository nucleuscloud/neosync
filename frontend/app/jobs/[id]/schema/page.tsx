'use client';
import SubPageHeader from '@/components/headers/SubPageHeader';
import {
  SchemaTable,
  getConnectionSchema,
} from '@/components/jobs/SchemaTable/schema-table';
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
import {
  JobMappingFormValues,
  SCHEMA_FORM_SCHEMA,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { getJob } from '../util';

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
      <SubPageHeader header="Schema" description="Manage job schema" />

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <SchemaTable data={form.getValues().mappings} />

          <div className="flex flex-row gap-1 justify-end">
            <Button disabled={!form.formState.isDirty} type="submit">
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

  const res = await getConnectionSchema(job?.source?.connectionId);
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
