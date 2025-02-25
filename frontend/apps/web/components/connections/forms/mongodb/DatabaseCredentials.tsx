import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { SecurePasswordInput } from '@/components/SecurePasswordInput';
import { Input } from '@/components/ui/input';
import { MongoDbFormValues } from '@/yup-validations/connections';
import { ReactElement } from 'react';
import { SecretRevealProps } from '../SharedFormInputs';

interface Props extends SecretRevealProps<MongoDbFormValues> {
  urlValue: MongoDbFormValues['url'];
  onUrlValueChange(value: MongoDbFormValues['url']): void;

  errors: Record<string, string>;
}

export default function DatabaseCredentials(props: Props): ReactElement<any> {
  const {
    urlValue,
    onUrlValueChange,
    errors,
    isViewMode,
    canViewSecrets,
    onRevealClick,
  } = props;

  return (
    <div className="flex flex-col gap-4">
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
    </div>
  );
}

interface UrlTabProps extends SecretRevealProps<MongoDbFormValues['url']> {
  urlValue: MongoDbFormValues['url'];
  onUrlValueChange(value: MongoDbFormValues['url']): void;
  error?: string;
}

function UrlTab(props: UrlTabProps): ReactElement<any> {
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
          placeholder="mongodb://username:password@hostname:port/database"
        />
      )}
      <FormErrorMessage message={error} />
    </div>
  );
}
