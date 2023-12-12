'use client';
import { useGetUserAccounts } from '@/libs/hooks/useUserAccounts';
import { UserAccount } from '@neosync/sdk';
import { useParams, useRouter } from 'next/navigation';
import {
  ReactElement,
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react';

interface AccountContextType {
  account: UserAccount | undefined;
  setAccount(updatedAccount: UserAccount): void;
  isLoading: boolean;
  mutateUserAccount(): void;
}
const AccountContext = createContext<AccountContextType>({
  account: undefined,
  setAccount: () => {},
  isLoading: false,
  mutateUserAccount() {},
});

interface Props {
  children: ReactNode;
}

export default function AccountProvider(props: Props): ReactElement {
  const { children } = props;
  const { account } = useParams();
  const accountName = account ?? 'personal';

  const { data: accountsResponse, isLoading, mutate } = useGetUserAccounts();
  const router = useRouter();

  const [userAccount, setUserAccount] = useState<UserAccount | undefined>(
    undefined
  );

  useEffect(() => {
    if (isLoading) {
      return;
    }
    if (userAccount?.name === accountName) {
      return;
    }
    const foundAccount = accountsResponse?.accounts.find(
      (a) => a.name === accountName
    );
    if (userAccount?.id === foundAccount?.id) {
      return;
    }
    if (foundAccount) {
      setUserAccount(foundAccount);
    }
  }, [userAccount, accountsResponse?.accounts.length, isLoading, accountName]);

  function setAccount(userAccount: UserAccount): void {
    if (userAccount.name !== accountName) {
      router.push(`/${userAccount.name}`);
      setUserAccount(userAccount);
    }
  }

  return (
    <AccountContext.Provider
      value={{
        account: userAccount,
        setAccount: setAccount,
        isLoading,
        mutateUserAccount: mutate,
      }}
    >
      {children}
    </AccountContext.Provider>
  );
}

export function useAccount(): AccountContextType {
  const account = useContext(AccountContext);
  return account;
}
