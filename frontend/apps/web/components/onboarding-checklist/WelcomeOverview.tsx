import { ArrowRightIcon } from '@radix-ui/react-icons';
import { useTheme } from 'next-themes';
import { ReactElement } from 'react';
import { Button } from '../ui/button';
import { Welcome } from './Welcome';
import { WelcomeDarkMode } from './WelcomeDarkMode';

interface Props {
  currentStep: number;
  setCurrentStep: (val: number) => void;
  setIsDialogOpen: (val: boolean) => void;
}

export default function WelcomeOverview(props: Props): ReactElement {
  const { currentStep, setCurrentStep, setIsDialogOpen } = props;
  const theme = useTheme();

  return (
    <div className="flex flex-col gap-12 justify-center items-center text-center">
      <h1 className="font-semibold text-2xl">Welcome to Neosync</h1>
      {theme.theme == 'light' ? <Welcome /> : <WelcomeDarkMode />}
      <p className="text-sm px-10">
        Neosync makes it easy to anonymize sensitive data, generate synthetic
        data and sync data across environments. Click{' '}
        <span className="font-semibold">Next</span> to get started.
      </p>
      <div className="flex flex-row justify-between w-full mt-4">
        <Button
          type="button"
          variant="outline"
          onClick={() => setIsDialogOpen(false)}
        >
          Skip
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
