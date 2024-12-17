'use client';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { getErrorMessage } from '@/util/util';
import { create } from '@bufbuild/protobuf';
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from '@connectrpc/connect-query';
import { GetAccountOnboardingConfigResponseSchema } from '@neosync/sdk';
import {
  getAccountOnboardingConfig,
  setAccountOnboardingConfig,
} from '@neosync/sdk/connectquery';
import { useQueryClient } from '@tanstack/react-query';
import { ReactElement, useEffect, useState } from 'react';
import { toast } from 'sonner';
import { useAccount } from '../providers/account-provider';
import Configure from './Configure';
import Connect from './Connect';
import StepProgress from './StepProgress';
import Sync from './Sync';
import WelcomeOverview from './WelcomeOverview';
import WelcomeRouter from './WelcomeRouter';

export type FormStepName =
  | 'welcome'
  | 'connect'
  | 'configure'
  | 'execute'
  | 'router';

export interface FormStep {
  name: FormStepName;
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

  const [currentStep, setCurrentStep] = useState<FormStepName>('welcome');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [showGuide, setShowGuide] = useState<boolean | null>(null);

  const handleNextStep = () => {
    setCurrentStep((prev) => {
      switch (prev) {
        case 'welcome':
          return 'connect';
        case 'connect':
          return 'configure';
        case 'configure':
          return 'execute';
        case 'execute':
          return 'router';
        case 'router':
        default:
          return prev;
      }
    });
  };

  const handlePreviousStep = () => {
    setCurrentStep((prev) => {
      switch (prev) {
        case 'connect':
          return 'welcome';
        case 'configure':
          return 'connect';
        case 'execute':
          return 'configure';
        case 'router':
          return 'execute';
        case 'welcome':
        default:
          return prev;
      }
    });
  };

  const multiStepForm: FormStep[] = [
    {
      name: 'welcome',
      component: (
        <WelcomeOverview
          onNextStep={handleNextStep}
          setIsDialogOpen={setShowGuide}
          completeForm={completeForm}
        />
      ),
    },
    {
      name: 'connect',
      component: (
        <Connect
          onNextStep={handleNextStep}
          onPreviousStep={handlePreviousStep}
        />
      ),
    },
    {
      name: 'configure',
      component: (
        <Configure
          onNextStep={handleNextStep}
          onPreviousStep={handlePreviousStep}
        />
      ),
    },
    {
      name: 'execute',
      component: (
        <Sync onNextStep={handleNextStep} onPreviousStep={handlePreviousStep} />
      ),
    },
    {
      name: 'router',
      component: (
        <WelcomeRouter
          onPreviousStep={handlePreviousStep}
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
          hasCompletedOnboarding: true,
        },
      });
      queryclient.setQueryData(
        createConnectQueryKey({
          schema: getAccountOnboardingConfig,
          input: {
            accountId: account.id,
          },
          cardinality: undefined,
        }),
        create(GetAccountOnboardingConfigResponseSchema, {
          config: resp.config,
        })
      );
      setShowGuide(false);
    } catch (e) {
      toast.error('Unable to complete onboarding', {
        description: getErrorMessage(e),
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  useEffect(() => {
    if (!isLoading && !error && !isValidating && data) {
      setShowGuide(!data.config?.hasCompletedOnboarding);
    }
  }, [isLoading, isValidating, error, data]);

  if (!showGuide) {
    return <></>;
  }
  const stepOrder: FormStepName[] = multiStepForm.map((item) => {
    return item.name;
  });

  return (
    <AlertDialog open={showGuide} onOpenChange={setShowGuide}>
      <AlertDialogContent className="max-w-2xl">
        <AlertDialogTitle />
        <AlertDialogDescription />
        <div className="flex flex-col gap-8 pt-6">
          <StepProgress
            steps={multiStepForm.map((step) => step)}
            currentStep={currentStep}
            stepOrder={stepOrder}
          />
          <div className="px-8">
            {multiStepForm.find((step) => step.name === currentStep)?.component}
          </div>
        </div>
      </AlertDialogContent>
    </AlertDialog>
  );
}
