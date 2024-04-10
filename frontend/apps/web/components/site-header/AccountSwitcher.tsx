'use client';
import {
  CaretSortIcon,
  CheckIcon,
  PlusCircledIcon,
} from '@radix-ui/react-icons';
import { ReactElement } from 'react';

import { useGetUserAccounts } from '@/libs/hooks/useUserAccounts';
import { cn } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { RESOURCE_NAME_REGEX } from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CreateTeamAccountRequest,
  CreateTeamAccountResponse,
  UserAccountType,
} from '@neosync/sdk';
import Link from 'next/link';
import { useState } from 'react';
import { UseFormReturn, useForm } from 'react-hook-form';
import * as Yup from 'yup';
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
import { Popover, PopoverContent, PopoverTrigger } from '../ui/popover';
import { Skeleton } from '../ui/skeleton';
import { toast } from '../ui/use-toast';

export const CreateTeamFormValues = Yup.object({
  name: Yup.string()
    .required()
    .min(3)
    .max(30)
    .test(
      'valid account name',
      'Account Name must be of length 3-30 and only include lowercased letters, numbers, and/or hyphens.',
      (value) => {
        if (!value || value.length < 3) {
          return false;
        }
        if (!RESOURCE_NAME_REGEX.test(value)) {
          return false;
        }
        // todo: test to make sure that account is valid across neosync
        return true;
      }
    ),
});
export type CreateTeamFormValues = Yup.InferType<typeof CreateTeamFormValues>;

interface Props {}

export default function AccountSwitcher(_: Props): ReactElement {
  const { account, setAccount } = useAccount();
  const { data, mutate, isLoading } = useGetUserAccounts();
  const [open, setOpen] = useState(false);
  const [showNewTeamDialog, setShowNewTeamDialog] = useState(false);
  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(CreateTeamFormValues),
    defaultValues: {
      name: '',
    },
  });

  const accounts = data?.accounts ?? [];
  const personalAccounts =
    accounts.filter((a) => a.type === UserAccountType.PERSONAL) ?? [];
  const teamAccounts =
    accounts.filter((a) => a.type === UserAccountType.TEAM) ?? [];

  async function onSubmit(values: CreateTeamFormValues): Promise<void> {
    // add acount type here
    try {
      await createTeamAccount(values.name);
      setShowNewTeamDialog(false);
      mutate();
      toast({
        title: 'Successfully created team!',
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
                    className="cursor-pointer"
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
      <CreateNewTeamDialog
        form={form}
        onSubmit={onSubmit}
        setShowNewTeamDialog={setShowNewTeamDialog}
        planType={account?.type ?? UserAccountType.PERSONAL}
      />
    </Dialog>
  );
}

interface CreateNewTeamDialogProps {
  form: UseFormReturn<
    {
      name: string;
    },
    any,
    undefined
  >;
  onSubmit: (values: CreateTeamFormValues) => Promise<void>;
  setShowNewTeamDialog: (val: boolean) => void;
  planType: UserAccountType;
}

export function CreateNewTeamDialog(
  props: CreateNewTeamDialogProps
): ReactElement {
  const { form, onSubmit, setShowNewTeamDialog, planType } = props;
  return (
    <div>
      {/* {(planType && planType == UserAccountType.PERSONAL) ||
      planType == UserAccountType.TEAM ? ( */}
      <UpgradeDialog planType={planType} />
      {/* ) : (
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
                      <Input placeholder="acme" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </Form>
          </div>
          <DialogFooter>
            <div className="flex flex-row justify-between w-full pt-6">
              <Button
                variant="outline"
                onClick={() => setShowNewTeamDialog(false)}
              >
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
      )} */}
    </div>
  );
}

interface UpgradeDialog {
  planType: UserAccountType;
}

function UpgradeDialog({ planType }: UpgradeDialog) {
  return (
    <div>
      <DialogContent className="flex flex-col gap-3">
        <DialogHeader>
          <DialogTitle>Upgrade your plan to create a Team</DialogTitle>
          <DialogDescription>
            Contact us in order to create a new team.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <div className="flex flex-row w-full pt-6 justify-center">
            <Button>
              <Link
                href="https://calendly.com/evis1/30min"
                className="w-[242px]"
                target="_blank"
              >
                Get in touch
              </Link>
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </div>
  );
}

export async function createTeamAccount(
  teamName: string
): Promise<CreateTeamAccountResponse | undefined> {
  const res = await fetch(`/api/users/accounts`, {
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
