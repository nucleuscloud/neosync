'use client';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { FormError } from '@/components/jobs/SchemaTable/FormErrorsCard';
import {
  SchemaTable,
  extractAllFormErrors,
} from '@/components/jobs/SchemaTable/SchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { toast } from '@/components/ui/use-toast';
import { useGetConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { useGetConnectionTableConstraints } from '@/libs/hooks/useGetConnectionTableConstraints';
import {
  JobMappingFormValues,
  SCHEMA_FORM_SCHEMA,
  SchemaFormValues,
  convertJobMappingTransformerFormToJobMappingTransformer,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  DatabaseColumn,
  ForeignConstraintTables,
  JobMapping,
  JobMappingTransformer,
  PrimaryConstraint,
  TransformerConfig,
  ValidateJobMappingsRequest,
  ValidateJobMappingsResponse,
} from '@neosync/sdk';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { FieldErrors, useFieldArray, useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import { getOnSelectedTableToggle } from '../../../jobs/[id]/source/components/util';
import JobsProgressSteps, { getJobProgressSteps } from '../JobsProgressSteps';
import { ConnectFormValues, SingleTableSchemaFormValues } from '../schema';

const isBrowser = () => typeof window !== 'undefined';

export interface ColumnMetadata {
  pk: { [key: string]: PrimaryConstraint };
  fk: { [key: string]: ForeignConstraintTables };
  isNullable: DatabaseColumn[];
}

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();

  const [validateMappingsResponse, setValidateMappingsResponse] = useState<
    ValidateJobMappingsResponse | undefined
  >();

  const [isValidatingMappings, setIsValidatingMappings] = useState(false);

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
      connectionId: '', // hack to track if source id changes
    }
  );

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaMapLoading,
    isValidating: isSchemaMapValidating,
  } = useGetConnectionSchemaMap(account?.id ?? '', connectFormValues.sourceId);

  const { data: tableConstraints, isValidating: isTableConstraintsValidating } =
    useGetConnectionTableConstraints(
      account?.id ?? '',
      connectFormValues.sourceId
    );

  const form = useForm<SchemaFormValues>({
    resolver: yupResolver<SchemaFormValues>(SCHEMA_FORM_SCHEMA),
    values: getFormValues(connectFormValues.sourceId, schemaFormData),
    context: { accountId: account?.id },
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

  async function validateMappings() {
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        connectFormValues,
        formMappings,
        account?.id || ''
      );
      setValidateMappingsResponse(res);
    } catch (error) {
      console.error('Failed to validate job mappings:', error);
      toast({
        title: 'Unable to validate job mappings',
        variant: 'destructive',
      });
    } finally {
      setIsValidatingMappings(false);
    }
  }

  const schemaConstraintHandler = useMemo(
    () =>
      getSchemaConstraintHandler(
        connectionSchemaDataMap?.schemaMap ?? {},
        tableConstraints?.primaryKeyConstraints ?? {},
        tableConstraints?.foreignKeyConstraints ?? {},
        tableConstraints?.uniqueConstraints ?? {}
      ),
    [isSchemaMapValidating, isTableConstraintsValidating]
  );
  const [selectedTables, setSelectedTables] = useState<Set<string>>(new Set());

  const { append, remove, fields } = useFieldArray<SchemaFormValues>({
    control: form.control,
    name: 'mappings',
  });
  const onSelectedTableToggle = getOnSelectedTableToggle(
    connectionSchemaDataMap?.schemaMap ?? {},
    selectedTables,
    setSelectedTables,
    fields,
    remove,
    append
  );

  useEffect(() => {
    const validateJobMappings = async () => {
      try {
        setIsValidatingMappings(true);
        const res = await validateJobMapping(
          connectFormValues,
          formMappings,
          account?.id || ''
        );
        setValidateMappingsResponse(res);
      } catch (error) {
        console.error('Failed to validate job mappings:', error);
        toast({
          title: 'Unable to validate job mappings',
          variant: 'destructive',
        });
      } finally {
        setIsValidatingMappings(false);
      }
    };

    validateJobMappings();
  }, [selectedTables]);

  useEffect(() => {
    if (
      isSchemaMapLoading ||
      selectedTables.size > 0 ||
      !connectFormValues.sourceId
    ) {
      return;
    }
    const js = getFormValues(connectFormValues.sourceId, schemaFormData);
    setSelectedTables(
      new Set(
        js.mappings.map((mapping) => `${mapping.schema}.${mapping.table}`)
      )
    );
  }, [isSchemaMapLoading, connectFormValues.sourceId]);

  const formMappings = form.watch('mappings');
  return (
    <div className="flex flex-col gap-5">
      <OverviewContainer
        Header={
          <PageHeader
            header="Schema"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps('data-sync')}
                stepName={'schema'}
              />
            }
          />
        }
        containerClassName="connect-page"
      >
        <div />
      </OverviewContainer>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <SchemaTable
            data={formMappings}
            jobType="sync"
            constraintHandler={schemaConstraintHandler}
            schema={connectionSchemaDataMap?.schemaMap ?? {}}
            isSchemaDataReloading={isSchemaMapValidating}
            isJobMappingsValidating={isValidatingMappings}
            selectedTables={selectedTables}
            onSelectedTableToggle={onSelectedTableToggle}
            formErrors={getAllFormErrors(
              form.formState.errors,
              formMappings,
              validateMappingsResponse
            )}
            onValidate={validateMappings}
          />
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

function getAllFormErrors(
  formErrors: FieldErrors<SchemaFormValues | SingleTableSchemaFormValues>,
  values: JobMappingFormValues[],
  validationErrors: ValidateJobMappingsResponse | undefined
): FormError[] {
  let messages: FormError[] = [];
  const formErr = extractAllFormErrors(formErrors, values);
  if (!validationErrors) {
    return formErr;
  }
  const colErr = validationErrors.columnErrors.map((e) => {
    return {
      path: `${e.schema}.${e.table}.${e.column}`,
      message: e.errors.join('. '),
    };
  });
  const dbErr = validationErrors.databaseErrors?.errors.map((e) => {
    return {
      path: '',
      message: e,
    };
  });
  messages = messages.concat(colErr, formErr);
  if (dbErr) {
    messages = messages.concat(dbErr);
  }

  return messages;
}

function getFormValues(
  connectionId: string,
  existingData: SchemaFormValues | undefined
): SchemaFormValues {
  const existingMappings = existingData?.mappings ?? [];
  if (
    existingData &&
    existingMappings.length > 0 &&
    existingData.connectionId === connectionId
  ) {
    return existingData;
  }

  return {
    mappings: [],
    connectionId,
  };
}

async function validateJobMapping(
  connectFormValues: ConnectFormValues,
  formMappings: JobMappingFormValues[],
  accountId: string
): Promise<ValidateJobMappingsResponse> {
  console.log(JSON.stringify(formMappings, undefined, 2));
  const body = new ValidateJobMappingsRequest({
    accountId,
    mappings: formMappings.map((m) => {
      return new JobMapping({
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer:
          m.transformer.source != 0
            ? convertJobMappingTransformerFormToJobMappingTransformer(
                m.transformer
              )
            : new JobMappingTransformer({
                source: 1,
                config: new TransformerConfig({
                  config: {
                    case: 'passthroughConfig',
                    value: {},
                  },
                }),
              }),
      });
    }),
    sourceConnectionId: connectFormValues.sourceId,
  });

  const res = await fetch(`/api/accounts/${accountId}/jobs/validate-mappings`, {
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

  return ValidateJobMappingsResponse.fromJson(await res.json());
}
