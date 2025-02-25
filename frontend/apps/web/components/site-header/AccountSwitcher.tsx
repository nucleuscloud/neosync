'use client';
import {
  CaretSortIcon,
  CheckIcon,
  PlusCircledIcon,
} from '@radix-ui/react-icons';
import { ReactElement } from 'react';

import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { cn } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { CreateTeamFormValues } from '@/yup-validations/account-switcher';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { UserAccount, UserAccountService, UserAccountType } from '@neosync/sdk';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { useAccount } from '../providers/account-provider';
import { Avatar, AvatarFallback, AvatarImage } from '../ui/avatar';
import { Button } from '../ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '../ui/command';
import { DialogTrigger } from '../ui/dialog';
import { Popover, PopoverContent, PopoverTrigger } from '../ui/popover';
import { Skeleton } from '../ui/skeleton';
import { CreateNewTeamDialog } from './CreateNewTeamDialog';

interface Props {}

export default function AccountSwitcher(_: Props): ReactElement | null {
  const { account, setAccount } = useAccount();
  const { data, isLoading } = useQuery(
    UserAccountService.method.getUserAccounts
  );
  const [showNewTeamDialog, setShowNewTeamDialog] = useState(false);
  const { data: systemAppConfigData } = useGetSystemAppConfig();
  const accounts = data?.accounts ?? [];
  const createNewTeamForm = useForm<CreateTeamFormValues>({
    mode: 'onChange',
    resolver: yupResolver(CreateTeamFormValues),
    defaultValues: {
      name: '',
      convertPersonalToTeam: false,
    },
  });

  const onSubmit = useGetOnCreateTeamSubmit({
    onDone() {
      setShowNewTeamDialog(false);
    },
  });

  if (isLoading) {
    return <Skeleton className=" h-full w-[200px]" />;
  }

  return (
    <CreateNewTeamDialog
      form={createNewTeamForm}
      open={showNewTeamDialog}
      onOpenChange={setShowNewTeamDialog}
      onSubmit={onSubmit}
      onCancel={() => setShowNewTeamDialog(false)}
      trigger={
        <AccountSwitcherPopover
          activeAccount={account}
          accounts={accounts}
          onAccountSelect={(a) => {
            setAccount(a);
            // the user has changed the active account, reset the create new team form
            createNewTeamForm.reset();
          }}
          onNewAccount={() => {
            setShowNewTeamDialog(true);
          }}
          showCreateTeamDialog={
            !systemAppConfigData?.isNeosyncCloud ||
            (systemAppConfigData.isNeosyncCloud &&
              systemAppConfigData.isStripeEnabled)
          }
        />
      }
      showSubscriptionInfo={
        (systemAppConfigData?.isNeosyncCloud ?? false) &&
        (systemAppConfigData?.isStripeEnabled ?? false)
      }
      showConvertPersonalToTeamOption={
        account?.type === UserAccountType.PERSONAL
      }
    />
  );
}

interface AccountSwitcherPopoverProps {
  activeAccount: UserAccount | undefined;
  accounts: UserAccount[];
  onAccountSelect(account: UserAccount): void;
  onNewAccount(): void;
  showCreateTeamDialog: boolean;
}

function AccountSwitcherPopover(
  props: AccountSwitcherPopoverProps
): ReactElement {
  const {
    activeAccount,
    accounts,
    onAccountSelect,
    onNewAccount,
    showCreateTeamDialog,
  } = props;
  const [open, setOpen] = useState(false);

  const personalAccounts =
    accounts.filter((a) => a.type === UserAccountType.PERSONAL) ?? [];
  const teamAccounts =
    accounts.filter(
      (a) =>
        a.type === UserAccountType.TEAM || a.type === UserAccountType.ENTERPRISE
    ) ?? [];

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          aria-label="Select a team"
          className="w-[200px] justify-between"
        >
          <Avatar className="mr-2 h-5 w-5">
            <AvatarImage
              src={`https://avatar.vercel.sh/${activeAccount?.id}.png`}
              alt={activeAccount?.name}
            />
            <AvatarFallback>SC</AvatarFallback>
          </Avatar>
          {activeAccount?.name}
          <CaretSortIcon className="ml-auto h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <Command>
          <CommandInput placeholder="Search account..." />
          <CommandList>
            <CommandEmpty>No Account found.</CommandEmpty>
            {personalAccounts.length > 0 && (
              <CommandGroup key="personal" heading="Personal">
                {personalAccounts.map((a) => (
                  <CommandItem
                    key={a.id}
                    onSelect={() => {
                      onAccountSelect(a);
                      setOpen(false);
                    }}
                    className="text-sm cursor-pointer"
                  >
                    <Avatar className="mr-2 h-5 w-5">
                      <AvatarImage
                        src={`https://avatar.vercel.sh/${a.name}.png`}
                        alt={a.name}
                        className="grayscale"
                      />
                      <AvatarFallback>SC</AvatarFallback>
                    </Avatar>
                    {a.name}
                    <CheckIcon
                      className={cn(
                        'ml-auto h-4 w-4',
                        activeAccount?.id === a.id ? 'opacity-100' : 'opacity-0'
                      )}
                    />
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
            {teamAccounts.length > 0 && (
              <CommandGroup key="team" heading="Team">
                {teamAccounts.map((a) => (
                  <CommandItem
                    key={a.id}
                    onSelect={() => {
                      onAccountSelect(a);
                      setOpen(false);
                    }}
                    className="text-sm cursor-pointer"
                  >
                    <Avatar className="mr-2 h-5 w-5">
                      <AvatarImage
                        src={`https://avatar.vercel.sh/${a.name}.png`}
                        alt={a.name}
                        className="grayscale"
                      />
                    </Avatar>
                    {a.name}
                    <CheckIcon
                      className={cn(
                        'ml-auto h-4 w-4',
                        activeAccount?.id === a.id ? 'opacity-100' : 'opacity-0'
                      )}
                    />
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
          <CommandSeparator />
          {showCreateTeamDialog && (
            <CommandList>
              <CommandGroup>
                <DialogTrigger asChild>
                  <CommandItem
                    onSelect={() => {
                      setOpen(false);
                      onNewAccount();
                    }}
                    className="cursor-pointer"
                  >
                    <PlusCircledIcon className="mr-2 h-5 w-5" />
                    Create Team
                  </CommandItem>
                </DialogTrigger>
              </CommandGroup>
            </CommandList>
          )}
        </Command>
      </PopoverContent>
    </Popover>
  );
}

interface UseGetOnCreateTeamSubmitProps {
  onDone?(): void;
}

export function useGetOnCreateTeamSubmit(
  props: UseGetOnCreateTeamSubmitProps
): (values: CreateTeamFormValues) => Promise<void> {
  const { onDone = () => undefined } = props;

  const { account, setAccount } = useAccount();

  const { mutateAsync: createTeamAccountAsync } = useMutation(
    UserAccountService.method.createTeamAccount
  );
  const { mutateAsync: convertPersonalToTeamAccountAsync } = useMutation(
    UserAccountService.method.convertPersonalToTeamAccount
  );
  const { refetch: refreshUserAccountsAsync } = useQuery(
    UserAccountService.method.getUserAccounts
  );
  const router = useRouter();

  return async (values) => {
    if (!account) {
      return;
    }
    if (
      values.convertPersonalToTeam &&
      account?.type !== UserAccountType.PERSONAL
    ) {
      toast.error(
        'Selected account must be personal account to issue account conversion.'
      );
      return;
    }
    try {
      if (values.convertPersonalToTeam) {
        const resp = await convertPersonalToTeamAccountAsync({
          name: values.name,
          accountId: account.id,
        });
        const mutatedResp = await refreshUserAccountsAsync();
        toast.success('Successfully converted personal to team!');

        if (resp.checkoutSessionUrl) {
          onDone();
          router.push(resp.checkoutSessionUrl);
        } else {
          const newAcc = mutatedResp.data?.accounts.find(
            (a) => a.name === values.name
          );
          if (newAcc) {
            setAccount(newAcc);
          } else {
            toast.error(
              'Team was created but was unable to navigate to new team. Please try refreshing the page.'
            );
          }
          onDone();
        }
      } else {
        const resp = await createTeamAccountAsync({
          name: values.name,
        });
        const mutatedResp = await refreshUserAccountsAsync();
        toast.success('Successfully created team!');
        if (resp.checkoutSessionUrl) {
          onDone();
          router.push(resp.checkoutSessionUrl);
        } else {
          const newAcc = mutatedResp.data?.accounts.find(
            (a) => a.name === values.name
          );
          if (newAcc) {
            setAccount(newAcc);
          } else {
            toast.error(
              'Team was created but was unable to navigate to new team. Please try refreshing the page.'
            );
          }
          onDone();
        }
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to create team', {
        description: getErrorMessage(err),
      });
    }
  };
}
