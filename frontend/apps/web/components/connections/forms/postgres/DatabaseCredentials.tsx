import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import OSSOnlyGuard from '@/components/guards/OSSOnlyGuard';
import { PasswordInput } from '@/components/PasswordComponent';
import { SecurePasswordInput } from '@/components/SecurePasswordInput';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  ActiveConnectionTab,
  PostgresFormValues,
  SSL_MODES,
} from '@/yup-validations/connections';
import { ReactElement } from 'react';
import { SecretRevealProps } from '../SharedFormInputs';

interface Props extends SecretRevealProps<PostgresFormValues> {
  dbValue: PostgresFormValues['db'];
  onDbValueChange(value: PostgresFormValues['db']): void;

  urlValue: PostgresFormValues['url'];
  onUrlValueChange(value: PostgresFormValues['url']): void;

  envVarValue: PostgresFormValues['envVar'];
  onEnvVarValueChange(value: PostgresFormValues['envVar']): void;

  errors: Record<string, string>;

  activeTab: ActiveConnectionTab;
  setActiveTab(tab: ActiveConnectionTab): void;
}

export default function DatabaseCredentials(props: Props): ReactElement {
  const {
    dbValue,
    onDbValueChange,
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
          isViewMode={isViewMode}
          canViewSecrets={canViewSecrets}
          onRevealClick={async () => {
            const values = await onRevealClick();
            return values?.url ?? '';
          }}
        />
      )}
      {activeTab === 'host' && (
        <HostTab
          dbValue={dbValue}
          onDbValueChange={onDbValueChange}
          errors={errors}
          isViewMode={isViewMode}
          canViewSecrets={canViewSecrets}
          onRevealClick={async () => {
            const values = await onRevealClick();
            return values?.db ?? {};
          }}
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
  activeTab: ActiveConnectionTab;
  setActiveTab(tab: ActiveConnectionTab): void;
}

function ActiveTabSelector(props: ActiveTabProps): ReactElement {
  const { activeTab, setActiveTab } = props;

  return (
    <RadioGroup
      defaultValue={activeTab}
      onValueChange={(e) => setActiveTab(e as ActiveConnectionTab)}
      value={activeTab}
    >
      <div className="flex flex-col md:flex-row gap-4">
        <div className="text-sm">Connect by:</div>
        <div className="flex items-center space-x-2">
          <RadioGroupItem value="url" id="r2" />
          <Label htmlFor="r2">URL</Label>
        </div>
        <div className="flex items-center space-x-2">
          <RadioGroupItem value="host" id="r1" />
          <Label htmlFor="r1">Host</Label>
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

interface UrlTabProps extends SecretRevealProps<PostgresFormValues['url']> {
  urlValue: PostgresFormValues['url'];
  onUrlValueChange(value: PostgresFormValues['url']): void;
  error?: string;
}

function UrlTab(props: UrlTabProps): ReactElement {
  const {
    urlValue,
    onUrlValueChange,
    error,
    isViewMode,
    canViewSecrets,
    onRevealClick,
  } = props;

  return (
    <div className="space-y-2">
      <FormHeader
        htmlFor="url"
        title="Connection URL"
        description="The URL of the database"
        isErrored={!!error}
        isRequired={true}
      />
      {isViewMode ? (
        <SecurePasswordInput
          value={urlValue || ''}
          maskedValue={urlValue ?? ''}
          disabled={!canViewSecrets}
          onRevealPassword={
            canViewSecrets
              ? async () => {
                  const values = await onRevealClick();
                  return values ?? '';
                }
              : undefined
          }
        />
      ) : (
        <Input
          id="url"
          autoCapitalize="off"
          data-1p-ignore // tells 1password extension to not autofill this field
          value={urlValue || ''}
          onChange={(e) => onUrlValueChange(e.target.value)}
          placeholder="postgres://username:password@hostname:port/database"
        />
      )}
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
        isRequired={true}
      />
      <Input
        id="envVar"
        autoCapitalize="off"
        data-1p-ignore // tells 1password extension to not autofill this field
        value={envVarValue || ''}
        onChange={(e) => onEnvVarValueChange(e.target.value)}
        placeholder="USER_DEFINED_POSTGRES_URL"
      />
      <FormErrorMessage message={error} />
    </div>
  );
}

interface HostTabProps extends SecretRevealProps<PostgresFormValues['db']> {
  dbValue: PostgresFormValues['db'];
  onDbValueChange(value: PostgresFormValues['db']): void;
  errors: Record<string, string>;
}

function HostTab(props: HostTabProps): ReactElement {
  const {
    dbValue,
    onDbValueChange,
    errors,
    isViewMode,
    canViewSecrets,
    onRevealClick,
  } = props;

  return (
    <>
      <div className="space-y-2">
        <FormHeader
          htmlFor="host"
          title="Host"
          description="The host name"
          isErrored={!!errors.host}
          isRequired={true}
        />
        <Input
          id="host"
          autoCapitalize="off"
          data-1p-ignore // tells 1password extension to not autofill this field
          value={dbValue.host || ''}
          onChange={(e) =>
            onDbValueChange({ ...dbValue, host: e.target.value })
          }
          placeholder="localhost"
        />
        <FormErrorMessage message={errors.host} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="port"
          title="Port"
          description="The port number"
          isErrored={!!errors.port}
          isRequired={true}
        />
        <Input
          id="port"
          type="number"
          value={dbValue.port || ''}
          onChange={(e) =>
            onDbValueChange({ ...dbValue, port: e.target.valueAsNumber })
          }
          placeholder="5432"
        />
        <FormErrorMessage message={errors.port} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="name"
          title="Database Name"
          description="The name of the database"
          isErrored={!!errors.name}
          isRequired={true}
        />
        <Input
          id="name"
          value={dbValue.name || ''}
          onChange={(e) =>
            onDbValueChange({ ...dbValue, name: e.target.value })
          }
          placeholder="postgres"
        />
        <FormErrorMessage message={errors.name} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="user"
          title="Username"
          description="The username for the database"
          isErrored={!!errors.user}
          isRequired={true}
        />
        <Input
          id="user"
          value={dbValue.user || ''}
          onChange={(e) =>
            onDbValueChange({ ...dbValue, user: e.target.value })
          }
          placeholder="postgres"
        />
        <FormErrorMessage message={errors.user} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="password"
          title="Password"
          description="The password for the database"
          isErrored={!!errors.pass}
          isRequired={true}
        />
        {isViewMode ? (
          <SecurePasswordInput
            value={dbValue.pass || ''}
            disabled={!canViewSecrets}
            onRevealPassword={
              canViewSecrets
                ? async () => {
                    const values = await onRevealClick();
                    return values?.pass ?? '';
                  }
                : undefined
            }
          />
        ) : (
          <PasswordInput
            id="password"
            value={dbValue.pass || ''}
            onChange={(e) =>
              onDbValueChange({ ...dbValue, pass: e.target.value })
            }
            placeholder="postgres"
          />
        )}

        <FormErrorMessage message={errors.pass} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="sslMode"
          title="SSL Mode"
          description="Turn on SSL Mode to use TLS for client/server encryption."
          isErrored={!!errors.sslMode}
          isRequired={true}
        />
        <Select
          onValueChange={(value) =>
            onDbValueChange({ ...dbValue, sslMode: value })
          }
          value={dbValue.sslMode || ''}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {SSL_MODES.map((mode) => (
              <SelectItem className="cursor-pointer" key={mode} value={mode}>
                {mode}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <FormErrorMessage message={errors.sslMode} />
      </div>
    </>
  );
}
