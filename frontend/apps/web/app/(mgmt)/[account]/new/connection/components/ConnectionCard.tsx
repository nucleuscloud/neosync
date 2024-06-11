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
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement } from 'react';
import { ConnectionType } from '../../../connections/util';

export interface ConnectionMeta {
  name: string;
  description: string;
  urlSlug: string;
  connectionType: ConnectionType;
  isExperimental?: boolean;
}

interface Props {
  connection: ConnectionMeta;
}

export default function ConnectionCard(props: Props): ReactElement {
  const { connection } = props;
  const router = useRouter();
  const { account } = useAccount();
  const searchParams = useSearchParams();
  return (
    <Card
      onClick={() =>
        router.push(
          `/${account?.name}/new/connection/${
            connection.urlSlug
          }?${searchParams.toString()}`
        )
      }
      className="cursor-pointer hover:border hover:border-gray-500 dark:border-gray-700 dark:hover:border-gray-600"
    >
      <CardHeader>
        <CardTitle>
          <div className="flex flex-row items-center space-x-2">
            <ConnectionIcon name={connection.name} />
            <p>{connection.name}</p>
            {connection.isExperimental ? <Badge>Experimental</Badge> : null}
          </div>
        </CardTitle>
        <CardDescription>{connection.description}</CardDescription>
      </CardHeader>
    </Card>
  );
}
