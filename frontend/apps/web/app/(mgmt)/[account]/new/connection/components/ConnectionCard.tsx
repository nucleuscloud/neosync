'use client';
import ConnectionIcon from '@/components/connections/ConnectionIcon';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { cn } from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { UserAccountService } from '@neosync/sdk';
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement } from 'react';
import { ConnectionMeta } from '../../../connections/util';

interface Props {
  connection: ConnectionMeta;
}

export default function ConnectionCard(props: Props): ReactElement<any> {
  const { connection } = props;
  const router = useRouter();
  const { account } = useAccount();
  const searchParams = useSearchParams();
  const { data: systemInfo } = useQuery(
    UserAccountService.method.getSystemInformation
  );
  const hasValidLicense = systemInfo?.license?.isValid ?? false;
  const isClickable =
    !connection.isLicenseOnly || (connection.isLicenseOnly && hasValidLicense);
  return (
    <Card
      onClick={() => {
        if (!isClickable) {
          return;
        }
        router.push(
          `/${account?.name}/new/connection/${
            connection.urlSlug
          }?${searchParams.toString()}`
        );
      }}
      className={cn(
        'cursor-pointer hover:border hover:border-gray-500 dark:border-gray-700 dark:hover:border-gray-600',
        !isClickable && 'opacity-50',
        !isClickable && 'cursor-not-allowed'
      )}
    >
      <CardHeader>
        <CardTitle>
          <div className="flex flex-row items-center space-x-2">
            <ConnectionIcon
              connectionType={connection.connectionType}
              connectionTypeVariant={connection.connectionTypeVariant}
            />
            <p>{connection.name}</p>
            {connection.isExperimental ? <Badge>Experimental</Badge> : null}
            {connection.isLicenseOnly ? <Badge>Enterprise</Badge> : null}
          </div>
        </CardTitle>
        <CardDescription>{connection.description}</CardDescription>
      </CardHeader>
    </Card>
  );
}
