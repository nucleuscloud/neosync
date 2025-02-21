import { ReactElement } from 'react';

import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { PasswordInput } from '@/components/PasswordComponent';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { AccountHookWebhookFormValues } from './validation';

interface Props {
  values: AccountHookWebhookFormValues;
  setValues(values: AccountHookWebhookFormValues): void;
  errors: Record<string, string>;
}

export default function AccountHookWebhookForm(props: Props): ReactElement {
  const { values, setValues, errors } = props;
  return (
    <>
      <div className="flex flex-col gap-3">
        <FormHeader
          title="Webhook URL"
          description="The URL that will be invoked"
          isErrored={!!errors['config.webhook.url']}
          isRequired={true}
        />
        <Input
          id="url"
          value={values.url}
          onChange={(e) => setValues({ ...values, url: e.target.value })}
        />
        <FormErrorMessage message={errors['config.webhook.url']} />
      </div>
      <div className="flex flex-col gap-3">
        <FormHeader
          title="Webhook Secret"
          description="The secret that will be used to authenticate the webhook"
          isErrored={!!errors['config.webhook.secret']}
          isRequired={true}
        />
        <PasswordInput
          id="secret"
          value={values.secret}
          onChange={(e) => setValues({ ...values, secret: e.target.value })}
        />
        <FormErrorMessage message={errors['config.webhook.secret']} />
      </div>
      <div className="flex flex-col gap-3">
        <FormHeader
          title="Disable SSL Verification"
          description="Whether to disable SSL verification for the webhook"
        />
        <Switch
          id="disableSslVerification"
          checked={values.disableSslVerification}
          onCheckedChange={(checked) =>
            setValues({ ...values, disableSslVerification: checked })
          }
        />
      </div>
    </>
  );
}
