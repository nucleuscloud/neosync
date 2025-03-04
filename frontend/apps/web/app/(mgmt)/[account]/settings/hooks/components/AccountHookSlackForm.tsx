import ButtonText from '@/components/ButtonText';
import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { AccountHookService } from '@neosync/sdk';
import { CheckCircledIcon, ReloadIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { GoXCircleFill } from 'react-icons/go';
import { PiPlugs } from 'react-icons/pi';
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
          title="Slack Channel ID"
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
  const {
    data: testSlackConnection,
    refetch: refetchTestSlackConnection,
    isFetching: isFetchingTestSlackConnection,
  } = useQuery(
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
    openSlackConnectionWindow(urlResp.url);
  }

  async function onRefreshClick(): Promise<void> {
    if (isFetchingTestSlackConnection) {
      return;
    }
    await refetchTestSlackConnection();
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-row gap-4 items-center">
        {testSlackConnection?.hasConfiguration ? (
          <div className="flex flex-row gap-2 items-center">
            <CheckCircledIcon className="w-4 h-4" />
            <p>
              Slack is connected to {testSlackConnection.testResponse?.team}
            </p>
          </div>
        ) : (
          <div className="flex flex-row gap-1 items-center">
            <GoXCircleFill className="w-4 h-4" />
            <p>Slack is not connected</p>
          </div>
        )}
        <div className="flex">
          <RecheckSlackConnectionButton onRefreshClick={onRefreshClick} />
        </div>
        <div className="flex">
          <ConnectSlackConnectionButton
            onConnect={onConnectClick}
            hasConfiguration={testSlackConnection?.hasConfiguration ?? false}
          />
        </div>
      </div>
      <div className="flex flex-row gap-3">
        {testSlackConnection?.hasConfiguration &&
          testSlackConnection?.error && (
            <Alert variant="destructive">
              <AlertTitle>{testSlackConnection.error}</AlertTitle>
            </Alert>
          )}
      </div>
    </div>
  );
}

interface RecheckSlackConnectionButtonProps {
  onRefreshClick(): Promise<void> | void;
}
function RecheckSlackConnectionButton(
  props: RecheckSlackConnectionButtonProps
): ReactElement {
  const { onRefreshClick } = props;

  const [isRefreshing, setIsRefreshing] = useState(false);

  async function onClick(): Promise<void> {
    if (isRefreshing) {
      return;
    }
    setIsRefreshing(true);
    try {
      await onRefreshClick();
    } finally {
      setIsRefreshing(false);
    }
  }

  return (
    <TooltipProvider>
      <Tooltip delayDuration={200}>
        <TooltipTrigger asChild>
          <Button type="button" variant="outline" onClick={onClick}>
            <ButtonText
              leftIcon={
                <ReloadIcon
                  className={cn('h-4 w-4', isRefreshing && 'animate-spin')}
                />
              }
              text="Verify"
            />
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>Verify that Neosync has a connection to your Slack</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

interface ConnectSlackConnectionButtonProps {
  onConnect(): Promise<void> | void;
  hasConfiguration: boolean;
}
function ConnectSlackConnectionButton(
  props: ConnectSlackConnectionButtonProps
): ReactElement {
  const { onConnect, hasConfiguration } = props;

  const [isConnecting, setIsConnecting] = useState(false);

  async function onClick(): Promise<void> {
    if (isConnecting) {
      return;
    }
    setIsConnecting(true);
    try {
      await onConnect();
    } finally {
      setIsConnecting(false);
    }
  }

  return (
    <TooltipProvider>
      <Tooltip delayDuration={200}>
        <TooltipTrigger asChild>
          <Button type="button" variant="outline" onClick={onClick}>
            <ButtonText
              leftIcon={<PiPlugs className="h-4 w-4" />}
              text={hasConfiguration ? 'Reconnect' : 'Connect'}
            />
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>
            {hasConfiguration
              ? 'Reconnect Neosync to your Slack'
              : 'Connect Neosync to your Slack'}
          </p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

function openSlackConnectionWindow(slackConnectionUrl: string): void {
  if (!window || !screen) {
    return;
  }
  const w = 1080;
  const h = 800;

  const left = screen.width / 2 - w / 2;
  const top = screen.height / 2 - h / 2;

  window.open(
    slackConnectionUrl,
    '_blank',
    `popup,width=${w},height=${h},left=${left},top=${top}`
  );
}
