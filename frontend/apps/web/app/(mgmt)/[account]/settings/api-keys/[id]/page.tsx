'use client';
import { ApiKeyValueSessionStore } from '@/app/(mgmt)/[account]/new/api-key/NewApiKeyForm';
import ButtonText from '@/components/ButtonText';
import { CopyButton } from '@/components/CopyButton';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import TruncatedText from '@/components/TruncatedText';
import { PageProps } from '@/components/types';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { formatDateTime } from '@/util/util';
import { timestampDate } from '@bufbuild/protobuf/wkt';
import { useQuery } from '@connectrpc/connect-query';
import { AccountApiKey, ApiKeyService } from '@neosync/sdk';
import { InfoCircledIcon, ReloadIcon } from '@radix-ui/react-icons';
import Error from 'next/error';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement, use, useEffect, useState } from 'react';
import { useSessionStorage } from 'usehooks-ts';
import RemoveAccountApiKeyButton from './components/RemoveAccountApiKeyButton';

export default function AccountApiKeyPage(props: PageProps): ReactElement {
  const params = use(props.params);
  const id = params?.id ?? '';
  const router = useRouter();
  const { account } = useAccount();
  const { data, isLoading } = useQuery(
    ApiKeyService.method.getAccountApiKey,
    { id },
    { enabled: !!id }
  );
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
    return <Error statusCode={404} />;
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
    <div>
      <div className="flex flex-row items-center justify-between">
        <div className="flex flex-col justify-start">
          <h1>
            <TruncatedText
              text={`API Key: ${data.apiKey.name}`}
              align="start"
              maxWidth={500}
              truncatedContainerClassName="text-xl font-bold tracking-tight"
            />
          </h1>
          <h3 className="text-muted-foreground text-sm">{data.apiKey.id}</h3>
        </div>
        <div className="flex flex-row gap-2">
          <RemoveAccountApiKeyButton
            id={id}
            onDeleted={() => router.push(`/${account?.name}/settings/api-keys`)}
          />
          <Link href={`/${account?.name}/settings/api-keys/${id}/regenerate`}>
            <Button type="button">
              <ButtonText
                leftIcon={<ReloadIcon className="h-4 w-4" />}
                text="Regenerate"
              />
            </Button>
          </Link>
        </div>
      </div>
      <div className="mt-10">
        <ApiKeyDetails apiKey={data.apiKey} keyValue={apiKeyValue} />
      </div>
    </div>
  );
}

interface ApiKeyDetailsProps {
  apiKey: AccountApiKey;
  keyValue?: string;
}

function ApiKeyDetails(props: ApiKeyDetailsProps): ReactElement {
  const { apiKey, keyValue } = props;
  return (
    <div className="flex flex-col gap-3">
      {keyValue && (
        <Alert variant="success">
          <div className="flex flex-row items-center gap-3">
            <InfoCircledIcon />
            <div className="font-semibold">
              Make sure to copy this access token now as you will not be able to
              see it again!
            </div>
          </div>
        </Alert>
      )}
      <div className="flex flex-col gap-6 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
        {keyValue && (
          <div className="flex flex-row gap-3">
            <Input value={keyValue} disabled={true} />
            <CopyButton
              buttonVariant="outline"
              textToCopy={keyValue ?? ''}
              onCopiedText="Success!"
              onHoverText="Copy the API key"
            />
          </div>
        )}
        <div className="flex flex-col gap-4">
          <div className="flex flex-row gap-2">
            <p className=" text-sm tracking-tight w-[100px]">Created At:</p>
            <Badge variant="outline">
              {formatDateTime(
                apiKey.createdAt ? timestampDate(apiKey.createdAt) : new Date()
              )}
            </Badge>
          </div>
          <div className="flex flex-row gap-2">
            <p className="text-sm tracking-tight w-[100px]">Updated At:</p>
            <Badge variant="outline">
              {formatDateTime(
                apiKey.updatedAt ? timestampDate(apiKey.updatedAt) : new Date()
              )}
            </Badge>
          </div>
          <div className="flex flex-row gap-2">
            <p className=" text-sm tracking-tight w-[100px]">Expires At:</p>
            <Badge variant="outline">
              {formatDateTime(
                apiKey.expiresAt ? timestampDate(apiKey.expiresAt) : new Date()
              )}
            </Badge>
          </div>
          <div className="flex flex-row gap-2">
            <p className="text-sm tracking-tight w-[100px]">User ID:</p>
            <Badge variant="outline">{apiKey.userId}</Badge>
          </div>
        </div>
      </div>
    </div>
  );
}
