'use client';

import { ReactElement } from 'react';

import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import * as React from 'react';

import { useGetUserAccounts } from '@/libs/hooks/useUserAccounts';
import { cn } from '@/libs/utils';
import { UserAccountType } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { useAccount } from './providers/account-provider';
import { Avatar, AvatarFallback, AvatarImage } from './ui/avatar';
import { Button } from './ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from './ui/command';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';

interface Props {}

export default function AccountSwitcher(_: Props): ReactElement {
  const { account, setAccount } = useAccount();
  const { data } = useGetUserAccounts();
  const [open, setOpen] = React.useState(false);

  const personalAccounts =
    data?.accounts.filter((a) => a.type == UserAccountType.PERSONAL) || [];
  const teamAccounts =
    data?.accounts.filter((a) => a.type == UserAccountType.TEAM) || [];

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
              src={`https://avatar.vercel.sh/${account?.id}.png`}
              alt={account?.name}
            />
            <AvatarFallback>SC</AvatarFallback>
          </Avatar>
          {account?.name}
          <CaretSortIcon className="ml-auto h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <Command>
          <CommandList>
            <CommandInput placeholder="Search account..." />
            <CommandEmpty>No Account found.</CommandEmpty>
            <CommandGroup key="personal" heading="Personal">
              {personalAccounts.map((a) => (
                <CommandItem
                  key={a.id}
                  onSelect={() => {
                    setAccount(a);
                    setOpen(false);
                  }}
                  className="text-sm"
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
                      account?.id === a.id ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                </CommandItem>
              ))}
            </CommandGroup>
            <CommandGroup key="team" heading="Team">
              {teamAccounts.map((a) => (
                <CommandItem
                  key={a.id}
                  onSelect={() => {
                    setAccount(a);
                    setOpen(false);
                  }}
                  className="text-sm"
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
                      account?.id === a.id ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
