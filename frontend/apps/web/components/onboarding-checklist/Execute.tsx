import { ArrowRightIcon } from '@radix-ui/react-icons';
import { useTheme } from 'next-themes';
import { ReactElement } from 'react';
import { Button } from '../ui/button';
import { HowItWorksImage } from './hiw';

interface Props {
  currentStep: number;
  setCurrentStep: (val: number) => void;
}

export default function Execute(props: Props): ReactElement {
  const { currentStep, setCurrentStep } = props;
  const theme = useTheme();

  return (
    <div className="flex flex-col gap-12 justify-center items-center text-center">
      <h1 className="font-semibold text-2xl">How it works</h1>
      {theme.resolvedTheme == 'light' ? (
        <HowItWorksImage />
      ) : (
        <HowItWorksImage />
      )}
      <p className="text-sm px-10">
        Now that you have an idea of Neosync works, click{' '}
        <span className="font-semibold">Next</span> to create a Job.
      </p>
      <div className="flex flex-row justify-between w-full py-6">
        <Button
          variant="outline"
          type="reset"
          onClick={() => setCurrentStep(currentStep - 1)}
        >
          Back
        </Button>
        <Button onClick={() => setCurrentStep(currentStep + 1)}>
          <div className="flex flex-row items-center gap-2">
            <div>Next</div> <ArrowRightIcon />
          </div>
        </Button>
      </div>
    </div>
  );
}
