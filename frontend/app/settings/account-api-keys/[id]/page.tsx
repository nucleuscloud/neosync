'use client';
import { ApiKeyValueSessionStore } from '@/app/new/account-api-key/NewApiKeyForm';
import ButtonText from '@/components/ButtonText';
import { CopyButton } from '@/components/CopyButton';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { useGetAccountApiKey } from '@/libs/hooks/useGetAccountApiKey';
import { formatDateTime } from '@/util/util';
import { ReloadIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement, useEffect, useState } from 'react';
import { useSessionStorage } from 'usehooks-ts';
import RemoveAccountApiKeyButton from './components/RemoveAccountApiKeyButton';

export default function AccountApiKeyPage({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading } = useGetAccountApiKey(id);
  const [sessionApiKeyValue] = useSessionStorage<
    ApiKeyValueSessionStore | undefined
  >(id, undefined);
  const [apiKeyValue, setApiKeyValue] = useState<string | undefined>(
    sessionApiKeyValue?.keyValue
  );
  // Don't persist the api key in session storage any longer than is necessary.
  useEffect(() => {
    if (!!sessionApiKeyValue) {
      window.sessionStorage.removeItem(id);
      setApiKeyValue(sessionApiKeyValue.keyValue);
    }
  }, [sessionApiKeyValue]);

  if (!id) {
    return <div>Not Found</div>;
  }
  if (isLoading) {
    return (
      <div>
        <SkeletonForm />
      </div>
    );
  }
  if (!data?.apiKey) {
    return (
      <div className="mt-10">
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve api key`}</AlertTitle>
        </Alert>
      </div>
    );
  }
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header={data.apiKey.name}
          description={data.apiKey.id}
          extraHeading={
            <div className="flex flex-row gap-2">
              <RemoveAccountApiKeyButton id={id} />
              <Link href={`/settings/account-api-keys/${id}/regenerate`}>
                <Button type="button">
                  <ButtonText
                    leftIcon={<ReloadIcon className="h-4 w-4" />}
                    text="Regenerate"
                  />
                </Button>
              </Link>
            </div>
          }
        />
      }
      containerClassName="mx-24"
    >
      <div className="flex flex-col gap-3">
        {apiKeyValue && (
          <div>
            <KeyValueAlert keyValue={apiKeyValue} />
          </div>
        )}

        <div className="flex flex-row gap-2">
          <p className="text-lg tracking-tight">Expires At:</p>
          <p className="text-lg tracking-tight">
            {formatDateTime(data.apiKey.expiresAt?.toDate())}
          </p>
        </div>
        <div className="flex flex-row gap-2">
          <p className="text-lg tracking-tight">Created At:</p>
          <p className="text-lg tracking-tight">
            {formatDateTime(data.apiKey.createdAt?.toDate())}
          </p>
        </div>
        <div className="flex flex-row gap-2">
          <p className="text-lg tracking-tight">Updated At:</p>
          <p className="text-lg tracking-tight">
            {formatDateTime(data.apiKey.updatedAt?.toDate())}
          </p>
        </div>
        <div className="flex flex-row gap-2">
          <p className="text-lg tracking-tight">User ID:</p>
          <p className="text-lg tracking-tight">{data.apiKey.userId}</p>
        </div>
      </div>
    </OverviewContainer>
  );
}

interface KeyValueAlertProps {
  keyValue?: string;
}
function KeyValueAlert(props: KeyValueAlertProps): ReactElement | null {
  const { keyValue } = props;

  if (!keyValue) {
    return null;
  }

  return (
    <Alert variant="default" className="flex flex-col gap-3">
      <AlertTitle>
        Make sure to copy this access token now as you will not be able to see
        it again!
      </AlertTitle>
      <AlertDescription>
        <div className="flex flex-row">
          <div className="flex w-full bg-transparent rounded-md border border-input py-1 px-3">
            <p>{keyValue}</p>
          </div>
          <CopyButton
            buttonVariant="outline"
            textToCopy={keyValue}
            onCopiedText="Success!"
            onHoverText="Copy the api key"
          />
        </div>
      </AlertDescription>
    </Alert>
  );
}
