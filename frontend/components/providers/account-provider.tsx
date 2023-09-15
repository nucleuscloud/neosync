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

export const AccountContext = createContext<UserAccount | undefined>(undefined);

interface Props {
  children: ReactNode;
}

export default function AccountProvider(props: Props): ReactElement {
  const { children } = props;
  const { data: accountsResponse, isLoading } = useGetUserAccounts();

  const [userAccount, setUserAccount] = useLocalStorage<
    UserAccount | undefined
  >('user-account', undefined);

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
