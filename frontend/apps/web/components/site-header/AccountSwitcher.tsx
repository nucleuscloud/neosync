'use client';
import {
  CaretSortIcon,
  CheckIcon,
  PlusCircledIcon,
} from '@radix-ui/react-icons';
import { ReactElement, ReactNode } from 'react';

import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { cn } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { CreateTeamFormValues } from '@/yup-validations/account-switcher';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { UserAccount, UserAccountType } from '@neosync/sdk';
import { createTeamAccount, getUserAccounts } from '@neosync/sdk/connectquery';
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../ui/dialog';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '../ui/form';
import { Input } from '../ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '../ui/popover';
import { Skeleton } from '../ui/skeleton';

interface Props {}

export default function AccountSwitcher(_: Props): ReactElement | null {
  const { account, setAccount } = useAccount();
  const { data, refetch: mutate, isLoading } = useQuery(getUserAccounts);
  const [showNewTeamDialog, setShowNewTeamDialog] = useState(false);
  const { data: systemAppConfigData } = useGetSystemAppConfig();
  const accounts = data?.accounts ?? [];
  const router = useRouter();

  const { mutateAsync: createTeamAccountAsync } =
    useMutation(createTeamAccount);

  async function onSubmit(values: CreateTeamFormValues): Promise<void> {
    try {
      const resp = await createTeamAccountAsync({
        name: values.name,
      });
      setShowNewTeamDialog(false);
      mutate();
      toast.success('Successfully created team!');
      if (resp.checkoutSessionUrl) {
        router.push(resp.checkoutSessionUrl);
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to create team', {
        description: getErrorMessage(err),
      });
    }
  }
  if (isLoading) {
    return <Skeleton className=" h-full w-[200px]" />;
  }

  return (
    <CreateNewTeamDialog
      open={showNewTeamDialog}
      onOpenChange={setShowNewTeamDialog}
      onSubmit={onSubmit}
      onCancel={() => setShowNewTeamDialog(false)}
      trigger={
        <AccountSwitcherPopover
          activeAccount={account}
          accounts={accounts}
          onAccountSelect={(a) => setAccount(a)}
          onNewAccount={() => {
            if (systemAppConfigData?.isNeosyncCloud) {
              router.push(`/${account?.name}/settings/billing`);
              return;
            }
            setShowNewTeamDialog(true);
          }}
        />
      }
    />
  );
}

interface AccountSwitcherPopoverProps {
  activeAccount: UserAccount | undefined;
  accounts: UserAccount[];
  onAccountSelect(account: UserAccount): void;
  onNewAccount(): void;
}

function AccountSwitcherPopover(
  props: AccountSwitcherPopoverProps
): ReactElement {
  const { activeAccount, accounts, onAccountSelect, onNewAccount } = props;
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
          {
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
          }
        </Command>
      </PopoverContent>
    </Popover>
  );
}

interface CreateNewTeamDialogProps {
  open: boolean;
  onOpenChange(val: boolean): void;
  trigger?: ReactNode;

  onSubmit(values: CreateTeamFormValues): Promise<void>;
  onCancel(): void;
}

function CreateNewTeamDialog(props: CreateNewTeamDialogProps): ReactElement {
  const { open, onOpenChange, trigger, onCancel, onSubmit } = props;
  const form = useForm<CreateTeamFormValues>({
    mode: 'onChange',
    resolver: yupResolver(CreateTeamFormValues),
    defaultValues: {
      name: '',
    },
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {trigger}
      <DialogContent className="flex flex-col gap-3">
        <DialogHeader>
          <DialogTitle>Create team</DialogTitle>
          <DialogDescription>
            Create a new team account to collaborate with your co-workers.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <Form {...form}>
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <Input
                      autoCapitalize="off" // we don't allow capitals in team names
                      data-1p-ignore // tells 1password extension to not autofill this field
                      placeholder="acme"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </Form>
        </div>
        <DialogFooter>
          <div className="flex flex-row justify-between w-full pt-6">
            <Button variant="outline" onClick={() => onCancel()}>
              Cancel
            </Button>
            <Button
              type="submit"
              onClick={(e) =>
                form.handleSubmit((values) => onSubmit(values))(e)
              }
              disabled={!form.formState.isValid}
            >
              Continue
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
