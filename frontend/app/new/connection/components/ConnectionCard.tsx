'use client';
import ConnectionIcon from '@/components/connections/ConnectionIcon';
import { Avatar } from '@/components/ui/avatar';
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { useRouter, useSearchParams } from 'next/navigation';

export interface ConnectionMeta {
  name: string;
  description: string;
  urlSlug: string;
}

interface Props {
  connection: ConnectionMeta;
}

export default function ConnectionCard(props: Props) {
  const { connection } = props;
  const router = useRouter();
  const searchParams = useSearchParams();
  return (
    <Card
      onClick={() =>
        router.push(
          `/new/connection/${connection.urlSlug}?${searchParams.toString()}`
        )
      }
      className="cursor-pointer"
    >
      <CardHeader>
        <CardTitle>
          <div className="flex flex-row items-center space-x-2">
            <Avatar>
              <ConnectionIcon name={connection.name} />
            </Avatar>
            <p>{connection.name}</p>
          </div>
        </CardTitle>
        <CardDescription>{connection.description}</CardDescription>
      </CardHeader>
    </Card>
  );
}
