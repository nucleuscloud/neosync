import { ReactElement } from 'react';

import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { PasswordInput } from '@/components/PasswordComponent';
import { Input } from '@/components/ui/input';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
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
          description="The URL to send the event to"
          isErrored={!!errors['config.webhook.url']}
          isRequired={true}
        />
        <Input
          id="url"
          value={values.url}
          onChange={(e) => setValues({ ...values, url: e.target.value })}
          placeholder="https://example.com/webhook"
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
          placeholder="your-secret-key"
        />
        <FormErrorMessage message={errors['config.webhook.secret']} />
      </div>
      <div className="flex flex-col gap-3">
        <FormHeader
          title="Disable SSL Verification"
          description="Whether to disable SSL verification for the webhook"
        />
        <ToggleGroup
          className="flex justify-start"
          type="single"
          value={values.disableSslVerification ? 'yes' : 'no'}
          onValueChange={(value) => {
            if (!value) {
              return;
            }
            setValues({
              ...values,
              disableSslVerification: value === 'yes',
            });
          }}
        >
          <ToggleGroupItem value="no">No</ToggleGroupItem>
          <ToggleGroupItem value="yes">Yes</ToggleGroupItem>
        </ToggleGroup>
      </div>
    </>
  );
}
