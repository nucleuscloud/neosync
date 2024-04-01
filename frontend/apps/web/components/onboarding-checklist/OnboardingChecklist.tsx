'use client';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Progress } from '@/components/ui/progress';
import { useGetAccountOnboardingConfig } from '@/libs/hooks/useGetAccountOnboardingConfig';
import {
  AccountOnboardingConfig,
  GetAccountOnboardingConfigResponse,
  SetAccountOnboardingConfigRequest,
  SetAccountOnboardingConfigResponse,
  UserAccountType,
} from '@neosync/sdk';
import {
  ArrowRightIcon,
  ChevronDownIcon,
  CircleIcon,
  RocketIcon,
} from '@radix-ui/react-icons';
import NextLink from 'next/link';
import { ReactElement, useEffect, useState } from 'react';
import { FaCheckCircle } from 'react-icons/fa';
import Spinner from '../Spinner';
import { useAccount } from '../providers/account-provider';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Separator } from '../ui/separator';
import { Skeleton } from '../ui/skeleton';
import { toast } from '../ui/use-toast';

interface OnboardingValues {
  hasCreatedSourceConnection: boolean;
  hasCreatedDestinationConnection: boolean;
  hasCreatedJob: boolean;
  hasInvitedMembers?: boolean;
}

interface Step {
  id: string;
  title: string;
  href: string;
  complete: boolean;
}

// create a map of {step:metadata} so we can construct that we need to we show in UI
const STEPS_METADATA: Record<
  keyof OnboardingValues,
  Pick<Step, 'title' | 'href'>
> = {
  hasCreatedSourceConnection: {
    title: 'Create a source connection',
    href: '/new/connection',
  },
  hasCreatedDestinationConnection: {
    title: 'Create a destination connection',
    href: '/new/connection',
  },
  hasCreatedJob: { title: 'Create a Job', href: '/new/job' },
  hasInvitedMembers: { title: 'Invite members', href: '/settings' },
};

export default function OnboardingChecklist(): ReactElement {
  const { account } = useAccount();
  const { data, isLoading, isValidating, mutate, error } =
    useGetAccountOnboardingConfig(account?.id ?? '');
  const [isOpen, setIsOpen] = useState(false);
  const [showGuide, setShowGuide] = useState<boolean>(false);
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

  const onboardingValues: OnboardingValues = data?.config
    ? {
        ...data.config,
        hasInvitedMembers:
          account?.type === UserAccountType.PERSONAL
            ? undefined
            : data.config.hasInvitedMembers,
      }
    : {
        hasCreatedDestinationConnection: false,
        hasCreatedJob: false,
        hasCreatedSourceConnection: false,
        hasInvitedMembers:
          account?.type === UserAccountType.PERSONAL ? undefined : false,
      };

  const progress = calculateProgress(onboardingValues);
  const isCompleted = isChecklistComplete(progress);

  useEffect(() => {
    if (error || isLoading) {
      return;
    }
    if (showGuide && isCompleted) {
      {
        setTimeout(() => setShowGuide(false), 1000);
        return;
      }
    }
    if (!showGuide && !isCompleted) {
      setShowGuide(true);
    }
  }, [isLoading, isValidating, error, isCompleted]);

  if (isLoading) {
    return <Skeleton />;
  }
  if (!showGuide) {
    return <></>;
  }

  async function completeForm() {
    if (!account?.id || isSubmitting) {
      return;
    }
    setIsSubmitting(true);
    try {
      const resp = await setOnboardingConfig(account.id, {
        hasCreatedDestinationConnection: true,
        hasCreatedSourceConnection: true,
        hasCreatedJob: true,
        hasInvitedMembers: true,
      });
      mutate(
        new GetAccountOnboardingConfigResponse({
          config: resp.config,
        })
      );
      setIsOpen(false);
    } catch (e) {
      toast({
        title: 'Unable to complete onboarding',
        variant: 'destructive',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  const steps = getSteps(onboardingValues);

  return (
    <div className="fixed right-[20px] bottom-[20px] z-50">
      <Popover onOpenChange={setIsOpen} open={isOpen}>
        <PopoverTrigger
          className="bg-background border border-gray-300 dark:border-gray-700 shadow-lg rounded-lg p-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer"
          asChild
        >
          <div className="flex flex-row items-center gap-2">
            <p className="select-none">Onboarding Guide</p>
            <Badge className="px-1">{progress}%</Badge>
            <div
              className={`transform transition-transform ${isOpen ? 'rotate-180' : ''}`}
            >
              <ChevronDownIcon className="w-4 h-4" />
            </div>
          </div>
        </PopoverTrigger>
        <PopoverContent className="md:w-[400px]" sideOffset={10} align="end">
          <div className="flex flex-col gap-4 p-2">
            <div className="flex flex-col gap-2">
              <div className="flex flex-row gap-2 items-center">
                <div>
                  <RocketIcon />
                </div>
                <p className="font-semibold">Welcome to Neosync!</p>
                {isValidating ? <Spinner /> : null}
              </div>
              <p className="text-sm pl-6">Follow these steps to get started</p>
            </div>
            <div className="flex flex-row gap-2 items-center">
              <Progress value={progress} />
              <div className="text-sm">{progress}%</div>
            </div>
            <Separator />
            <div className="flex flex-col gap-2">
              {steps.map((step) => (
                <div
                  className="flex flex-row items-center justify-between"
                  key={step.id}
                >
                  <div className="flex flex-row items-center gap-2">
                    <div>
                      {step.complete ? (
                        <FaCheckCircle className="text-green-600 w-4 h-4 " />
                      ) : (
                        <CircleIcon />
                      )}
                    </div>
                    <div className="text-sm">{step.title}</div>
                  </div>
                  <NextLink href={`/${account?.name}/${step.href}`}>
                    <Button variant="ghost">
                      <ArrowRightIcon className="w-4 h-4" />
                    </Button>
                  </NextLink>
                </div>
              ))}
            </div>

            <div className=" flex flex-row items-center justify-between pt-6">
              <Button variant="outline" onClick={() => setIsOpen(false)}>
                Close
              </Button>
              <Button variant="default" onClick={completeForm}>
                Complete
              </Button>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}

function getSteps(data: OnboardingValues): Step[] {
  return Object.entries(data)
    .filter(([_, value]) => isBoolean(value))
    .map(
      ([key, value]): Step => ({
        id: key,
        complete: value,
        title: STEPS_METADATA[key as keyof OnboardingValues].title,
        href: STEPS_METADATA[key as keyof OnboardingValues].href,
      })
    );
}

function calculateProgress(data: OnboardingValues): number {
  const validValues = Object.values(data).filter(isBoolean);
  const totalSteps = validValues.length;
  const completedSteps = validValues.filter((x) => x).length;
  return Math.round((completedSteps / totalSteps) * 100);
}

function isChecklistComplete(progress: number): boolean {
  return progress === 100;
}

function isBoolean(input: unknown): input is boolean {
  return typeof input === 'boolean';
}

export async function setOnboardingConfig(
  accountId: string,
  values: OnboardingValues
): Promise<SetAccountOnboardingConfigResponse> {
  const res = await fetch(
    `/api/users/accounts/${accountId}/onboarding-config`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new SetAccountOnboardingConfigRequest({
          accountId,
          config: new AccountOnboardingConfig({
            hasCreatedSourceConnection: values.hasCreatedSourceConnection,
            hasCreatedDestinationConnection:
              values.hasCreatedDestinationConnection,
            hasCreatedJob: values.hasCreatedJob,
            hasInvitedMembers: values.hasInvitedMembers,
          }),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return SetAccountOnboardingConfigResponse.fromJson(await res.json());
}
