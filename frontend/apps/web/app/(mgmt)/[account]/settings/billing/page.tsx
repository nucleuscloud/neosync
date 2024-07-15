'use client';
import { useAccount } from '@/components/providers/account-provider';
import {
  CreateNewTeamDialog,
  CreateTeamFormValues,
} from '@/components/site-header/AccountSwitcher';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog } from '@/components/ui/dialog';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { useToast } from '@/components/ui/use-toast';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { getErrorMessage, toTitleCase } from '@/util/util';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { UserAccountType } from '@neosync/sdk';
import { createTeamAccount, getUserAccounts } from '@neosync/sdk/connectquery';
import { CheckCircledIcon, DiscordLogoIcon } from '@radix-ui/react-icons';
import Error from 'next/error';
import Link from 'next/link';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';

const plans = [
  {
    name: 'Personal',
    features: [
      '100k records/month',
      'Unlimited Jobs',
      '1 user',
      'US Region',
      'Social Login',
      'Shared Infrastructure',
      'Community Discord',
    ],
    price: 'Free',
  },
  {
    name: 'Team',
    features: [
      'Unlimited record',
      'Unlimited Jobs',
      'Unlimited users',
      'US and EU Regions',
      'Social Login, SSO',
      'Shared Infrastructure',
      'Private Discord/Slack',
    ],
    price: 'Contact us',
  },
  {
    name: 'Enterprise',
    features: [
      'Unlimited Records',
      'Unlimited Jobs',
      'Unlimited Users',
      'Dedicated Infrastructure',
      'Hybrid Deployment',
      'Social Login, SSO',
      'Private Discord/Slack',
    ],
    price: 'Contact Us',
  },
];

export default function Billing(): ReactElement {
  const { account } = useAccount();
  const { data: systemAppConfigData, isLoading: isSystemAppConfigDataLoading } =
    useGetSystemAppConfig();
  const { toast } = useToast();
  const [showNewTeamDialog, setShowNewTeamDialog] = useState<boolean>(false);
  const { refetch: mutate } = useQuery(getUserAccounts);
  const { mutateAsync: createTeamAccountAsync } =
    useMutation(createTeamAccount);
  async function onSubmit(values: CreateTeamFormValues): Promise<void> {
    try {
      await createTeamAccountAsync({ name: values.name });
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

  if (isSystemAppConfigDataLoading || !account?.type) {
    return <Skeleton />;
  }

  if (!systemAppConfigData?.isNeosyncCloud) {
    return <Error statusCode={404} />;
  }

  return (
    <div>
      <div className="flex flex-col gap-10">
        <div>
          <div className="text-xl font-seminold">Billing</div>
          <div className="text-sm">
            {`Manage your workspace's plan and billing information`}
          </div>
          <Separator />
        </div>
        <NeedHelp />
        <Plans
          planType={account.type ?? UserAccountType.PERSONAL}
          setShowNewTeamDialog={setShowNewTeamDialog}
          showNewTeamDialog={showNewTeamDialog}
          onSubmit={onSubmit}
          upgradeHref={systemAppConfigData.calendlyUpgradeLink}
        />
      </div>
    </div>
  );
}

function NeedHelp(): ReactElement {
  return (
    <div className="flex flex-col gap-2">
      <div className="text-md">Need help?</div>
      <div className="text-lg">
        <div className="flex flex-row items-center gap-2">
          <div className="text-sm">Ask us on</div>
          <Link
            href="https://discord.com/invite/MFAMgnp4HF"
            className="flex flex-row items-center text-sm gap-2 underline"
          >
            <DiscordLogoIcon className="w-4 h-4" />
            <div>Discord</div>
          </Link>
          <div className="text-sm">
            or send an email to <u>support@neosync.dev</u>
          </div>
        </div>
      </div>
    </div>
  );
}

interface PlanProps {
  planType: UserAccountType;
  setShowNewTeamDialog: (val: boolean) => void;
  showNewTeamDialog: boolean;
  onSubmit: (values: CreateTeamFormValues) => Promise<void>;
  upgradeHref: string;
}

function Plans({
  planType,
  showNewTeamDialog,
  setShowNewTeamDialog,
  onSubmit,
  upgradeHref,
}: PlanProps): ReactElement {
  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(CreateTeamFormValues),
    defaultValues: {
      name: '',
    },
  });

  return (
    <div>
      <div className="border border-gray-200 rounded-xl p-6 flex flex-col gap-4">
        <div className="flex flex-row items-center gap-2 text-sm">
          <div className="font-semibold">Current Plan:</div>
          <div>
            <Badge>{toTitleCase(UserAccountType[planType])} Plan</Badge>
          </div>
        </div>
        <div>
          <div className="flex flex-row justify-between items-center"></div>
        </div>
        <Separator />
        <div className="flex flex-row items-center gap-2 justify-center">
          <div className="flex flex-col lg:flex-row gap-2 pt-6">
            {plans.map((plan, ind) => (
              <div key={plan.name}>
                {planType == ind + 1 && (
                  <div className="flex justify-center bg-gradient-to-t from-[#191919] to-[#484848] text-white p-4 shadow-lg rounded-t-xl">
                    Current Plan
                  </div>
                )}
                <div
                  className={
                    planType == ind + 1
                      ? `flex flex-col items-center gap-2 border-4 border-gray-800 p-6 rounded-b-xl lg:w-[350px] h-[459px]`
                      : `flex flex-col items-center gap-2 border border-gray-300 p-6 rounded-xl lg:w-[350px] mt-[56px] h-[459px]`
                  }
                >
                  <div className="flex flex-col gap-6">
                    <div className="flex justify-center">
                      <Badge variant="outline">{plan.name} Plan</Badge>
                    </div>
                    <div className="flex justify-center flex-row gap-2">
                      <div className="text-3xl ">{plan.price}</div>
                      {plan.name == 'Team' && (
                        <div className="text-sm self-end pb-1">/mo</div>
                      )}
                    </div>
                    <Dialog
                      open={showNewTeamDialog}
                      onOpenChange={setShowNewTeamDialog}
                    >
                      <CreateNewTeamDialog
                        form={form}
                        onSubmit={onSubmit}
                        setShowNewTeamDialog={setShowNewTeamDialog}
                        planType={planType}
                        upgradeHref={upgradeHref}
                      />
                    </Dialog>
                    <div className="flex flex-col gap-2">
                      {plan.features.map((feat) => (
                        <div
                          key={feat}
                          className="flex flex-row items-center gap-2"
                        >
                          <CheckCircledIcon className="w-4 h-4 text-green-800 bg-green-200 rounded-full" />
                          <div>{feat}</div>
                        </div>
                      ))}
                    </div>
                    <PlanButton
                      plan={plan.name}
                      planType={planType}
                      upgradeHref={upgradeHref}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

interface PlanButtonProps {
  plan: string;
  planType: UserAccountType;
  upgradeHref: string;
}

function PlanButton(props: PlanButtonProps): ReactElement {
  const { plan, planType, upgradeHref } = props;
  switch (plan) {
    case 'Personal':
      if (planType == UserAccountType.PERSONAL) {
        return <div></div>;
      } else {
        return <Button>Get Started</Button>;
      }
    case 'Team':
      if (planType == UserAccountType.TEAM) {
        return <div></div>;
      } else {
        return (
          <Button>
            <Link href={upgradeHref} className="w-[242px]" target="_blank">
              Get in touch
            </Link>
          </Button>
        );
      }
    case 'Enterprise':
      if (planType == UserAccountType.ENTERPRISE) {
        return <div></div>;
      } else {
        return (
          <Button>
            <Link href={upgradeHref} className="w-[242px]" target="_blank">
              Get in touch
            </Link>
          </Button>
        );
      }
  }
  return <div />;
}
