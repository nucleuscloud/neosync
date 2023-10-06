'use client';
import { useGetUserAccounts } from '@/libs/hooks/useUserAccounts';
import { UserAccount } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import {
  ReactElement,
  ReactNode,
  createContext,
  useContext,
  useEffect,
} from 'react';
import { useLocalStorage } from 'usehooks-ts';

const AccountContext = createContext<UserAccount | undefined>(undefined);

const USER_ACCOUNT_KEY = 'user-account';

interface Props {
  children: ReactNode;
}

export default function AccountProvider(props: Props): ReactElement {
  const { children } = props;
  const { data: accountsResponse, isLoading } = useGetUserAccounts();

  const [userAccount, setUserAccount] = useLocalStorage<
    UserAccount | undefined
  >(USER_ACCOUNT_KEY, undefined);

  useEffect(() => {
    if (
      !userAccount &&
      accountsResponse?.accounts &&
      accountsResponse.accounts.length > 0
    ) {
      setUserAccount(accountsResponse.accounts[0]);
    } else if (
      userAccount &&
      accountsResponse?.accounts &&
      !accountsResponse.accounts.some((acc) => acc.id === userAccount.id)
    ) {
      setUserAccount(accountsResponse.accounts[0]);
    }
  }, [userAccount, accountsResponse?.accounts.length, isLoading]);

  return (
    <AccountContext.Provider value={userAccount}>
      {children}
    </AccountContext.Provider>
  );
}

export function useAccount(): UserAccount | undefined {
  const account = useContext(AccountContext);
  return account;
}

// Retrieves the account from local storage. Useful if need to retrieve outside of a hook
export function getAccount(): UserAccount | undefined {
  if (!localStorage) {
    return undefined;
  }
  const item = localStorage.getItem(USER_ACCOUNT_KEY);
  if (!item) {
    return undefined;
  }
  try {
    const val = JSON.parse(item) as UserAccount;
    return val;
  } catch (err) {
    console.error(err);
    return undefined;
  }
}
