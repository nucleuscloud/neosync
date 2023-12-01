import { Step } from '@/components/progress-steps/Step';
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

export type JobProgressStep = 'define' | 'connect' | 'schema' | 'subset';

export default function JobsProgressSteps(props: Props): ReactElement {
  const { steps, stepName } = props;

  const lastStep = steps[steps.length - 1];
  const isCompleted = stepName === lastStep;

  return (
    <div>
      <div className="flex flex-row items-center justify-center mt-5">
        {steps.map((step) => {
          return (
            <Step
              key={step}
              name={step}
              isCompleted={isCompleted}
              isActive={step === stepName}
              isLastStep={step === lastStep}
            />
          );
        })}
      </div>
    </div>
  );
}
