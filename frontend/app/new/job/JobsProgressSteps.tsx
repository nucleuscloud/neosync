import { Step } from '@/components/progress-steps/Step';
import { useStep } from '@/components/progress-steps/useStep';
import { ReactElement, useEffect } from 'react';

interface OnboardStep {
  name: string;
}

interface Props {
  stepName: string;
}

export default function JobsProgressSteps(props: Props): ReactElement {
  const { stepName } = props;

  //maxStep must match the steps.length below
  const [currentStep, { setStep }] = useStep({
    maxStep: 4,
    initialStep: 0,
  });

  const steps: OnboardStep[] = [
    {
      name: 'define',
    },
    {
      name: 'connect',
    },
    {
      name: 'schema',
    },
    {
      name: 'subset',
    },
  ];

  useEffect(() => {
    const ind = steps?.findIndex((item) => item.name == stepName);
    setStep(ind);
  }, [stepName]);

  return (
    <div>
      <div className="flex flex-row items-center justify-center mt-5">
        {steps.map((step, idx) => {
          return (
            <Step
              isCompleted={currentStep > idx}
              key={step.name}
              isActive={currentStep === idx}
              isLastStep={steps.length === idx + 1}
              name={step.name}
            />
          );
        })}
      </div>
    </div>
  );
}
