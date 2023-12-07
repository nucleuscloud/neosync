'use client';
import { useGetUserAccounts } from '@/libs/hooks/useUserAccounts';
import { UserAccount } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
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
  setAccount: (updatedAccount: UserAccount) => void;
  isLoading: boolean;
}
const AccountContext = createContext<AccountContextType>({
  account: undefined,
  setAccount: () => {},
  isLoading: false,
});

const USER_ACCOUNT_KEY = 'user-account';

interface Props {
  children: ReactNode;
  params: Record<string, string>;
}

export default function AccountProvider(props: Props): ReactElement {
  const { children, params } = props;

  const accountName = params.account ?? 'personal';

  const { data: accountsResponse, isLoading, mutate } = useGetUserAccounts();
  // const [] = useState(accountName);

  // const [userAccount, setUserAccount] = useLocalStorage<
  //   UserAccount | undefined
  // >(USER_ACCOUNT_KEY, undefined);
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

    // if (
    //   !userAccount &&
    //   accountsResponse?.accounts &&
    //   accountsResponse.accounts.length > 0
    // ) {
    //   setUserAccount(accountsResponse.accounts[0]);
    // } else if (
    //   userAccount &&
    //   accountsResponse?.accounts &&
    //   !accountsResponse.accounts.some((acc) => acc.id === userAccount.id)
    // ) {
    //   setUserAccount(accountsResponse.accounts[0]);
    // }
  }, [userAccount, accountsResponse?.accounts.length, isLoading, accountName]);

  function setAccount(userAccount?: UserAccount): void {
    // mutate().then(() => setUserAccount(userAccount));
    setUserAccount(userAccount);
  }

  return (
    <AccountContext.Provider
      value={{ account: userAccount, setAccount: setAccount, isLoading }}
    >
      {children}
    </AccountContext.Provider>
  );
}

export function useAccount(): AccountContextType {
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
