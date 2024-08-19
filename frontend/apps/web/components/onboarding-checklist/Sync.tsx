import { ArrowRightIcon, CheckIcon } from '@radix-ui/react-icons';
import { useTheme } from 'next-themes';
import { ReactElement } from 'react';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Input } from '../ui/input';

interface Props {
  currentStep: number;
  setCurrentStep: (val: number) => void;
}

export default function Sync(props: Props): ReactElement {
  const { currentStep, setCurrentStep } = props;
  const theme = useTheme();

  return (
    <div className="flex flex-col gap-12 justify-center items-center text-center">
      <h1 className="font-semibold text-2xl">Sync</h1>
      <p className="text-sm px-10">
        Run your job on a set schedule or run it ad-hoc whenever you'd like. Now
        that you have an idea of Neosync works, click{' '}
        <span className="font-semibold">Next</span> to create a Job.
      </p>
      {/* {theme.resolvedTheme == 'light' ? (
        <ConnectLightMode />
      ) : (
        <ConnectDarkMode />
      )} */}
      <div className=" flex flex-col gap-8">
        <div className=" flex flex-col gap-2 text-left">
          <p className="text-sm font-semibold">Schedule</p>
          <Input
            disabled={true}
            value="0 0 1 1 *"
            className="shadow-lg border border-gray-400 dark:dark:border-[#0D47F0] w-[200px]"
          />
        </div>
        <div className=" flex flex-col gap-2 text-left">
          <p className="text-sm font-semibold">Job Runs</p>
          <div className=" flex flex-col gap-4 p-4 border border-gray-300 dark:dark:border-[#0D47F0] rounded-lg text-xs shadow-lg">
            <div className="flex flex-row gap-4 justify-between" id="header">
              <div className="w-[100px] font-semibold">Job</div>
              <div className="w-[100px] font-semibold">Start</div>
              <div className="w-[100px] font-semibold">Complete</div>
              <div className="w-[100px] font-semibold">Status</div>
            </div>
            <div className="flex flex-row items-center gap-4 justify-between">
              <div className="w-[100px]">sync-prod</div>
              <div className="w-[100px]">08/14/2024 12:15:54 PM</div>
              <div className="w-[100px]">08/14/2024 12:16:00 PM</div>
              <Badge variant="success" className="w-[100px]">
                <CheckIcon className="w-3 h-3" /> <p>Complete</p>
              </Badge>
            </div>
          </div>
        </div>
      </div>
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
