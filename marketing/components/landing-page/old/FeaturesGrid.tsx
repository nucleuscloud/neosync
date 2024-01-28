import { Button } from '@/components/ui/button';
import {
  ArrowRightIcon,
  CalendarIcon,
  LinkBreak1Icon,
} from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';
import { BiExpand } from 'react-icons/bi';
import { BsRecycle } from 'react-icons/bs';
import { MdScience } from 'react-icons/md';
import { PiLink } from 'react-icons/pi';

export default function FeaturesGrid(): ReactElement {
  const features = [
    {
      title: 'Synthetic Data Generation',
      description:
        'Neosync can generate synthetic data that is structurally and nearly statistically identical to your sensitive data set',
      icon: <MdScience />,
    },
    {
      title: 'Anonymization',
      description:
        'Neosync ships with 40+ transformers that allow you to anonymize and transform sensitive data.',
      icon: <LinkBreak1Icon />,
    },
    {
      title: 'Referential integrity',
      description: `Neosync automatically handles and preserves your data's referential integrity.`,
      icon: <PiLink />,
    },
    {
      title: 'Subsetting',
      description:
        'Create subsets of your database to easily reproduce bugs or fit your database locally. Neosync maintains full referential integrity.',
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
  ];

  return (
    <div className="bg-[#F5F5F5] pt-20">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className=" flex flex-col space-y-20 bg-gradient-to-tr from-[#0F0F0F] to-[#262626] rounded-2xl py-12 px-6 shadow-lg">
          <div>
            <div className="text-gray-100 font-semibold text-2xl lg:text-4xl font-satoshi pt-10">
              Built for Security and Scalability
            </div>
            <div className="text-lg text-gray-400 font-satoshi font-light pt-8 lg:w-[70%]">
              Neosync is built for teams and data of all sizes. Whether
              you&apos;re an enterprise who needs an air-gapped deployment or a
              startup looking to get started quickly, Neosync can help.
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
          </div>
        </div>
        <div className="text-lg text-gray-400 font-satoshi font-light py-20 grid grid-cols-1 lg:grid-cols-3 gap-10">
          {features.map((item) => (
            <div
              key={item.title}
              className="bg-[#282828] border border-[#484848] rounded-lg p-6 hover:-translate-y-2 duration-150 shadow-xl"
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
  );
}
