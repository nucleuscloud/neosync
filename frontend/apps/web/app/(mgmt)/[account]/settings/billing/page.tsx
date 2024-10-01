'use client';
import ButtonText from '@/components/ButtonText';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { useGetOnCreateTeamSubmit } from '@/components/site-header/AccountSwitcher';
import { CreateNewTeamDialog } from '@/components/site-header/CreateNewTeamDialog';
import Spinner from '@/components/Spinner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { getErrorMessage, toTitleCase } from '@/util/util';
import { CreateTeamFormValues } from '@/yup-validations/account-switcher';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { UserAccount, UserAccountType } from '@neosync/sdk';
import {
  getAccountBillingCheckoutSession,
  getAccountBillingPortalSession,
  isAccountStatusValid,
} from '@neosync/sdk/connectquery';
import { CheckCircledIcon, DiscordLogoIcon } from '@radix-ui/react-icons';
import Error from 'next/error';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';

type PlanName = 'Personal' | 'Team' | 'Enterprise';

interface Plan {
  name: PlanName;
  features: string[];
  price: string;
  planType: UserAccountType;
}

const ALL_PLANS: Plan[] = [
  {
    name: 'Personal',
    features: [
      '20k records/month',
      'Unlimited Jobs',
      '1 user',
      'US Region',
      'Social Login',
      'Neosync Infrastructure',
      'Community Discord',
    ],
    price: 'Free',
    planType: UserAccountType.PERSONAL,
  },
  {
    name: 'Team',
    features: [
      'Volume-based pricing',
      'Unlimited Jobs',
      'Unlimited Users',
      'US and EU Regions',
      'Social, SSO',
      'Neosync Infrastructure',
      'Private Discord/Slack',
    ],
    price: 'Pay as you go',
    planType: UserAccountType.TEAM,
  },
  {
    name: 'Enterprise',
    features: [
      'Unlimited Records',
      'Unlimited Jobs',
      'Unlimited Users',
      'Dedicated Infrastructure',
      'Hybrid Deployment',
      'Social, SSO',
      'Private Discord/Slack',
    ],
    price: 'Contact Us',
    planType: UserAccountType.ENTERPRISE,
  },
];

export default function Billing(): ReactElement {
  const { account } = useAccount();
  const { data: systemAppConfigData, isLoading: isSystemAppConfigDataLoading } =
    useGetSystemAppConfig();

  if (isSystemAppConfigDataLoading || !account?.type) {
    return <Skeleton />;
  }

  if (!systemAppConfigData?.isNeosyncCloud) {
    return <Error statusCode={404} />;
  }

  return (
    <div className="flex flex-col gap-5">
      <SubPageHeader
        header="Billing"
        description="Manage your workspace's plan and billing information"
      />
      <div className="py-4">
        <NeedHelp />
      </div>
      <Plans
        account={account}
        upgradeHref={systemAppConfigData?.calendlyUpgradeLink ?? ''}
        plans={ALL_PLANS}
        isStripeEnabled={systemAppConfigData?.isStripeEnabled ?? false}
      />
    </div>
  );
}

interface ManageSubscriptionProps {
  account: UserAccount;
}

function ManageSubscription(props: ManageSubscriptionProps): ReactElement {
  const { account } = props;

  const { data: isAccountStatusValidResp, isLoading } = useQuery(
    isAccountStatusValid,
    { accountId: account.id },
    { enabled: !!account.id }
  );

  const { mutateAsync: getAccountBillingPortalSessionAsync } = useMutation(
    getAccountBillingPortalSession
  );
  const { mutateAsync: getAccountBillingCheckoutSessionAsync } = useMutation(
    getAccountBillingCheckoutSession
  );
  const router = useRouter();
  const [isGeneratingUrl, setIsGeneratingUrl] = useState(false);

  if (account.type === UserAccountType.PERSONAL) {
    return <div />;
  }
  async function onActivateSubscriptionClick(): Promise<void> {
    if (isGeneratingUrl) {
      return;
    }
    setIsGeneratingUrl(true);
    try {
      const resp = await getAccountBillingCheckoutSessionAsync({
        accountId: account.id,
      });
      router.push(resp.checkoutSessionUrl);
    } catch (err) {
      console.error(err);
      toast.error('Unable to generate billing checkout session url', {
        description: getErrorMessage(err),
      });
    } finally {
      setIsGeneratingUrl(false);
    }
  }
  if (account.hasStripeCustomerId && isLoading) {
    return <Skeleton />;
  }

  if (!account.hasStripeCustomerId || !isAccountStatusValidResp?.isValid) {
    return (
      <div>
        <Button type="button" onClick={() => onActivateSubscriptionClick()}>
          <ButtonText
            leftIcon={isGeneratingUrl ? <Spinner /> : null}
            text="Activate Subscription"
          />
        </Button>
      </div>
    );
  }

  async function onManageSubscriptionClick(): Promise<void> {
    if (isGeneratingUrl) {
      return;
    }
    setIsGeneratingUrl(true);
    try {
      const resp = await getAccountBillingPortalSessionAsync({
        accountId: account.id,
      });
      router.push(resp.portalSessionUrl);
    } catch (err) {
      console.error(err);
      toast.error('Unable to generate billing portal session url', {
        description: getErrorMessage(err),
      });
    } finally {
      setIsGeneratingUrl(false);
    }
  }

  return (
    <div>
      <Button type="button" onClick={() => onManageSubscriptionClick()}>
        <ButtonText
          leftIcon={isGeneratingUrl ? <Spinner /> : null}
          text="Manage Subscription"
        />
      </Button>
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

interface PlansProps {
  account: UserAccount;
  plans: Plan[];
  upgradeHref: string;
  isStripeEnabled: boolean;
}

function Plans({
  account,
  upgradeHref,
  plans,
  isStripeEnabled,
}: PlansProps): ReactElement {
  return (
    <div className="border border-gray-200 rounded-xl">
      <div className="flex flex-col gap-3">
        <div>
          <div className="flex flex-row gap-3 justify-between items-center px-6">
            <div className="flex flex-row items-center gap-2 text-sm py-6">
              <p className="font-semibold">Current Plan:</p>
              <Badge>{toTitleCase(UserAccountType[account.type])} Plan</Badge>
            </div>
            {isStripeEnabled && (
              <div>
                <ManageSubscription account={account} />
              </div>
            )}
          </div>
          <Separator className="dark:bg-gray-600" />
        </div>
        <div className="flex flex-col xl:flex-row gap-2 justify-center p-6">
          {plans.map((plan) => (
            <PlanInfo
              key={plan.name}
              plan={plan}
              activePlan={account.type}
              upgradeHref={upgradeHref}
              accountSlug={account.name}
            />
          ))}
        </div>
      </div>
    </div>
  );
}

interface PlanInfoProps {
  plan: Plan;
  activePlan: UserAccountType;
  upgradeHref: string;
  accountSlug: string;
}
function PlanInfo(props: PlanInfoProps): ReactElement {
  const { plan, activePlan, upgradeHref, accountSlug } = props;
  const isCurrentPlan = activePlan === plan.planType;
  return (
    <div>
      {isCurrentPlan && <CurrentPlan />}
      <div
        className={
          isCurrentPlan
            ? `flex flex-col items-center gap-2 border-4 border-gray-800 p-6 rounded-b-xl xl:w-[350px] h-[459px]`
            : `flex flex-col items-center gap-2 border border-gray-300 p-6 rounded-xl xl:w-[350px] mt-[56px] h-[459px]`
        }
      >
        <div className="flex flex-col gap-6">
          <div className="flex justify-center">
            <Badge variant="outline">{plan.name}</Badge>
          </div>
          <div className="flex justify-center flex-row gap-2">
            <div className="text-3xl">{plan.price}</div>
          </div>
          <div className="flex flex-col gap-2">
            {plan.features.map((feat) => (
              <div key={feat} className="flex flex-row items-center gap-2">
                <CheckCircledIcon className="w-4 h-4 text-green-800 bg-green-200 rounded-full" />
                <div>{feat}</div>
              </div>
            ))}
          </div>
          <PlanButton
            plan={plan.name}
            planType={activePlan}
            upgradeHref={upgradeHref}
            accountSlug={accountSlug}
          />
        </div>
      </div>
    </div>
  );
}

function CurrentPlan(): ReactElement {
  return (
    <div className="flex justify-center bg-gradient-to-t from-[#191919] to-[#484848] text-white p-4 shadow-lg rounded-t-xl">
      <p>Current Plan</p>
    </div>
  );
}

interface PlanButtonProps {
  plan: PlanName;
  planType: UserAccountType;
  upgradeHref: string;
  accountSlug: string;
}

function PlanButton(props: PlanButtonProps): ReactElement {
  const { plan, planType, upgradeHref, accountSlug } = props;
  switch (plan) {
    case 'Personal':
      if (planType == UserAccountType.PERSONAL) {
        return <GetStartedButton accountSlug={accountSlug} />;
      } else {
        return <></>;
      }
    case 'Team':
      if (planType == UserAccountType.TEAM) {
        return <GetStartedButton accountSlug={accountSlug} />;
      } else {
        return <CreateNewTeamButton />;
      }
    case 'Enterprise':
      if (planType == UserAccountType.ENTERPRISE) {
        return <GetStartedButton accountSlug={accountSlug} />;
      } else {
        return (
          <Button type="button">
            <Link href={upgradeHref} target="_blank">
              Get in touch
            </Link>
          </Button>
        );
      }
  }
}

interface GetStartedButtonProps {
  accountSlug: string;
}

function GetStartedButton(props: GetStartedButtonProps): ReactElement {
  const { accountSlug } = props;
  return (
    <Button type="button">
      <Link href={`/${accountSlug}/new/job`}>Get Started</Link>
    </Button>
  );
}

function CreateNewTeamButton(): ReactElement {
  const form = useForm<CreateTeamFormValues>({
    mode: 'onChange',
    resolver: yupResolver(CreateTeamFormValues),
    defaultValues: {
      name: '',
      convertPersonalToTeam: false,
    },
  });
  const [showNewTeamDialog, setShowNewTeamDialog] = useState(false);
  const { account } = useAccount();

  const onSubmit = useGetOnCreateTeamSubmit({
    onDone() {
      setShowNewTeamDialog(false);
    },
  });

  return (
    <CreateNewTeamDialog
      form={form}
      open={showNewTeamDialog}
      onOpenChange={(val) => {
        setShowNewTeamDialog(val);
        form.reset();
      }}
      onSubmit={onSubmit}
      onCancel={() => {
        setShowNewTeamDialog(false);
        form.reset();
      }}
      trigger={
        <Button onClick={() => setShowNewTeamDialog(true)}>Create Team</Button>
      }
      showSubscriptionInfo={true} // This is only rendered in neosync cloud
      showConvertPersonalToTeamOption={
        account?.type === UserAccountType.PERSONAL
      }
    />
  );
}
