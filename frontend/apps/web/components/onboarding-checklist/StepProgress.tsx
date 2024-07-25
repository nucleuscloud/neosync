import { cn } from '@/libs/utils';
import { ReactElement } from 'react';
import { FormStep } from './WelcomeDialog';

interface Props {
  steps: FormStep[];
  currentStep: number;
}

export default function StepProgress(props: Props): ReactElement {
  const { steps, currentStep } = props;

  return (
    <div className="flex flex-row items-center gap-1 justify-center">
      {steps.map((step, index) => (
        <div
          key={step.name}
          className={cn(
            'rounded-2xl w-2 h-2',
            index < currentStep || index === currentStep
              ? 'bg-gray-900 dark:bg-blue-600'
              : 'bg-gray-300 dark:bg-gray-700' // Upcoming steps
          )}
        ></div>
      ))}
    </div>
  );
}
