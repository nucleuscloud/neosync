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
import { useLocalStorage } from 'usehooks-ts';

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

const DEFAULT_ACCOUNT_NAME = 'personal';
const STORAGE_ACCOUNT_KEY = 'account';

export default function AccountProvider(props: Props): ReactElement {
  const { children } = props;
  const { account } = useParams();
  const accountName = useGetAccountName();
  const [, setLastSelectedAccount] = useLocalStorage(
    STORAGE_ACCOUNT_KEY,
    accountName ?? DEFAULT_ACCOUNT_NAME
  );

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
      setLastSelectedAccount(foundAccount.name);
      const accountParam = getSingleOrUndefined(account);
      if (!accountParam || accountParam !== foundAccount.name) {
        router.push(`/${foundAccount.name}/jobs`);
      }
    }
  }, [
    userAccount?.id,
    userAccount?.name,
    accountsResponse?.accounts.length,
    isLoading,
    accountName,
  ]);

  function setAccount(userAccount: UserAccount): void {
    if (userAccount.name !== accountName) {
      router.push(`/${userAccount.name}`);
      setUserAccount(userAccount);
      setLastSelectedAccount(userAccount.name);
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

function useGetAccountName(): string {
  const { account } = useParams();
  const [storedAccount] = useLocalStorage(
    STORAGE_ACCOUNT_KEY,
    account ?? DEFAULT_ACCOUNT_NAME
  );

  const accountParam = getSingleOrUndefined(account);
  if (accountParam) {
    return accountParam;
  }
  const singleStoredAccount = getSingleOrUndefined(storedAccount);
  if (singleStoredAccount) {
    return singleStoredAccount;
  }
  return DEFAULT_ACCOUNT_NAME;
}

function getSingleOrUndefined(val: string | string[]): string | undefined {
  if (Array.isArray(val)) {
    return val.length > 0 ? val[0] : undefined;
  }
  return val;
}

export function useAccount(): AccountContextType {
  const account = useContext(AccountContext);
  return account;
}
