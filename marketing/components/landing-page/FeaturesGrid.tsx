import { Button } from '@/components/ui/button';
import {
  ArrowRightIcon,
  CalendarIcon,
  LockClosedIcon,
} from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';
import { AiOutlineAudit } from 'react-icons/ai';
import { BiExpand } from 'react-icons/bi';
import { BsRecycle } from 'react-icons/bs';
import { PiLink } from 'react-icons/pi';

export default function FeaturesGrid(): ReactElement {
  const features = [
    {
      title: 'Complete referential integrity',
      description: `Neosync automatically handles and preserves your data's referential integrity.`,
      icon: <PiLink />,
    },
    {
      title: 'Retries',
      description:
        'Neosync handles automatically handles retries in the case of an error during the sync process.',
      icon: <BsRecycle />,
    },
    {
      title: 'Scheduling',
      description:
        'Run jobs ad-hoc, pause existing jobs or schedule jobs to run on any cadence you decide.',
      icon: <CalendarIcon />,
    },
    {
      title: 'Scalable',
      description:
        "Neosync's async pipeline is horizontally scalable making it a great choice for bigger data sets.",
      icon: <BiExpand />,
    },
    {
      title: 'Security',
      description:
        'Neosync comes with RBAC out of the box and is designed to be multi-tenant for teams deploying on-prem.',
      icon: <LockClosedIcon />,
    },
    {
      title: 'Audit',
      description:
        'Neosync ships with audit controls so security and compliance teams have a full audit trail over the system.',
      icon: <AiOutlineAudit />,
    },
  ];

  return (
    <div className="bg-[#F5F5F5] pt-20">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className=" flex flex-col space-y-20 bg-gradient-to-tr from-[#0F0F0F] to-[#262626] rounded-2xl py-12 px-6 shadow-lg">
          <div>
            <div className="text-gray-100 font-semibold text-2xl lg:text-5xl font-satoshi pt-10">
              Built for Security and Scalability
            </div>
            <div className="text-lg text-gray-400 font-satoshi font-light pt-8 lg:w-[70%]">
              Neosync is built for teams of all sizes and ships with features
              that put security and compliance teams of all sizes at ease.
              Whether you&apos;re an enterprise who needs an air-gapped
              deployment or a startup looking to get started quickly, Neosync
              can help.
            </div>
            <div className="pt-10">
              <Button variant="secondary" className="px-4">
                <Link href="https://docs.neosync.dev">
                  <div className="flex flex-row">
                    Documentation <ArrowRightIcon className="ml-2 h-5 w-5" />
                  </div>
                </Link>
              </Button>
            </div>
            <div className="text-lg text-gray-400 font-satoshi font-light py-20 grid grid-cols-1 lg:grid-cols-3 gap-6">
              {features.map((item) => (
                <div
                  key={item.title}
                  className="bg-[#282828] border border-[#484848] rounded-lg p-4 hover:-translate-y-2 duration-150 shadow-xl"
                >
                  <div className="flex flex-row items-center space-x-3">
                    <div className="text-gray-100">{item.icon}</div>
                    <div className="text-gray-100 text-xl">{item.title}</div>
                  </div>
                  <div className="text-sm">{item.description}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
