import {
  JobHook,
  JobHookConfig,
  JobHookConfig_JobSqlHook,
  JobHookConfig_JobSqlHook_Timing,
  JobHookTimingPostSync,
  JobHookTimingPreSync,
  NewJobHook,
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

const JobHookNameFormValue = yup
  .string()
  .required('Name is required')
  .min(3, 'Name must be at least 3 characters');

const JobHookDescriptionFormValue = yup
  .string()
  .required('Description is required');

const JobHookPriorityFormValue = yup
  .number()
  .required('Priority is required')
  .min(0, 'Priority must be between 0 and 100')
  .max(100, 'Priority must be between 0 and 100');

export const EditJobHookFormValues = yup.object({
  name: JobHookNameFormValue,
  description: JobHookDescriptionFormValue,
  priority: JobHookPriorityFormValue,
  enabled: yup.boolean(),
  hookType: HookTypeFormValue,
  sql: JobHookSqlFormValues.when('hookType', (values, schema) => {
    const [hooktype] = values;
    return hooktype === 'sql'
      ? schema.required('SQL config is required when hook type is sql')
      : schema;
  }),
});
export type EditJobHookFormValues = yup.InferType<typeof EditJobHookFormValues>;

export function toEditFormData(input: JobHook): EditJobHookFormValues {
  return {
    name: input.name,
    description: input.description,
    hookType: toHookType(input.config ?? new JobHookConfig()),
    priority: input.priority,
    enabled: input.enabled,
    sql: toSqlConfig(input.config ?? new JobHookConfig()),
  };
}

export const NewJobHookFormValues = yup.object({
  name: JobHookNameFormValue,
  description: JobHookDescriptionFormValue,
  priority: JobHookPriorityFormValue,
  enabled: yup.boolean(),
  hookType: HookTypeFormValue,
  sql: JobHookSqlFormValues.when('hookType', (values, schema) => {
    const [hooktype] = values;
    return hooktype === 'sql'
      ? schema.required('SQL config is required when hook type is sql')
      : schema;
  }),
});
export type NewJobHookFormValues = yup.InferType<typeof NewJobHookFormValues>;

function toSqlConfig(input: JobHookConfig): JobHookSqlFormValues {
  switch (input.config.case) {
    case 'sql': {
      return {
        connectionId: input.config.value.connectionId,
        query: input.config.value.query,
        timing: toSqlTimingConfig(
          input.config.value.timing ?? new JobHookConfig_JobSqlHook_Timing()
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
  return new JobHook({
    ...input,
    ...newFormDataToNewJobHook(values),
  });
}

export function newFormDataToNewJobHook(
  values: NewJobHookFormValues
): NewJobHook {
  return new NewJobHook({
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
      return new JobHookConfig({
        config: {
          case: 'sql',
          value: new JobHookConfig_JobSqlHook({
            connectionId: values.sql.connectionId,
            query: values.sql.query,
            timing: toJobHookSqlTimingConfig(values.sql.timing),
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
      return new JobHookConfig_JobSqlHook_Timing({
        timing: {
          case: 'preSync',
          value: new JobHookTimingPreSync(),
        },
      });
    }
    case 'postSync': {
      return new JobHookConfig_JobSqlHook_Timing({
        timing: {
          case: 'postSync',
          value: new JobHookTimingPostSync(),
        },
      });
    }
    default: {
      return new JobHookConfig_JobSqlHook_Timing({
        timing: {
          case: 'preSync',
          value: new JobHookTimingPreSync(),
        },
      });
    }
  }
}
