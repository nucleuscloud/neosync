import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { ReactElement } from 'react';
import AccountHookWebhookForm from './AccountHookWebhookForm';
import { AccountHookConfigFormValues, HookTypeFormValue } from './validation';

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
        isRequired={true}
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

interface EnabledProps {
  error?: string;
  value: boolean;
  onChange(value: boolean): void;
}

export function Enabled(props: EnabledProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="flex flex-col gap-4">
      <FormHeader
        htmlFor="enabled"
        title="Enabled"
        description="Whether or not this hook will be invoked during a job run"
        isErrored={!!error}
      />
      <ToggleGroup
        className="flex justify-start"
        type="single"
        onValueChange={(value) => {
          if (!value) {
            return;
          }
          onChange(value === 'yes');
        }}
        value={value ? 'yes' : 'no'}
      >
        <ToggleGroupItem value="yes">Yes</ToggleGroupItem>
        <ToggleGroupItem value="no">No</ToggleGroupItem>
      </ToggleGroup>
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
        description="The type of hook. Currently only webhooks are supported"
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
        <ToggleGroupItem value="webhook">Webhook</ToggleGroupItem>
      </ToggleGroup>
      <FormErrorMessage message={error} />
    </div>
  );
}

interface AccountHookConfigProps {
  errors: Record<string, string>;
  hookType: HookTypeFormValue;
  value: AccountHookConfigFormValues;
  onChange(value: AccountHookConfigFormValues): void;
}

export function AccountHookConfig(props: AccountHookConfigProps): ReactElement {
  const { errors, hookType, value, onChange } = props;

  return (
    <div className="flex flex-col gap-4">
      {hookType === 'webhook' && (
        <AccountHookWebhookForm
          values={value.webhook}
          setValues={(newWebhookData) => {
            onChange({ webhook: newWebhookData });
          }}
          errors={errors}
        />
      )}
    </div>
  );
}
