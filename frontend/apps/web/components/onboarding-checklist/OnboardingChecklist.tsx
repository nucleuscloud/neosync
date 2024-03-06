'use client';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Progress } from '@/components/ui/progress';
import {
  ArrowRightIcon,
  CheckCircledIcon,
  CircleIcon,
  RocketIcon,
} from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { useAccount } from '../providers/account-provider';
import { Button } from '../ui/button';
import { Separator } from '../ui/separator';

const steps = [
  {
    id: '1',
    title: 'Create a source connection',
    href: '/connection',
  },
  {
    id: '2',
    title: 'Create a destination connection',
    href: '/connection',
  },
  {
    id: '3',
    title: 'Create a job',
    href: '/job',
  },
  {
    id: '4',
    title: 'Invite teammates',
    href: '/settings',
  },
];

export default function OnboardingChecklist() {
  const { account } = useAccount();
  const [progress, setProgress] = useState(13);
  const [complete, isComplete] = useState(false);
  const [isOpen, setIsOpen] = useState(false);

  const router = useRouter();

  return (
    <div className="fixed right-[160px] bottom-[20px] z-50 ">
      <Popover onOpenChange={() => setIsOpen(isOpen ? false : true)}>
        <PopoverTrigger className="border border-gray-400 rounded-lg p-2">
          {isOpen ? 'Close Guide' : 'Open Guide'}
        </PopoverTrigger>
        <PopoverContent className="w-[400px]">
          <div className="flex flex-col gap-4 p-2">
            <div className="flex flex-col gap-2">
              <div className="flex flex-row gap-2 items-center">
                <div>
                  <RocketIcon />
                </div>
                <div className="font-semibold">Welcome to Neosync!</div>
              </div>
              <div className="text-sm pl-6">
                Get started by completing these steps
              </div>
            </div>
            <div className="flex flex-row gap-2 items-center">
              <Progress value={progress} />
              <div className="text-sm">{progress}%</div>
            </div>
            <Separator />
            <div className="flex flex-col gap-2">
              {steps.map((step, index) => (
                <div className="flex flex-row items-center justify-between">
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
                    <div className="text-sm">{step.title}</div>
                  </div>
                  <Button
                    variant="ghost"
                    onClick={() =>
                      router.push(`/${account?.name}/new/${step.href}/`)
                    }
                  >
                    <ArrowRightIcon className="w-4 h-4" />
                  </Button>
                </div>
              ))}
            </div>
            <Separator />
            <div className=" flex flex-row items-center justify-between pt-6">
              <Button variant="outline">Don't show again</Button>
              <Button variant="default">Complete</Button>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
