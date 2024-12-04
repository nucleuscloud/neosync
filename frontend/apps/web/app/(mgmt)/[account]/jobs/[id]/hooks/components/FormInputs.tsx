import FormErrorMessage from '@/components/FormErrorMessage';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { ReactElement } from 'react';
import FormHeader from './FormHeader';
import JobConfigSqlForm from './JobConfigSqlForm';
import { HookTypeFormValue, JobHookConfigFormValues } from './validation';

interface NameProps {
  error?: string;
  value: string;
  onChange(value: string): void;
}

export function Name(props: NameProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="space-y-2">
      <FormHeader
        htmlFor="name"
        title="Name"
        description="Name of the hook for display and reference, must be unique"
        isErrored={!!error}
      />
      <Input
        id="name"
        autoCapitalize="off" // we don't allow capitals in team names
        data-1p-ignore // tells 1password extension to not autofill this field
        value={value || ''}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Hook name"
      />
      <FormErrorMessage message={error} />
    </div>
  );
}

interface DescriptionProps {
  error?: string;
  value: string;
  onChange(value: string): void;
}

export function Description(props: DescriptionProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="space-y-2">
      <FormHeader
        htmlFor="description"
        title="Description"
        description="What this hook does"
        isErrored={!!error}
      />
      <Textarea
        id="description"
        value={value || ''}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Hook description"
        rows={3}
      />
      <FormErrorMessage message={error} />
    </div>
  );
}

interface PriorityProps {
  error?: string;
  value: number;
  onChange(value: number): void;
}

export function Priority(props: PriorityProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="grid grid-cols-2 gap-4">
      <div className="space-y-2">
        <FormHeader
          htmlFor="priority"
          title="Priority (0-100)"
          description="Determines execution order. Lower values are higher priority"
          isErrored={!!error}
        />
        <Input
          id="priority"
          type="number"
          value={value}
          onChange={(e) => onChange(e.target.valueAsNumber)}
        />
        <FormErrorMessage message={error} />
      </div>
    </div>
  );
}

interface EnabledProps {
  error?: string;
  value: boolean;
  onChange(value: boolean): void;
}

export function Enabled(props: EnabledProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="flex items-center gap-4">
      <FormHeader
        htmlFor="enabled"
        title="Enabled"
        description="Whether or not this hook will be invoked during a job run"
        isErrored={!!error}
      />
      <Switch
        id="enabled"
        checked={value}
        onCheckedChange={(checked) => onChange(checked)}
      />
      <FormErrorMessage message={error} />
    </div>
  );
}

interface HookTypeProps {
  error?: string;
  value: HookTypeFormValue;
  onChange(value: HookTypeFormValue): void;
}

export function HookType(props: HookTypeProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="flex flex-col gap-4">
      <FormHeader
        title="Hook Type"
        description="The type of hook. Currently only SQL hooks are supported"
        isErrored={!!error}
      />
      <ToggleGroup
        className="flex justify-start"
        type="single"
        onValueChange={(value) => {
          if (value) {
            onChange(value as HookTypeFormValue);
          }
        }}
        value={value}
      >
        <ToggleGroupItem value="sql">SQL</ToggleGroupItem>
      </ToggleGroup>
      <FormErrorMessage message={error} />
    </div>
  );
}

interface JobConfigProps {
  errors: Record<string, string>;
  hookType: HookTypeFormValue;
  value: JobHookConfigFormValues;
  onChange(value: JobHookConfigFormValues): void;
  jobConnectionIds: string[];
}

export function JobConfig(props: JobConfigProps): ReactElement {
  const { errors, hookType, value, onChange, jobConnectionIds } = props;

  return (
    <div className="flex flex-col gap-4">
      {hookType === 'sql' && (
        <JobConfigSqlForm
          values={value.sql}
          setValues={(newSqlData) => {
            onChange({ sql: newSqlData });
          }}
          jobConnectionIds={jobConnectionIds}
          errors={errors}
        />
      )}
    </div>
  );
}
