import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { PasswordInput } from '@/components/PasswordComponent';
import { SecurePasswordInput } from '@/components/SecurePasswordInput';
import { Input } from '@/components/ui/input';
import { OpenAiFormValues } from '@/yup-validations/connections';
import { ReactElement } from 'react';

interface Props {
  errors: Record<string, string>;
  value: OpenAiFormValues['sdk'];
  onChange(value: OpenAiFormValues['sdk']): void;

  isViewMode?: boolean;
  canViewSecrets?: boolean;
  onRevealPassword?(): Promise<string>;
}

export default function Sdk(props: Props): ReactElement {
  const {
    errors,
    value,
    onChange,
    isViewMode,
    canViewSecrets,
    onRevealPassword,
  } = props;

  return (
    <div className="flex flex-col gap-4">
      <SdkUrl
        error={errors['url']}
        value={value.url}
        onChange={(apiUrl) => onChange({ ...value, url: apiUrl })}
      />
      <SdkApiKey
        error={errors['apiKey']}
        value={value.apiKey}
        onChange={(apiKey) => onChange({ ...value, apiKey })}
        isViewMode={isViewMode}
        canViewSecrets={canViewSecrets}
        onRevealPassword={onRevealPassword}
      />
    </div>
  );
}

interface SdkUrlProps {
  error?: string;
  value: string;
  onChange(value: string): void;
}

function SdkUrl(props: SdkUrlProps): ReactElement {
  const { error, value, onChange } = props;

  return (
    <div className="space-y-2">
      <FormHeader
        title="API Url"
        description="The url of the OpenAI API (or equivalent) server"
        isErrored={!!error}
      />
      <Input
        autoCapitalize="off" // we don't allow capitals
        data-1p-ignore // tells 1password extension to not autofill this field
        value={value || ''}
        onChange={(e) => onChange(e.target.value)}
        placeholder="https://api.openai.com/v1"
      />
      <FormErrorMessage message={error} />
    </div>
  );
}

interface SdkApiKeyProps {
  error?: string;
  value: string;
  onChange(value: string): void;
  isViewMode?: boolean;
  canViewSecrets?: boolean;
  onRevealPassword?(): Promise<string>;
}

function SdkApiKey(props: SdkApiKeyProps): ReactElement {
  const {
    error,
    value,
    onChange,
    isViewMode,
    canViewSecrets,
    onRevealPassword,
  } = props;

  return (
    <div className="space-y-2">
      <FormHeader
        title="API Key"
        description="The API key for the OpenAI API (or equivalent) server"
        isErrored={!!error}
      />
      {isViewMode ? (
        <SecurePasswordInput
          maskedValue="••••••••"
          value={value || ''}
          disabled={!canViewSecrets}
          onRevealPassword={canViewSecrets ? onRevealPassword : undefined}
          placeholder="sk-..."
        />
      ) : (
        <PasswordInput
          autoCapitalize="off"
          data-1p-ignore
          value={value || ''}
          onChange={(e) => onChange(e.target.value)}
          placeholder="sk-..."
        />
      )}
      <FormErrorMessage message={error} />
    </div>
  );
}
