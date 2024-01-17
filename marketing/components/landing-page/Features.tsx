import { LinkBreak1Icon, Share1Icon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { GoWorkflow } from 'react-icons/go';
import { PiFlaskLight } from 'react-icons/pi';

export default function Features(): ReactElement {
  const features = [
    {
      title: 'Orchestration',
      description:
        'Neosync orchestrates data across environments and automatically handles scheduling, retries and executions.',
      icon: <Share1Icon className="w-14 h-14" />,
    },
    {
      title: 'Synthetic Data',
      description:
        'Neosync can generate synthetic data that is structurally and statistically consistent to your source data set.',
      icon: <PiFlaskLight className="w-14 h-14" />,
    },
    {
      title: 'Anonymization',
      description:
        'Neosync ships with 40+ transformers that allow you to anonymize and transform sensitive data.',
      icon: <LinkBreak1Icon className="w-14 h-14" />,
    },
    {
      title: 'Referential integrity',
      description: `Neosync automatically handles referential integrity whether you have 1 table or 1000 tables.`,
      icon: <GoWorkflow className="w-14 h-14" />,
    },
  ];

  return (
    <div className="bg-[#F5F5F5]">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className=" flex flex-col px-6">
          <div className="text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi text-center">
            Built for Security and Scalability
          </div>
          <div className="text-lg text-gray-600 font-satoshi font-semibold pt-8 px-40 text-center">
            Neosync is built for teams and data of all sizes. Whether
            you&apos;re an enterprise who needs an air-gapped deployment or a
            startup looking to get started quickly, Neosync can help.
          </div>
        </div>
        <div className="text-lg text-gray-400 font-satoshi font-light py-20 flex flex-row gap-6">
          {features.map((item) => (
            <div
              key={item.title}
              className="bg-gradient-to-tr from-[#0F0F0F] to-[#2e2e2e] rounded-xl p-8 shadow-xl items-center flex flex-col gap-6 text-center w-[480px]"
            >
              <div className="text-gray-100">{item.icon}</div>
              <div className="text-gray-100 text-2xl">{item.title}</div>
              <div className="text-[16px] ">{item.description}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
