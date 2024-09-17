'use client';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { toTitleCase } from '@/util/util';
import { UserAccountType } from '@neosync/sdk';
import { CheckCircledIcon, DiscordLogoIcon } from '@radix-ui/react-icons';
import Error from 'next/error';
import Link from 'next/link';
import { ReactElement } from 'react';

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
      'Shared Infrastructure',
      'Community Discord',
    ],
    price: 'Free',
    planType: UserAccountType.PERSONAL,
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
      'Social Login, SSO',
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
        accountType={account.type}
        upgradeHref={systemAppConfigData.calendlyUpgradeLink}
        plans={ALL_PLANS}
      />
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
  accountType: UserAccountType;
  plans: Plan[];
  upgradeHref: string;
}

function Plans({ accountType, upgradeHref, plans }: PlansProps): ReactElement {
  return (
    <div className="border border-gray-200 rounded-xl">
      <div className="flex flex-col gap-3">
        <div>
          <div className="flex flex-row items-center gap-2 text-sm p-6">
            <p className="font-semibold">Current Plan:</p>
            <Badge>{toTitleCase(UserAccountType[accountType])} Plan</Badge>
          </div>
          <Separator className="dark:bg-gray-600" />
        </div>
        <div className="flex flex-col xl:flex-row gap-2 justify-center p-6">
          {plans.map((plan) => (
            <PlanInfo
              key={plan.name}
              plan={plan}
              activePlan={accountType}
              upgradeHref={upgradeHref}
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
}
function PlanInfo(props: PlanInfoProps): ReactElement {
  const { plan, activePlan, upgradeHref } = props;
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
            <Badge variant="outline">{plan.name} Plan</Badge>
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
}

function PlanButton(props: PlanButtonProps): ReactElement {
  const { plan, planType, upgradeHref } = props;
  switch (plan) {
    case 'Personal':
      if (planType == UserAccountType.PERSONAL) {
        return <div></div>;
      } else {
        return <Button type="button">Get Started</Button>;
      }
    case 'Team':
      if (planType == UserAccountType.TEAM) {
        return <div></div>;
      } else {
        return (
          <Button type="button">
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
          <Button type="button">
            <Link href={upgradeHref} className="w-[242px]" target="_blank">
              Get in touch
            </Link>
          </Button>
        );
      }
  }
}
