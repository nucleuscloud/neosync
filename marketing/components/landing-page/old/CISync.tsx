import { CheckCircledIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import { ReactElement } from 'react';
import { Badge } from '../../ui/badge';

export default function CISync(): ReactElement {
  const citesting = [
    {
      name: 'gitops',
      description:
        'Automatically hydrate your CI database with safe, production-like data ',
    },
    {
      name: 'yaml',
      description:
        'Declaratively define a step in your CI pipeline with Neosync',
    },
    {
      name: 'risk',
      description: 'Protect sensitive data from your CI pipelines',
    },
    {
      name: 'subset',
      description: 'Subset your database to reduce your ',
    },
  ];

  return (
    <div className="bg-[#F5F5F5]">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className="border border-gray-400 bg-white p-6 rounded-xl shadow-lg">
          <div>
            <Badge
              variant="secondary"
              className="text-md border border-gray-700"
            >
              Developers
            </Badge>
          </div>
          <div className="flex flex-col-reverse lg:flex-row gap-8 pt-6 ">
            <div className="lg:w-[60%] rounded-lg">
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/cicode.svg"
                alt="pre"
                className="w-full"
                width="549"
                height="554"
              />
            </div>
            <div className="flex flex-col space-y-8 lg:w-[60%]">
              <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi z-10">
                Hydrate your CI databases with synthetic data on every PR
              </div>
              <div className="flex flex-col justify-start">
                {citesting.map((item) => (
                  <div
                    className="flex flex-row space-x-4 items-center pt-6"
                    key={item.name}
                  >
                    <CheckCircledIcon className="w-6 h-6" />{' '}
                    <div className="text-sm lg:text-lg text-gray-700  font-satoshi font-semibold">
                      {item.description}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
