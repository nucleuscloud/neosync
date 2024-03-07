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
  SetAccountOnboardingConfigRequest,
  SetAccountOnboardingConfigResponse,
} from '@neosync/sdk';
import {
  ArrowRightIcon,
  ChevronDownIcon,
  CircleIcon,
  RocketIcon,
} from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { FaCheckCircle } from 'react-icons/fa';
import { useAccount } from '../providers/account-provider';
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
const stepsMap = {
  hasCreatedSourceConnection: {
    title: 'Create a source connection',
    href: '/connection',
  },
  hasCreatedDestinationConnection: {
    title: 'Create a destination connection',
    href: '/connection',
  },
  hasCreatedJob: { title: 'Create a job', href: '/job' },
  hasInvitedMembers: { title: 'Invite members', href: '/settings' },
};

export default function OnboardingChecklist() {
  const { account } = useAccount();
  const { data, isLoading, isValidating } = useGetAccountOnboardingConfig(
    account?.id ?? ''
  );
  const [progress, setProgress] = useState<number>(0);
  const [isOpen, setIsOpen] = useState(false);
  const [values, setValues] = useState<Step[]>([]);
  const [showGuide, setShowGuide] = useState<boolean>(false);

  useEffect(() => {
    if (data?.config) {
      // if account is a personal account remove the invite a team member step
      const updatedConfig =
        account?.type == 1
          ? (({ hasInvitedMembers, ...rest }) => rest)(data.config)
          : data.config;

      // check if any of the steps are false, if so, then surface the guide
      if (account?.type == 1 && !didUserCompleteOnboarding(updatedConfig)) {
        setShowGuide(true);
        setIsOpen(true);
      }

      // construct the stepsArray so that we can map over it.
      const stepsArray = Object.entries(updatedConfig).map(([key, value]) => {
        return {
          id: key,
          title: stepsMap[key as keyof typeof stepsMap].title,
          href: stepsMap[key as keyof typeof stepsMap].href,
          complete: value,
        };
      });
      setValues(stepsArray);

      // calculate the progress percentage
      const progressPercentage =
        account?.type == 1
          ? calculateProgress(updatedConfig)
          : calculateProgress(data.config);
      setProgress(progressPercentage);
    }
  }, [isValidating]);

  const router = useRouter();

  if (isLoading) {
    return <Skeleton />;
  }

  async function completeForm() {
    setProgress(100);
    setValues((prevValues) =>
      prevValues.map((step) =>
        step.id === step.id ? { ...step, complete: true } : step
      )
    );
    setShowGuide(false);
    try {
      await setOnboardingConfig(account?.id ?? '', {
        hasCreatedDestinationConnection: true,
        hasCreatedSourceConnection: true,
        hasCreatedJob: true,
        hasInvitedMembers: true,
      });
      setIsOpen(false);
    } catch (e) {
      console.log('unable to complete form');
      toast({
        title: 'Unable to complete onboarding',
        variant: 'destructive',
      });
    }
  }

  return (
    <>
      {showGuide && (
        <div className="fixed right-[20px] bottom-[20px] z-50">
          <Popover
            onOpenChange={(open) => {
              if (open) {
                setIsOpen(true);
              }
            }}
            open={isOpen}
          >
            <PopoverTrigger className="border border-gray-300 shadow-lg rounded-lg p-2 text-sm hover:bg-gray-100">
              <div className="flex flex-row items-center gap-2">
                <div>Onboarding Guide</div>
                <div
                  className={`transform transition-transform ${isOpen ? 'rotate-180' : ''}`}
                >
                  <ChevronDownIcon className="w-4 h-4" />
                </div>
              </div>
            </PopoverTrigger>
            <PopoverContent className="w-[400px]" sideOffset={10} align="end">
              <div className="flex flex-col gap-4 p-2">
                <div className="flex flex-col gap-2">
                  <div className="flex flex-row gap-2 items-center">
                    <div>
                      <RocketIcon />
                    </div>
                    <div className="font-semibold">Welcome to Neosync!</div>
                  </div>
                  <div className="text-sm pl-6">
                    Follow these steps to get started
                  </div>
                </div>
                <div className="flex flex-row gap-2 items-center">
                  <Progress value={progress} />
                  <div className="text-sm">{progress}%</div>
                </div>
                <Separator />
                <div className="flex flex-col gap-2">
                  {values.map((step) => (
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
                      <Button
                        variant="ghost"
                        onClick={() =>
                          step.title == 'Invite members'
                            ? router.push(`/${account?.name}/${step.href}/`)
                            : router.push(`/${account?.name}/new/${step.href}/`)
                        }
                      >
                        <ArrowRightIcon className="w-4 h-4" />
                      </Button>
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
      )}
    </>
  );
}

function calculateProgress(data: OnboardingValues): number {
  const totalSteps = Object.keys(data).length;
  const completedSteps = Object.values(data).filter(Boolean).length;
  return Math.round((completedSteps / totalSteps) * 100);
}

function didUserCompleteOnboarding(data: OnboardingValues): boolean {
  return Object.values(data).every(Boolean);
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
