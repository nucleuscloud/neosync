'use client';
import { ExternalLinkIcon } from '@radix-ui/react-icons';
import { ReactElement, ReactNode } from 'react';

import { cn } from '@/libs/utils';
import { CreateTeamFormValues } from '@/yup-validations/account-switcher';
import Link from 'next/link';
import { UseFormReturn } from 'react-hook-form';
import { Button } from '../ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '../ui/form';
import { Input } from '../ui/input';
import { Switch } from '../ui/switch';

interface Props {
  open: boolean;
  onOpenChange(val: boolean): void;
  trigger?: ReactNode;

  onSubmit(values: CreateTeamFormValues): Promise<void>;
  onCancel(): void;
  showSubscriptionInfo: boolean;
  showConvertPersonalToTeamOption: boolean;
  form: UseFormReturn<CreateTeamFormValues>;
}

export function CreateNewTeamDialog(props: Props): ReactElement<any> {
  const {
    open,
    onOpenChange,
    trigger,
    onCancel,
    onSubmit,
    showSubscriptionInfo,
    showConvertPersonalToTeamOption,
    form,
  } = props;

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
        <div className="space-y-2 py-2">
          <Form {...form}>
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Team Name</FormLabel>
                  <FormControl>
                    <Input
                      autoCapitalize="off" // we don't allow capitals in team names
                      data-1p-ignore // tells 1password extension to not autofill this field
                      placeholder="acme"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage className="mt-0" />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="convertPersonalToTeam"
              render={({ field }) => (
                <FormItem
                  className={cn(
                    showConvertPersonalToTeamOption ? undefined : 'hidden'
                  )}
                >
                  <FormLabel>Personal Account Conversion</FormLabel>
                  <FormDescription>
                    Move all settings from personal to team account.
                  </FormDescription>
                  <FormControl className="m-0">
                    <Switch
                      ref={field.ref}
                      onBlur={field.onBlur}
                      onCheckedChange={(newval) => field.onChange(newval)}
                      checked={field.value}
                      disabled={field.disabled}
                      name={field.name}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
          </Form>
          {showSubscriptionInfo && <ShowSubscriptionInfo />}
        </div>
        <DialogFooter>
          <div className="flex flex-row justify-between w-full">
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

interface ShowSubscriptionInfoProps {}
function ShowSubscriptionInfo(props: ShowSubscriptionInfoProps): ReactElement<any> {
  const {} = props;

  return (
    <div>
      <div className="flex flex-row gap-2">
        <p className="text-sm tracking-tight">
          Continuing will start a monthly Team plan subscription.
        </p>
        <Link
          href="https://neosync.dev/pricing"
          target="_blank"
          className="hover:underline inline-flex gap-1 flex-row items-center text-sm tracking-tight"
        >
          Learn More
          <ExternalLinkIcon className="w-3 h-3" />
        </Link>
      </div>
    </div>
  );
}
