'use client';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Progress } from '@/components/ui/progress';
import {
  CheckCircledIcon,
  CircleIcon,
  RocketIcon,
} from '@radix-ui/react-icons';
import { useState } from 'react';
import { Button } from '../ui/button';

const steps = [
  {
    id: '1',
    title: 'Create a source connection',
    description: 'Create your source connection',
  },
  {
    id: '2',
    title: 'Create a destination connection',
    description: 'Create your destination connection',
  },
  {
    id: '3',
    title: 'Create a job',
    description: 'Create your first job',
  },
  {
    id: '4',
    title: 'Invite teammates',
    description: 'Invite your teammates to collaborate',
  },
];

export default function OnboardingChecklist() {
  const [progress, setProgress] = useState(13);
  const [complete, isComplete] = useState(false);

  return (
    <div className="fixed right-[120px] bottom-[20px] z-50 ">
      <Popover>
        <PopoverTrigger className="border border-gray-400 rounded-lg p-2">
          Open Guide
        </PopoverTrigger>
        <PopoverContent>
          <div className="flex flex-col gap-6">
            <div className="flex flex-col gap-2">
              <div className="flex flex-row gap-2 items-center">
                <div>
                  <RocketIcon />
                </div>
                <div className="font-semibold">Welcome to Neosync!</div>
              </div>
              <div className="text-sm">
                Get started by completing these steps
              </div>
            </div>
            <div className="flex flex-row gap-2 items-center">
              <Progress value={progress} />
              <div>{progress}</div>
            </div>
            <div className="flex flex-col gap-2">
              {steps.map((step, index) => (
                <div className="flex flex-col justify-left ">
                  <div
                    key={step.id}
                    className="flex flex-row items-center gap-2"
                  >
                    <div>
                      {index == -1 ? (
                        <CheckCircledIcon className="w-4 h-4" />
                      ) : (
                        <CircleIcon />
                      )}
                    </div>
                    <div className="text-md">{step.title}</div>
                  </div>
                  <div className="text-xs pl-6">{step.description}</div>
                </div>
              ))}
            </div>
            <Button variant="outline">Do not show again</Button>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
