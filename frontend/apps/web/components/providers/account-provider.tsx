'use client';
import { useNeosyncUser } from '@/libs/hooks/useNeosyncUser';
import { getSingleOrUndefined } from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { UserAccount, UserAccountService } from '@neosync/sdk';
import { useParams, useRouter } from 'next/navigation';
import {
  ReactElement,
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react';
import { useLocalStorage, useSessionStorage } from 'usehooks-ts';

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
  const accountName = useGetAccountName();

  // Use both session and local storage
  const [, setLastSelectedAccountSession] = useSessionStorage(
    STORAGE_ACCOUNT_KEY,
    accountName ?? DEFAULT_ACCOUNT_NAME
  );
  const [, setLastSelectedAccountLocal] = useLocalStorage(
    STORAGE_ACCOUNT_KEY,
    accountName ?? DEFAULT_ACCOUNT_NAME
  );

  const { isLoading: isUserLoading } = useNeosyncUser();

  const {
    data: accountsResponse,
    isLoading,
    refetch: mutate,
  } = useQuery(UserAccountService.method.getUserAccounts, undefined, {
    enabled: !isUserLoading,
  });

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
      // Update both storages
      setLastSelectedAccountSession(foundAccount.name);
      setLastSelectedAccountLocal(foundAccount.name);
    } else if (accountName !== DEFAULT_ACCOUNT_NAME) {
      // Update both storages
      setLastSelectedAccountSession(DEFAULT_ACCOUNT_NAME);
      setLastSelectedAccountLocal(DEFAULT_ACCOUNT_NAME);
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
      // Update both storages before routing
      setLastSelectedAccountSession(userAccount.name);
      setLastSelectedAccountLocal(userAccount.name);
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
  // Check session storage first, then fall back to local storage
  const [sessionAccount] = useSessionStorage(
    STORAGE_ACCOUNT_KEY,
    account ?? DEFAULT_ACCOUNT_NAME
  );
  const [localAccount] = useLocalStorage(
    STORAGE_ACCOUNT_KEY,
    account ?? DEFAULT_ACCOUNT_NAME
  );

  const accountParam = getSingleOrUndefined(account);
  if (accountParam) {
    return accountParam;
  }
  // Prefer session storage account over local storage
  const singleSessionAccount = getSingleOrUndefined(sessionAccount);
  if (singleSessionAccount) {
    return singleSessionAccount;
  }
  const singleLocalAccount = getSingleOrUndefined(localAccount);
  if (singleLocalAccount) {
    return singleLocalAccount;
  }
  return DEFAULT_ACCOUNT_NAME;
}

export function useAccount(): AccountContextType {
  const account = useContext(AccountContext);
  return account;
}
