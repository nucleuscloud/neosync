import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import OSSOnlyGuard from '@/components/guards/OSSOnlyGuard';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import {
  MssqlActiveConnectionTab,
  MssqlFormValues,
  PostgresFormValues,
} from '@/yup-validations/connections';
import { ReactElement } from 'react';
import { SecretRevealProps } from '../SharedFormInputs';

interface Props extends SecretRevealProps<MssqlFormValues> {
  urlValue: MssqlFormValues['url'];
  onUrlValueChange(value: MssqlFormValues['url']): void;

  envVarValue: MssqlFormValues['envVar'];
  onEnvVarValueChange(value: MssqlFormValues['envVar']): void;

  errors: Record<string, string>;

  activeTab: MssqlActiveConnectionTab;
  setActiveTab(tab: MssqlActiveConnectionTab): void;
}

export default function DatabaseCredentials(props: Props): ReactElement {
  const {
    urlValue,
    onUrlValueChange,
    envVarValue,
    onEnvVarValueChange,
    errors,
    isViewMode,
    canViewSecrets,
    onRevealClick,
    activeTab,
    setActiveTab,
  } = props;

  return (
    <div className="flex flex-col gap-4">
      <ActiveTabSelector activeTab={activeTab} setActiveTab={setActiveTab} />
      {activeTab === 'url' && (
        <UrlTab
          urlValue={urlValue}
          onUrlValueChange={onUrlValueChange}
          error={errors.url}
        />
      )}
      {activeTab === 'url-env' && (
        <UrlEnvTab
          envVarValue={envVarValue}
          onEnvVarValueChange={onEnvVarValueChange}
          error={errors.envVar}
        />
      )}
    </div>
  );
}

interface ActiveTabProps {
  activeTab: MssqlActiveConnectionTab;
  setActiveTab(tab: MssqlActiveConnectionTab): void;
}

function ActiveTabSelector(props: ActiveTabProps): ReactElement {
  const { activeTab, setActiveTab } = props;

  return (
    <RadioGroup
      defaultValue={activeTab}
      onValueChange={(e) => setActiveTab(e as MssqlActiveConnectionTab)}
      value={activeTab}
    >
      <div className="flex flex-col md:flex-row gap-4">
        <div className="text-sm">Connect by:</div>
        <div className="flex items-center space-x-2">
          <RadioGroupItem value="url" id="r2" />
          <Label htmlFor="r2">URL</Label>
        </div>
        <OSSOnlyGuard>
          <div className="flex items-center space-x-2">
            <RadioGroupItem value="url-env" id="r3" />
            <Label htmlFor="r3">Environment Variable</Label>
          </div>
        </OSSOnlyGuard>
      </div>
    </RadioGroup>
  );
}

interface UrlTabProps {
  urlValue: MssqlFormValues['url'];
  onUrlValueChange(value: MssqlFormValues['url']): void;
  error?: string;
}

function UrlTab(props: UrlTabProps): ReactElement {
  const { urlValue, onUrlValueChange, error } = props;

  return (
    <div className="space-y-2">
      <FormHeader
        htmlFor="url"
        title="Connection URL"
        description="The URL of the database"
        isErrored={!!error}
      />
      <Input
        id="url"
        autoCapitalize="off"
        data-1p-ignore // tells 1password extension to not autofill this field
        value={urlValue || ''}
        onChange={(e) => onUrlValueChange(e.target.value)}
        placeholder="sqlserver://username:password@host:port/instance?param1=value&param2=value"
      />
      <FormErrorMessage message={error} />
    </div>
  );
}

interface UrlEnvTabProps {
  envVarValue: PostgresFormValues['envVar'];
  onEnvVarValueChange(value: PostgresFormValues['envVar']): void;
  error?: string;
}

function UrlEnvTab(props: UrlEnvTabProps): ReactElement {
  const { envVarValue, onEnvVarValueChange, error } = props;

  return (
    <div className="space-y-2">
      <FormHeader
        htmlFor="envVar"
        title="Environment Variable"
        description={`The environment variable that contains the connection URL.
Must start with "USER_DEFINED_". Must be present on
both the backend and the worker processes for full
functionality.`}
        isErrored={!!error}
      />
      <Input
        id="envVar"
        autoCapitalize="off"
        data-1p-ignore // tells 1password extension to not autofill this field
        value={envVarValue || ''}
        onChange={(e) => onEnvVarValueChange(e.target.value)}
        placeholder="USER_DEFINED_MSSQL_URL"
      />
      <FormErrorMessage message={error} />
    </div>
  );
}
