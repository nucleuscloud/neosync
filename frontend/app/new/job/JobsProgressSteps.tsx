import Step from '@/components/progress-steps/Step';
import { ReactElement } from 'react';

interface Props {
  steps: JobProgressStep[];
  stepName: JobProgressStep;
}

export const DATA_SYNC_STEPS: JobProgressStep[] = [
  'define',
  'connect',
  'schema',
  'subset',
];
export const DATA_GEN_STEPS: JobProgressStep[] = [
  'define',
  'connect',
  'schema',
];

type JobProgressStep = 'define' | 'connect' | 'schema' | 'subset';

export default function JobsProgressSteps(props: Props): ReactElement {
  const { steps, stepName } = props;

  const currentStepIndex = steps.indexOf(stepName);

  return (
    <div>
      <div className="flex flex-row items-center justify-center mt-5">
        {steps.map((step, index) => {
          const isCompleted = index < currentStepIndex;
          const isActive = step === stepName;
          return (
            <Step
              key={step}
              name={step}
              isCompleted={isCompleted}
              isActive={isActive}
              isLastStep={index === steps.length - 1}
            />
          );
        })}
      </div>
    </div>
  );
}
