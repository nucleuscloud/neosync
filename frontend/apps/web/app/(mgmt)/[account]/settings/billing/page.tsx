'use client';
import ButtonText from '@/components/ButtonText';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { useGetOnCreateTeamSubmit } from '@/components/site-header/AccountSwitcher';
import { CreateNewTeamDialog } from '@/components/site-header/CreateNewTeamDialog';
import Spinner from '@/components/Spinner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { cn } from '@/libs/utils';
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
import { CheckIcon } from '@radix-ui/react-icons';
import Error from 'next/error';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';

type PlanName = 'Trial' | 'Team' | 'Enterprise';

interface Plan {
  name: PlanName;
  description: string;
  price: string;
  buttonText: string;
  features: string[];
  planType: UserAccountType;
}

const ALL_PLANS: Plan[] = [
  {
    name: 'Trial',
    description:
      'The easiest way to try out Neosync and see if it is right for your organization.',
    price: 'Free 14-day Trial',
    buttonText: 'Start free trial',
    features: [
      'Unlimited Records',
      'Unlimited Jobs',
      '1 user',
      'All Integrations',
      'Data Anonymization',
      'Synthetic Data Generation',
      'US Region',
      'Neosync Infrastructure',
      'Community Discord',
    ],
    planType: UserAccountType.PERSONAL,
  },
  {
    name: 'Team',
    description:
      'The simplest and fastest way to get started with Neosync for Teams.',
    price: 'Pay-as-you-go',
    buttonText: 'Get Started',
    features: [
      'Volume-based pricing',
      'Unlimited Jobs',
      'Unlimited Users',
      'US or EU Region',
      'Free-form text PII anonymization',
      'Private Discord/Slack',
    ],
    planType: UserAccountType.TEAM,
  },
  {
    name: 'Enterprise',
    description:
      'Best for organizations who want to self-host Neosync and are operating at scale.',
    price: 'Custom',
    buttonText: `Let's talk`,
    features: [
      'Unlimited Records',
      'SSO',
      'RBAC',
      'Webhooks',
      'Self-hosted or Neosync Infrastructure',
      'Audit Logs',
      'Streaming Mode',
      'White-glove implementation',
    ],
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
    <Button type="button" onClick={() => onManageSubscriptionClick()}>
      <ButtonText
        leftIcon={isGeneratingUrl ? <Spinner /> : null}
        text="Manage Subscription"
      />
    </Button>
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
    <div className="flex flex-col gap-3">
      <div className="flex flex-row gap-3 justify-between items-center ">
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
      <div className="grid grid-cols1 md:grid-cols-3 grid-rows-1 w-full border border-gray-300 dark:border-gray-700 bg-transparent rounded-lg mt-4 h-full shadow-lg">
        {plans.map((plan, index) => (
          <div
            className={cn(
              index == 1 && 'bg-gray-100 dark:bg-gray-800',
              'flex flex-col'
            )}
            key={plan.name}
          >
            <div
              className={cn(
                index == 1 && 'border-x border-gray-300 dark:border-gray-700',
                'h-1/2 px-6 pt-6 flex flex-col justify-between gap-20 w-full'
              )}
            >
              <div className="space-y-2">
                <div className=" items-center flex flex-row gap-4">
                  <div className="text-2xl font-semibold">{plan.name}</div>
                  {plan.name == 'Team' && (
                    <Badge className="bg-blue-700 hover:bg-blue-700">
                      Most Popular
                    </Badge>
                  )}
                </div>
                <div className="text-sm text-foreground">
                  {plan.description}
                </div>
              </div>
              <div className="flex flex-col gap-6 w-full">
                <div className="text-2xl md:text-3xl flex flex-row gap-2 items-end">
                  {plan.price}{' '}
                  {plan.name == 'Team' && (
                    <div className="text-sm  text-foreground">per month</div>
                  )}
                </div>
                <div></div>
                <div className={cn(plan.name !== 'Trial' && 'pb-6', 'w-full')}>
                  <PlanButton
                    plan={plan.name}
                    planType={account.type}
                    upgradeHref={upgradeHref}
                    accountSlug={account.name}
                    buttonText={plan.buttonText}
                  />
                  <div className="text-md flex">
                    {plan.name == 'Trial' && '*No credit card required'}
                  </div>
                </div>
              </div>
            </div>
            <div
              className={cn(
                index == 1 && 'border-x border-gray-300 dark:border-gray-700',
                'border-t border-gray-300 dark:border-gray-700 h-1/2 p-6 flex flex-col gap-6'
              )}
            >
              <div className="font-bold">
                {plan.name == 'Team'
                  ? 'Everything in Individual + '
                  : plan.name == 'Enterprise' && 'Everything in Team + '}
              </div>
              <div className="flex flex-col gap-2">
                {plan.features.map((item, index) => (
                  <div key={index}>
                    <div className="flex flex-row items-center gap-2">
                      <CheckIcon className="w-6 h-6" />
                      <div>{item}</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

interface PlanButtonProps {
  plan: string;
  planType: UserAccountType;
  upgradeHref: string;
  accountSlug: string;
  buttonText: string;
}

function PlanButton(props: PlanButtonProps) {
  const { plan, planType, upgradeHref, accountSlug } = props;
  switch (plan) {
    case 'Trial':
      if (planType == UserAccountType.PERSONAL) {
        return <GetStartedButton accountSlug={accountSlug} />;
      } else {
        return <div id="empty-personal-button" />;
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
    <Button type="button" className="w-full">
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
        <Button className="w-full" onClick={() => setShowNewTeamDialog(true)}>
          Create Team
        </Button>
      }
      showSubscriptionInfo={true} // This is only rendered in neosync cloud
      showConvertPersonalToTeamOption={
        account?.type === UserAccountType.PERSONAL
      }
    />
  );
}
