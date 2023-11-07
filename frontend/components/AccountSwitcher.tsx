'use client';
import {
  CaretSortIcon,
  CheckIcon,
  PlusCircledIcon,
} from '@radix-ui/react-icons';
import * as React from 'react';
import { ReactElement } from 'react';

import { useGetUserAccounts } from '@/libs/hooks/useUserAccounts';
import { cn } from '@/libs/utils';
import {
  CreateTeamAccountRequest,
  CreateTeamAccountResponse,
  UserAccountType,
} from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { getErrorMessage } from '@/util/util';
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
  CommandSeparator,
} from './ui/command';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './ui/dialog';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';
import { Skeleton } from './ui/skeleton';
import { toast } from './ui/use-toast';

interface Props {}

export default function AccountSwitcher(_: Props): ReactElement {
  const { account, setAccount } = useAccount();
  const { data, mutate, isLoading } = useGetUserAccounts();
  const [open, setOpen] = React.useState(false);
  const [showNewTeamDialog, setShowNewTeamDialog] = React.useState(false);
  const [teamName, setTeamName] = React.useState('');

  const personalAccounts =
    data?.accounts.filter((a) => a.type == UserAccountType.PERSONAL) || [];
  const teamAccounts =
    data?.accounts.filter((a) => a.type == UserAccountType.TEAM) || [];

  async function onSubmit(teamName: string): Promise<void> {
    try {
      await createTeamAccount(teamName);
      setShowNewTeamDialog(false);
      mutate();
      toast({
        title: 'Successfully created team!',
        variant: 'success',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to create team',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }
  if (isLoading) {
    return <Skeleton className=" h-full w-[200px]" />;
  }

  return (
    <Dialog open={showNewTeamDialog} onOpenChange={setShowNewTeamDialog}>
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
            <CommandSeparator />
            <CommandList>
              <CommandGroup>
                <DialogTrigger asChild>
                  <CommandItem
                    onSelect={() => {
                      setOpen(false);
                      setShowNewTeamDialog(true);
                    }}
                  >
                    <PlusCircledIcon className="mr-2 h-5 w-5" />
                    Create Team
                  </CommandItem>
                </DialogTrigger>
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create team</DialogTitle>
          <DialogDescription>Add a new team to manage jobs.</DialogDescription>
        </DialogHeader>
        <div>
          <div className="space-y-4 py-2 pb-4">
            <div className="space-y-2">
              <Label htmlFor="name">Team name</Label>
              <Input
                id="name"
                placeholder="Acme Inc."
                onChange={(event) => setTeamName(event.target.value)}
              />
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setShowNewTeamDialog(false)}>
            Cancel
          </Button>
          <Button type="submit" onClick={() => onSubmit(teamName)}>
            Continue
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

async function createTeamAccount(
  teamName: string
): Promise<CreateTeamAccountResponse | undefined> {
  const res = await fetch(`/api/teams`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new CreateTeamAccountRequest({
        name: teamName,
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateTeamAccountResponse.fromJson(await res.json());
}
