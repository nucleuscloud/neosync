import { useAccount } from '@/components/providers/account-provider';
import { Connection } from '@neosync/sdk';
import { useRouter, useSearchParams } from 'next/navigation';

export function useOnCreateSuccess(): (conn: Connection) => Promise<void> {
  const router = useRouter();
  const { account } = useAccount();
  const searchParams = useSearchParams();
  const returnTo = searchParams.get('returnTo');

  return async (conn: Connection): Promise<void> => {
    if (!account) {
      return;
    }
    if (returnTo) {
      router.push(returnTo);
    } else if (conn.id) {
      router.push(`/${account.name}/connections/${conn.id}`);
    }
  };
}
