import Step from '@/components/progress-steps/Step';
import { ReactElement } from 'react';
import { NewJobType } from './job-form-validations';

const DATA_SYNC_STEPS_NO_SUBSET: JobProgressStep[] = [
  'define',
  'connect',
  'schema',
];
const DATA_SYNC_STEPS_WITH_SUBSET: JobProgressStep[] = [
  'define',
  'connect',
  'schema',
  'subset',
];

const DATA_GEN_STEPS: JobProgressStep[] = ['define', 'connect', 'schema'];

type JobProgressStep = 'define' | 'connect' | 'schema' | 'subset';

interface Props {
  steps: JobProgressStep[];
  stepName: JobProgressStep;
}

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

export function getJobProgressSteps(
  jobtype: 'data-sync',
  includeSubsetting: boolean
): JobProgressStep[];
export function getJobProgressSteps(
  jobtype: 'ai-generate-table' | 'generate-table'
): JobProgressStep[];
export function getJobProgressSteps(jobtype: NewJobType): JobProgressStep[];
export function getJobProgressSteps(
  jobtype: NewJobType,
  includeSubsetting?: boolean
): JobProgressStep[] {
  switch (jobtype) {
    case 'data-sync':
      return includeSubsetting
        ? DATA_SYNC_STEPS_WITH_SUBSET
        : DATA_SYNC_STEPS_NO_SUBSET;
    case 'generate-table':
      return DATA_GEN_STEPS;
    case 'ai-generate-table':
      return DATA_GEN_STEPS;
    default:
      return [];
  }
}
