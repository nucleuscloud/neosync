'use client';
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog';
import { getErrorMessage } from '@/util/util';
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from '@connectrpc/connect-query';
import { GetAccountOnboardingConfigResponse } from '@neosync/sdk';
import {
  getAccountOnboardingConfig,
  setAccountOnboardingConfig,
} from '@neosync/sdk/connectquery';
import { DialogDescription } from '@radix-ui/react-dialog';
import { useQueryClient } from '@tanstack/react-query';
import { ReactElement, useEffect, useState } from 'react';
import { toast } from 'sonner';
import { useAccount } from '../providers/account-provider';
import { Skeleton } from '../ui/skeleton';
import HowItWorks from './HowItWorks';
import StepProgress from './StepProgress';
import WelcomeOverview from './WelcomeOverview';
import WelcomeRouter from './WelcomeRouter';

export interface FormStep {
  name: string;
  component: JSX.Element;
}

export default function WelcomeDialog(): ReactElement {
  const { account } = useAccount();
  const {
    data,
    isLoading,
    isFetching: isValidating,
    error,
  } = useQuery(
    getAccountOnboardingConfig,
    { accountId: account?.id ?? '' },
    { enabled: !!account?.id }
  );
  const queryclient = useQueryClient();
  const { mutateAsync: setOnboardingConfigAsync } = useMutation(
    setAccountOnboardingConfig
  );

  const [currentStep, setCurrentStep] = useState<number>(0); // controls progress through the welcome modal
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [isComplete, setIsComplete] = useState<boolean>(false);
  const [showGuide, setShowGuide] = useState<boolean>(false);

  const multiStepForm: FormStep[] = [
    {
      name: 'welcome',
      component: (
        <WelcomeOverview
          currentStep={currentStep}
          setCurrentStep={setCurrentStep}
          setIsDialogOpen={setShowGuide}
          completeForm={completeForm}
        />
      ),
    },
    {
      name: 'how-it-works',
      component: (
        <HowItWorks currentStep={currentStep} setCurrentStep={setCurrentStep} />
      ),
    },
    {
      name: 'router',
      component: (
        <WelcomeRouter
          currentStep={currentStep}
          setCurrentStep={setCurrentStep}
          setIsDialogOpen={setShowGuide}
          completeForm={completeForm}
        />
      ),
    },
  ];

  async function completeForm() {
    if (!account?.id || isSubmitting) {
      return;
    }
    setIsSubmitting(true);
    try {
      const resp = await setOnboardingConfigAsync({
        accountId: account.id,
        config: {
          hasCompletedOnboarding: isComplete,
        },
      });
      queryclient.setQueryData(
        createConnectQueryKey(getAccountOnboardingConfig, {
          accountId: account.id,
        }),
        new GetAccountOnboardingConfigResponse({
          config: resp.config,
        })
      );
      setShowGuide(false);
      setIsComplete(true);
    } catch (e) {
      toast.error('Unable to complete onboarding', {
        description: getErrorMessage(e),
      });
    } finally {
      setIsSubmitting(false);
      setIsComplete(false);
    }
  }

  useEffect(() => {
    if (error || isLoading) {
      return;
    }

    setIsComplete(data?.config?.hasCompletedOnboarding ?? false);
    if (isComplete) {
      {
        setTimeout(() => setShowGuide(false), 1000);
        return;
      }
    }
    if (!isComplete) {
      setShowGuide(true);
    }
  }, [isLoading, isValidating, error, isComplete]);

  if (isLoading) {
    return <Skeleton />;
  }
  if (!showGuide) {
    return <></>;
  }

  console.log('isCOmplete', isComplete);

  return (
    <Dialog open={showGuide} onOpenChange={setShowGuide}>
      <DialogContent className="max-w-2xl">
        <DialogTitle />
        <DialogDescription />
        <div className="flex flex-col gap-8 pt-6">
          <StepProgress
            steps={multiStepForm.map((step) => step)}
            currentStep={currentStep}
          />
          <div className="px-8">{multiStepForm[currentStep].component}</div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
