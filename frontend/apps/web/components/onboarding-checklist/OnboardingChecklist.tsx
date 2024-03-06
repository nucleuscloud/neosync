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
import { ArrowRightIcon, CircleIcon, RocketIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { FaCheckCircle } from 'react-icons/fa';
import { useAccount } from '../providers/account-provider';
import { Button } from '../ui/button';
import { Separator } from '../ui/separator';
import { Skeleton } from '../ui/skeleton';

interface OnboardingValues {
  hasCreatedSourceConnection: boolean;
  hasCreatedDestinationConnection: boolean;
  hasCreatedJob: boolean;
  hasInvitedMembers: boolean;
}

interface Step {
  id: string;
  title: string;
  href: string;
  complete: boolean;
}

export default function OnboardingChecklist() {
  const { account } = useAccount();
  const { data, isLoading } = useGetAccountOnboardingConfig(account?.id ?? '');
  const [progress, setProgress] = useState<number>(0);
  const [complete, isComplete] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [error, setError] = useState(false);
  const [values, setValues] = useState<Step[]>([]);

  // calculate the progress percentage
  // we're always doing a wholesale replace when we update the db so including just one field in the dependency array should be fine
  useEffect(() => {
    if (data?.config) {
      const stepsArray = Object.entries(data.config).map(([key, value]) => {
        return {
          id: key,
          title: stepsMap[key as keyof typeof stepsMap].title,
          href: stepsMap[key as keyof typeof stepsMap].href,
          complete: value,
        };
      });

      setValues(stepsArray);
      const progressPercentage = calculateProgress(data.config);
      setProgress(progressPercentage);
    }
  }, [data?.config?.hasCreatedDestinationConnection]);

  const router = useRouter();

  if (isLoading) {
    return <Skeleton />;
  }

  // create a map of {step:metadata} we show in UI
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

  async function completeForm() {
    setProgress(100);
    setValues((prevValues) =>
      prevValues.map((step) =>
        step.id === step.id ? { ...step, complete: true } : step
      )
    );
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
      setError(true);
    }
  }

  console.log('step array', values);

  return (
    <div className="fixed right-[160px] bottom-[20px] z-50">
      <Popover
        onOpenChange={() => setIsOpen(isOpen ? false : true)}
        open={isOpen}
      >
        <PopoverTrigger className="border border-gray-300 rounded-lg p-2">
          {isOpen ? 'Close Guide' : 'Open Guide'}
        </PopoverTrigger>
        <PopoverContent className="w-[400px]">
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
            <Separator />
            <div className=" flex flex-row items-center justify-end pt-6">
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

function calculateProgress(data: OnboardingValues): number {
  const totalSteps = Object.keys(data).length;
  const completedSteps = Object.values(data).filter(Boolean).length;
  return (completedSteps / totalSteps) * 100;
}

async function setOnboardingConfig(
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
