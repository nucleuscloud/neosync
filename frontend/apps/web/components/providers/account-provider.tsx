'use client';
import { useQuery } from '@connectrpc/connect-query';
import { UserAccount } from '@neosync/sdk';
import { getUserAccounts } from '@neosync/sdk/connectquery';
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

  const {
    data: accountsResponse,
    isLoading,
    refetch: mutate,
  } = useQuery(getUserAccounts);
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
    if (userAccount && foundAccount && userAccount.id === foundAccount.id) {
      return;
    }
    if (foundAccount) {
      setUserAccount(foundAccount);
      setLastSelectedAccount(foundAccount.name);
      const accountParam = getSingleOrUndefined(account);
      if (!accountParam || accountParam !== foundAccount.name) {
        router.push(`/${foundAccount.name}/jobs`);
      }
    } else if (accountName !== DEFAULT_ACCOUNT_NAME) {
      setLastSelectedAccount(DEFAULT_ACCOUNT_NAME);
      router.push(`/${DEFAULT_ACCOUNT_NAME}/jobs`);
    }
  }, [
    userAccount?.id,
    userAccount?.name,
    userAccount?.type,
    accountsResponse?.accounts.length,
    isLoading,
    accountName,
  ]);

  function setAccount(userAccount: UserAccount): void {
    if (userAccount.name !== accountName) {
      // this order matters. Otherwise if we push first,
      // when it routes to the page, there is no account param and it defaults to personal /shrug
      // by setting this here, it finds the last selected account and is able to effectively route to the correct spot.
      setLastSelectedAccount(userAccount.name);
      setUserAccount(userAccount);
      router.push(`/${userAccount.name}`);
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
