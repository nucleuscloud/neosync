'use client';
import ButtonText from '@/components/ButtonText';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { convertNanosecondsToMinutes } from '@/util/util';
import {
  SchemaFormValues,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { Job, JobMappingTransformer, UserAccount } from '@neosync/sdk';
import { nanoid } from 'nanoid';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { GrClone } from 'react-icons/gr';
import {
  ConnectFormValues,
  DefineFormValues,
  NewJobType,
  SingleTableAiConnectFormValues,
  SingleTableAiSchemaFormValues,
  SingleTableConnectFormValues,
  SingleTableSchemaFormValues,
  SubsetFormValues,
} from '../../../new/job/schema';
import { getDefaultDestinationFormValues } from '../destinations/components/DestinationConnectionCard';
import {
  getSingleTableAiGenerateNumRows,
  getSingleTableAiSchemaTable,
} from '../source/components/AiDataGenConnectionCard';
import { getSingleTableGenerateNumRows } from '../source/components/DataGenConnectionCard';

interface Props {
  job: Job;
}

export default function JobCloneButton(props: Props): ReactElement {
  const { job } = props;
  const { account } = useAccount();
  const router = useRouter();

  function onCloneClick(): void {
    if (!account) {
      return;
    }
    const sessionId = nanoid();
    setDefaultNewJobFormValues(window.sessionStorage, job, sessionId);
    router.push(getJobCloneUrlFromJob(account, job, sessionId));
  }

  return (
    <Button onClick={onCloneClick}>
      <ButtonText text="Clone Job" leftIcon={<GrClone className="mr-1" />} />
    </Button>
  );
}

export function setDefaultNewJobFormValues(
  storage: Storage,
  job: Job,
  sessionId: string
): void {
  setDefaultDefineFormValues(storage, job, sessionId);
  setDefaultConnectFormValues(storage, job, sessionId);
  setDefaultSchemaFormValues(storage, job, sessionId);
  setDefaultSubsetFormValues(storage, job, sessionId);
}

export function getJobCloneUrlFromJob(
  account: UserAccount,
  job: Job,
  sessionId: string
): string {
  const urlParams = new URLSearchParams({
    jobType: getNewJobTypeFromJob(job),
    sessionId: sessionId,
  });
  return `/${account.name}/new/job?${urlParams.toString()}`;
}

function getNewJobTypeFromJob(job: Job): NewJobType {
  if (job.source?.options?.config.case === 'aiGenerate') {
    return 'ai-generate-table';
  }
  if (job.source?.options?.config.case === 'generate') {
    return 'generate-table';
  }
  return 'data-sync';
}

function setDefaultDefineFormValues(
  storage: Storage,
  job: Job,
  sessionPrefix: string
): void {
  const values: DefineFormValues = {
    jobName: `${job.name}-copy`,
    cronSchedule: job.cronSchedule,
    initiateJobRun: false,
    syncActivityOptions: job.syncOptions
      ? {
          retryPolicy: job.syncOptions.retryPolicy,
          scheduleToCloseTimeout: job.syncOptions.scheduleToCloseTimeout
            ? convertNanosecondsToMinutes(
                job.syncOptions.scheduleToCloseTimeout
              )
            : undefined,
          startToCloseTimeout: job.syncOptions.startToCloseTimeout
            ? convertNanosecondsToMinutes(job.syncOptions.startToCloseTimeout)
            : undefined,
        }
      : undefined,
    workflowSettings: job.workflowOptions
      ? {
          runTimeout: job.workflowOptions.runTimeout
            ? convertNanosecondsToMinutes(job.workflowOptions.runTimeout)
            : undefined,
        }
      : undefined,
  };
  storage.setItem(`${sessionPrefix}-new-job-define`, JSON.stringify(values));
}

function setDefaultConnectFormValues(
  storage: Storage,
  job: Job,
  sessionPrefix: string
): void {
  switch (job.source?.options?.config.case) {
    case 'aiGenerate': {
      const values: SingleTableAiConnectFormValues = {
        sourceId: job.source.options.config.value.aiConnectionId,
        fkSourceConnectionId:
          job.source.options.config.value.fkSourceConnectionId ?? '',
        destination:
          job.destinations.length > 0
            ? getDefaultDestinationFormValues(job.destinations[0])
            : {
                connectionId: '',
                destinationOptions: {},
              },
      };
      storage.setItem(
        `${sessionPrefix}-new-job-single-ai-table-connect`,
        JSON.stringify(values)
      );
      return;
    }
    case 'generate': {
      const values: SingleTableConnectFormValues = {
        fkSourceConnectionId:
          job.source.options.config.value.fkSourceConnectionId ?? '',
        destination:
          job.destinations.length > 0
            ? getDefaultDestinationFormValues(job.destinations[0])
            : {
                connectionId: '',
                destinationOptions: {},
              },
      };

      storage.setItem(
        `${sessionPrefix}-new-job-single-table-connect`,
        JSON.stringify(values)
      );
      return;
    }
    case 'mongodb': {
      const values: ConnectFormValues = {
        sourceId: job.source.options.config.value.connectionId,
        sourceOptions: {},
        destinations: job.destinations.map((dest) =>
          getDefaultDestinationFormValues(dest)
        ),
      };

      storage.setItem(
        `${sessionPrefix}-new-job-connect`,
        JSON.stringify(values)
      );
      return;
    }
    case 'mysql': {
      const values: ConnectFormValues = {
        sourceId: job.source.options.config.value.connectionId,
        sourceOptions: {
          haltOnNewColumnAddition:
            job.source.options.config.value.haltOnNewColumnAddition,
        },
        destinations: job.destinations.map((dest) =>
          getDefaultDestinationFormValues(dest)
        ),
      };

      storage.setItem(
        `${sessionPrefix}-new-job-connect`,
        JSON.stringify(values)
      );
      return;
    }
    case 'postgres': {
      const values: ConnectFormValues = {
        sourceId: job.source.options.config.value.connectionId,
        sourceOptions: {
          haltOnNewColumnAddition:
            job.source.options.config.value.haltOnNewColumnAddition,
        },
        destinations: job.destinations.map((dest) =>
          getDefaultDestinationFormValues(dest)
        ),
      };

      storage.setItem(
        `${sessionPrefix}-new-job-connect`,
        JSON.stringify(values)
      );
      return;
    }
  }
}

function setDefaultSchemaFormValues(
  storage: Storage,
  job: Job,
  sessionPrefix: string
): void {
  switch (job.source?.options?.config.case) {
    case 'aiGenerate': {
      const values: SingleTableAiSchemaFormValues = {
        numRows: getSingleTableAiGenerateNumRows(
          job.source.options.config.value
        ),
        userPrompt: job.source.options.config.value.userPrompt,
        model: job.source.options.config.value.modelName,
        ...getSingleTableAiSchemaTable(job.source.options.config.value),
      };

      storage.setItem(
        `${sessionPrefix}-new-job-single-table-schema`,
        JSON.stringify(values)
      );
      return;
    }
    case 'generate': {
      const values: SingleTableSchemaFormValues = {
        numRows: getSingleTableGenerateNumRows(job.source.options.config.value),
        mappings: job.mappings.map((mapping) => {
          return {
            ...mapping,
            transformer: mapping.transformer
              ? convertJobMappingTransformerToForm(mapping.transformer)
              : convertJobMappingTransformerToForm(new JobMappingTransformer()),
          };
        }),
      };

      storage.setItem(
        `${sessionPrefix}-new-job-single-table-schema`,
        JSON.stringify(values)
      );
      return;
    }
    case 'mysql':
    case 'mongodb':
    case 'postgres': {
      const values: SchemaFormValues = {
        connectionId: job.source.options.config.value.connectionId,
        mappings: job.mappings.map((mapping) => {
          return {
            ...mapping,
            transformer: mapping.transformer
              ? convertJobMappingTransformerToForm(mapping.transformer)
              : convertJobMappingTransformerToForm(new JobMappingTransformer()),
          };
        }),
      };

      storage.setItem(
        `${sessionPrefix}-new-job-schema`,
        JSON.stringify(values)
      );
      return;
    }
  }
}

function setDefaultSubsetFormValues(
  storage: Storage,
  job: Job,
  sessionPrefix: string
): void {
  switch (job.source?.options?.config.case) {
    case 'postgres':
    case 'mysql': {
      const values: SubsetFormValues = {
        subsets: job.source.options.config.value.schemas.flatMap(
          (schema): SubsetFormValues['subsets'] => {
            return schema.tables.map((table) => {
              return {
                schema: schema.schema,
                table: table.table,
                whereClause: table.whereClause,
              };
            });
          }
        ),
        subsetOptions: {
          subsetByForeignKeyConstraints:
            job.source.options.config.value.subsetByForeignKeyConstraints,
        },
      };
      storage.setItem(
        `${sessionPrefix}-new-job-subset`,
        JSON.stringify(values)
      );
      return;
    }
  }
}
