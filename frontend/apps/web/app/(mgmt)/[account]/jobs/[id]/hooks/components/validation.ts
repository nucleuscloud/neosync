import { RESOURCE_NAME_REGEX } from '@/yup-validations/connections';
import { create } from '@bufbuild/protobuf';
import {
  JobHook,
  JobHookConfig,
  JobHookConfig_JobSqlHook_Timing,
  JobHookConfig_JobSqlHook_TimingSchema,
  JobHookConfig_JobSqlHookSchema,
  JobHookConfigSchema,
  JobHookSchema,
  JobHookTimingPostSyncSchema,
  JobHookTimingPreSyncSchema,
  NewJobHook,
  NewJobHookSchema,
} from '@neosync/sdk';
import * as yup from 'yup';

const SqlTimingFormValue = yup
  .string()
  .oneOf(['preSync', 'postSync'])
  .required('Must select preSync or postSync');
export type SqlTimingFormValue = yup.InferType<typeof SqlTimingFormValue>;

const JobHookSqlFormValues = yup.object({
  query: yup.string().required('SQL query is required'),
  connectionId: yup.string().uuid().required('Connection ID is required'),
  timing: SqlTimingFormValue,
});
export type JobHookSqlFormValues = yup.InferType<typeof JobHookSqlFormValues>;

const HookTypeFormValue = yup
  .string()
  .oneOf(['sql'], 'Only SQL hooks are currently supported')
  .required('Hook type is required');
export type HookTypeFormValue = yup.InferType<typeof HookTypeFormValue>;

const EnabledFormValue = yup
  .boolean()
  .required('Must provide an enabled value');

const JobHookNameFormValue = yup
  .string()
  .required('Name is required')
  .min(3, 'Name must be at least 3 characters')
  .max(100, 'The Hook name must be at most 100 characters')
  .test(
    'resourceName',
    'Name must be between 3-100 characters and may only include lowercase letters, numbers, and hyphens',
    (value) => {
      return RESOURCE_NAME_REGEX.test(value);
    }
  );

const JobHookDescriptionFormValue = yup
  .string()
  .required('Description is required');

const JobHookPriorityFormValue = yup
  .number()
  .required('Priority is required')
  .min(0, 'Priority must be between 0 and 100')
  .max(100, 'Priority must be between 0 and 100');

const JobHookConfigFormValues = yup.object({
  sql: JobHookSqlFormValues.when('hookType', (values, schema) => {
    const [hooktype] = values;
    return hooktype === 'sql'
      ? schema.required('SQL config is required when hook type is sql')
      : schema;
  }),
});
export type JobHookConfigFormValues = yup.InferType<
  typeof JobHookConfigFormValues
>;

export const EditJobHookFormValues = yup.object({
  name: JobHookNameFormValue,
  description: JobHookDescriptionFormValue,
  priority: JobHookPriorityFormValue,
  enabled: EnabledFormValue,
  hookType: HookTypeFormValue,
  config: JobHookConfigFormValues,
});
export type EditJobHookFormValues = yup.InferType<typeof EditJobHookFormValues>;

export function toEditFormData(input: JobHook): EditJobHookFormValues {
  return {
    name: input.name,
    description: input.description,
    hookType: toHookType(input.config ?? create(JobHookConfigSchema)),
    priority: input.priority,
    enabled: input.enabled,
    config: {
      sql: toSqlConfig(input.config ?? create(JobHookConfigSchema)),
    },
  };
}

export const NewJobHookFormValues = yup.object({
  name: JobHookNameFormValue,
  description: JobHookDescriptionFormValue,
  priority: JobHookPriorityFormValue,
  enabled: EnabledFormValue,
  hookType: HookTypeFormValue,
  config: JobHookConfigFormValues,
});
export type NewJobHookFormValues = yup.InferType<typeof NewJobHookFormValues>;

function toSqlConfig(input: JobHookConfig): JobHookSqlFormValues {
  switch (input.config.case) {
    case 'sql': {
      return {
        connectionId: input.config.value.connectionId,
        query: input.config.value.query,
        timing: toSqlTimingConfig(
          input.config.value.timing ??
            create(JobHookConfig_JobSqlHook_TimingSchema)
        ),
      };
    }
    default: {
      return {
        connectionId: '',
        query: 'DEFAULT QUERY',
        timing: 'preSync',
      };
    }
  }
}
function toSqlTimingConfig(
  input: JobHookConfig_JobSqlHook_Timing
): SqlTimingFormValue {
  switch (input.timing.case) {
    case 'preSync': {
      return 'preSync';
    }
    case 'postSync': {
      return 'postSync';
    }
    default: {
      return 'preSync';
    }
  }
}

function toHookType(input: JobHookConfig): HookTypeFormValue {
  switch (input.config.case) {
    case 'sql': {
      return 'sql';
    }
    default: {
      return 'sql';
    }
  }
}

export function editFormDataToJobHook(
  input: JobHook,
  values: EditJobHookFormValues
): JobHook {
  const newValues = newFormDataToNewJobHook(values);
  return create(JobHookSchema, {
    ...input,
    name: newValues.name,
    description: newValues.description,
    enabled: newValues.enabled,
    priority: newValues.priority,
    config: newValues.config,
  });
}

export function newFormDataToNewJobHook(
  values: NewJobHookFormValues
): NewJobHook {
  return create(NewJobHookSchema, {
    name: values.name,
    description: values.description,
    enabled: values.enabled,
    priority: values.priority,
    config: toJobHookConfig(values),
  });
}

function toJobHookConfig(
  values: EditJobHookFormValues
): JobHookConfig | undefined {
  switch (values.hookType) {
    case 'sql': {
      return create(JobHookConfigSchema, {
        config: {
          case: 'sql',
          value: create(JobHookConfig_JobSqlHookSchema, {
            connectionId: values.config.sql.connectionId,
            query: values.config.sql.query,
            timing: toJobHookSqlTimingConfig(values.config.sql.timing),
          }),
        },
      });
    }
  }
}

function toJobHookSqlTimingConfig(
  values: JobHookSqlFormValues['timing']
): JobHookConfig_JobSqlHook_Timing {
  switch (values) {
    case 'preSync': {
      return create(JobHookConfig_JobSqlHook_TimingSchema, {
        timing: {
          case: 'preSync',
          value: create(JobHookTimingPreSyncSchema),
        },
      });
    }
    case 'postSync': {
      return create(JobHookConfig_JobSqlHook_TimingSchema, {
        timing: {
          case: 'postSync',
          value: create(JobHookTimingPostSyncSchema),
        },
      });
    }
    default: {
      return create(JobHookConfig_JobSqlHook_TimingSchema, {
        timing: {
          case: 'preSync',
          value: create(JobHookTimingPreSyncSchema),
        },
      });
    }
  }
}
