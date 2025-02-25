import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { AccountHookService } from '@neosync/sdk';
import { ReactElement } from 'react';
import { AccountHookSlackFormValues } from './validation';

interface Props {
  values: AccountHookSlackFormValues;
  setValues(values: AccountHookSlackFormValues): void;
  errors: Record<string, string>;
}

export default function AccountHookSlackForm(props: Props): ReactElement {
  const { values, setValues, errors } = props;

  return (
    <>
      <ConnectToSlackButton />
      <div className="flex flex-col gap-3">
        <FormHeader
          title="Slack Channel"
          description="The Slack channel to send the event to"
          isErrored={!!errors['config.slack.channelId']}
          isRequired={true}
        />
        <Input
          id="channelId"
          value={values.channelId}
          onChange={(e) => setValues({ ...values, channelId: e.target.value })}
        />
        <FormErrorMessage message={errors['config.slack.channelId']} />
      </div>
    </>
  );
}

interface ConnectToSlackButtonProps {}
function ConnectToSlackButton(props: ConnectToSlackButtonProps): ReactElement {
  const {} = props;
  const { account } = useAccount();
  const { data: testSlackConnection, isLoading: isTestSlackConnectionLoading } =
    useQuery(
      AccountHookService.method.testSlackConnection,
      {
        accountId: account?.id,
      },
      {
        enabled: !!account?.id,
      }
    );
  const { mutateAsync: getSlackConnectionUrl } = useMutation(
    AccountHookService.method.getSlackConnectionUrl
  );

  async function onConnectClick(): Promise<void> {
    const urlResp = await getSlackConnectionUrl({
      accountId: account?.id,
    });
    console.log(urlResp.url);
  }

  return (
    <>
      <Button onClick={onConnectClick}>Connect to Slack</Button>
      {testSlackConnection?.hasConfiguration ? 'YES!!!' : 'NO!!!'}
    </>
  );
}
