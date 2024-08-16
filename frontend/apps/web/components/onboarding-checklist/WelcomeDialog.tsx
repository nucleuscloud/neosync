'use client';
import { Dialog, DialogContent } from '@/components/ui/dialog';
import { ReactElement, useState } from 'react';
import { Button } from '../ui/button';
import StepProgress from './StepProgress';
import WelcomeOverview from './WelcomeOverview';
import WelcomeRouter from './WelcomeRouter';

export interface FormStep {
  name: string;
  component: JSX.Element;
}

export default function WelcomeDialog(): ReactElement {
  const [currentStep, setCurrentStep] = useState<number>(0);
  const [isDialogOpen, setIsDialogOpen] = useState<boolean>(true); //revert after testing

  const multiStepForm: FormStep[] = [
    {
      name: 'welcome',
      component: (
        <WelcomeOverview
          currentStep={currentStep}
          setCurrentStep={setCurrentStep}
          setIsDialogOpen={setIsDialogOpen}
        />
      ),
    },
    {
      name: 'router',
      component: (
        <WelcomeRouter
          currentStep={currentStep}
          setCurrentStep={setCurrentStep}
          setIsDialogOpen={setIsDialogOpen}
        />
      ),
    },
  ];

  return (
    <>
      <Button onClick={() => setIsDialogOpen(true)}>Open Dialog</Button>
      <Dialog
        open={isDialogOpen}
        onOpenChange={() => setIsDialogOpen(!isDialogOpen)}
      >
        <DialogContent className="max-w-2xl">
          <div className="flex flex-col gap-8 pt-6">
            <StepProgress
              steps={multiStepForm.map((step) => step)}
              currentStep={currentStep}
            />
            <div className="px-8">{multiStepForm[currentStep].component}</div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
