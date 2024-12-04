import { FormEvent, ReactElement, useEffect, useMemo } from 'react';
import * as yup from 'yup';
import { create } from 'zustand';

import ConnectionSelectContent from '@/app/(mgmt)/[account]/new/job/connect/ConnectionSelectContent';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { splitConnections } from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { Editor } from '@monaco-editor/react';
import {
  JobHook,
  JobHookConfig,
  JobHookConfig_JobSqlHook,
  JobHookConfig_JobSqlHook_Timing,
  JobHookTimingPostSync,
  JobHookTimingPreSync,
} from '@neosync/sdk';
import { getConnections } from '@neosync/sdk/connectquery';
import { editor } from 'monaco-editor';
import { useTheme } from 'next-themes';

const SqlTimingFormValues = yup
  .string()
  .oneOf(['preSync', 'postSync'])
  .required('Must select preSync or postSync');
type SqlTimingFormValues = yup.InferType<typeof SqlTimingFormValues>;

const JobHookSqlFormValues = yup.object({
  query: yup.string().required('SQL query is required'),
  connectionId: yup.string().uuid().required('Connection ID is required'),
  timing: SqlTimingFormValues,
});
type JobHookSqlFormValues = yup.InferType<typeof JobHookSqlFormValues>;

const HookTypeFormValues = yup
  .string()
  .oneOf(['sql'], 'Only SQL hooks are currently supported')
  .required('Hook type is required');
type HookTypeFormValues = yup.InferType<typeof HookTypeFormValues>;

const EditJobHookFormValues = yup.object().shape({
  name: yup
    .string()
    .required('Name is required')
    .min(3, 'Name must be at least 3 characters'),
  description: yup.string(),
  priority: yup
    .number()
    .required('Priority is required')
    .min(0, 'Priority must be between 0 and 100')
    .max(100, 'Priority must be between 0 and 100'),
  enabled: yup.boolean(),
  hookType: HookTypeFormValues,
  sql: JobHookSqlFormValues.when('hookType', (values, schema) => {
    const [hooktype] = values;
    return hooktype === 'sql'
      ? schema.required('SQL config is required when hook type is sql')
      : schema;
  }),
});
type EditJobHookFormValues = yup.InferType<typeof EditJobHookFormValues>;

interface EditHookStore {
  formData: EditJobHookFormValues;
  errors: Record<string, string>;
  isSubmitting: boolean;
  setFormData: (data: Partial<EditJobHookFormValues>) => void;
  setErrors: (errors: Record<string, string>) => void;
  setSubmitting: (isSubmitting: boolean) => void;
  resetForm: () => void;
}

const useEditHookStore = create<EditHookStore>((set) => ({
  formData: {
    hookType: 'sql',
    name: 'my-initial-job-hook',
    priority: 0,
    sql: { query: 'INITIAL FORM VALUE', timing: 'preSync', connectionId: '' },
  },
  errors: {},
  isSubmitting: false,
  setFormData: (data) =>
    set((state) => ({ formData: { ...state.formData, ...data } })),
  setErrors: (errors) => set({ errors }),
  setSubmitting: (isSubmitting) => set({ isSubmitting }),
  resetForm: () =>
    set({
      formData: {
        hookType: 'sql',
        name: 'my-initial-job-hook',
        priority: 0,
        sql: { query: 'RESET FORM VALUE', timing: 'preSync', connectionId: '' },
      },
      errors: {},
      isSubmitting: false,
    }),
}));

interface EditHookFormProps {
  hook: JobHook;
  onSubmit: (values: JobHook) => Promise<void>;

  jobConnectionIds: string[];
}

function toFormData(input: JobHook): EditJobHookFormValues {
  return {
    name: input.name,
    description: input.description,
    hookType: toHookType(input.config ?? new JobHookConfig()),
    priority: input.priority,
    enabled: input.enabled,
    sql: toSqlConfig(input.config ?? new JobHookConfig()),
  };
}

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
): SqlTimingFormValues {
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

function toHookType(input: JobHookConfig): HookTypeFormValues {
  switch (input.config.case) {
    case 'sql': {
      return 'sql';
    }
    default: {
      return 'sql';
    }
  }
}

function formDataToJobHook(
  input: JobHook,
  values: EditJobHookFormValues
): JobHook {
  return new JobHook({
    ...input,
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

export function EditHookForm({
  hook,
  onSubmit,
  jobConnectionIds,
}: EditHookFormProps) {
  const {
    formData,
    errors,
    setFormData,
    setErrors,
    setSubmitting,
    isSubmitting,
  } = useEditHookStore();

  useEffect(() => {
    // Initialize form with hook data
    const formData = toFormData(hook);
    setFormData(formData);
  }, [hook, setFormData]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    try {
      setSubmitting(true);
      setErrors({});

      const validatedData = await EditJobHookFormValues.validate(formData, {
        abortEarly: false,
      });

      await onSubmit(formDataToJobHook(hook, validatedData));
    } catch (err) {
      if (err instanceof yup.ValidationError) {
        const validationErrors: Record<string, string> = {};
        err.inner.forEach((error) => {
          if (error.path) {
            validationErrors[error.path] = error.message;
          }
        });
        setErrors(validationErrors);
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="name">Name</Label>
        <Input
          id="name"
          autoCapitalize="off" // we don't allow capitals in team names
          data-1p-ignore // tells 1password extension to not autofill this field
          value={formData.name || ''}
          onChange={(e) => setFormData({ name: e.target.value })}
          placeholder="Hook name"
        />
        {errors.name && <p className="text-sm text-red-500">{errors.name}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          value={formData.description || ''}
          onChange={(e) => setFormData({ description: e.target.value })}
          placeholder="Hook description"
          rows={3}
        />
        {errors.description && (
          <p className="text-sm text-red-500">{errors.description}</p>
        )}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="priority">Priority (0-100)</Label>
          <Input
            id="priority"
            type="number"
            value={formData.priority}
            onChange={(e) => setFormData({ priority: e.target.valueAsNumber })}
          />
          {errors.priority && (
            <p className="text-sm text-red-500">{errors.priority}</p>
          )}
        </div>
      </div>

      <div className="flex items-center gap-4">
        <Label htmlFor="enabled">Enabled</Label>
        <Switch
          id="enabled"
          checked={formData.enabled || false}
          onCheckedChange={(checked) => setFormData({ enabled: checked })}
        />
      </div>

      <div className="flex flex-col gap-4">
        <Label htmlFor="hookType">Hook Type</Label>
        <ToggleGroup
          className="flex justify-start"
          type="single"
          onValueChange={(value) => {
            if (value) {
              setFormData({ hookType: value as HookTypeFormValues });
            }
          }}
          value={formData.hookType}
        >
          <ToggleGroupItem value="sql">SQL</ToggleGroupItem>
        </ToggleGroup>
      </div>

      <div className="flex flex-col gap-4">
        {formData.hookType === 'sql' && (
          <JobConfigSqlForm
            values={formData.sql}
            setValues={(newSqlData) => {
              setFormData({ sql: newSqlData });
            }}
            jobConnectionIds={jobConnectionIds}
          />
        )}
      </div>

      <div className="flex justify-end gap-3">
        <Button
          type="submit"
          disabled={isSubmitting}
          className="w-full sm:w-auto"
        >
          {isSubmitting ? 'Saving...' : 'Save Changes'}
        </Button>
      </div>
    </form>
  );
}

interface JobConfigFormProps {
  values: JobHookSqlFormValues;
  setValues(values: JobHookSqlFormValues): void;
  jobConnectionIds: string[];
}
function JobConfigSqlForm(props: JobConfigFormProps): ReactElement {
  const { values, setValues, jobConnectionIds } = props;
  return (
    <>
      <SelectConnections
        connectionIds={jobConnectionIds}
        selectedConnectionId={values.connectionId}
        setSelectedConnectionId={(updatedId) => {
          setValues({ ...values, connectionId: updatedId });
        }}
      />
      <div className="flex flex-col gap-3">
        <Label htmlFor="query">Query</Label>
        <EditSqlQuery
          query={values.query}
          setQuery={(query) => setValues({ ...values, query })}
        />
      </div>
      <div className="flex flex-col gap-3">
        <Label htmlFor="timing">Timing</Label>
        <Select
          name="timing"
          value={values.timing}
          onValueChange={(newValue) => {
            if (newValue) {
              setValues({ ...values, timing: newValue as SqlTimingFormValues });
            }
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select a timing value" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="preSync">Pre Sync</SelectItem>
            <SelectItem value="postSync">Post Sync</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </>
  );
}

interface EditSqlQueryProps {
  query: string;
  setQuery(query: string): void;
}

const sqlEditorOptions: editor.IStandaloneEditorConstructionOptions = {
  minimap: { enabled: false },
  wordWrap: 'on',
  lineNumbers: 'off',
};

function EditSqlQuery(props: EditSqlQueryProps): ReactElement {
  const { query, setQuery } = props;

  const { resolvedTheme } = useTheme();

  return (
    <div>
      <Editor
        height="5vh"
        width="100%"
        language="sql"
        theme={resolvedTheme === 'dark' ? 'vs-dark' : 'cobalt'}
        options={sqlEditorOptions}
        value={query}
        onChange={(updatedValue) => setQuery(updatedValue ?? '')}
      />
    </div>
  );
}

interface SelectConnectionsProps {
  connectionIds: string[];

  selectedConnectionId: string;
  setSelectedConnectionId(id: string): void;
}
function SelectConnections(props: SelectConnectionsProps): ReactElement {
  const { connectionIds, selectedConnectionId, setSelectedConnectionId } =
    props;
  const { account } = useAccount();

  const { data: connectionsResp, isLoading } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const { postgres, mysql, mssql } = useMemo(() => {
    const connections = connectionsResp?.connections ?? [];
    const uniqueConnectionIds = new Set(connectionIds);
    const filtered = connections.filter((c) => uniqueConnectionIds.has(c.id));
    return splitConnections(filtered);
  }, [connectionsResp?.connections, isLoading, connectionIds]);

  if (isLoading) {
    return <Skeleton />;
  }

  return (
    <div className="flex flex-col gap-3">
      <Label htmlFor="connectionId">Connection</Label>
      <Select
        name="connectionId"
        value={selectedConnectionId}
        onValueChange={(newValue) => {
          if (newValue) {
            setSelectedConnectionId(newValue);
          }
        }}
      >
        <SelectTrigger>
          <SelectValue placeholder="Select a connection..." />
        </SelectTrigger>
        <SelectContent>
          <ConnectionSelectContent
            postgres={postgres}
            mysql={mysql}
            mssql={mssql}
          />
        </SelectContent>
      </Select>
    </div>
  );
}
