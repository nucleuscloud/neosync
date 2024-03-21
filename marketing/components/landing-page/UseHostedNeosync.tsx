import { ConfigureDash } from '@/public/images/Configure';
import { ConnectDash } from '@/public/images/Connect';
import { SyncDash } from '@/public/images/Sync';
import { ReactElement } from 'react';
import { GoCheckCircleFill } from 'react-icons/go';

export default function UseHostedNeosync(): ReactElement {
  const steps = [
    {
      step: '1',
      title: 'Connect',
      description: `Connect your source and destinations databases. Neosync supports any Postgres and Mysql compatible database as well as S3.`,
      image: <ConnectDash />,
    },
    {
      step: '2',
      title: 'Configure',
      description: `Configure your schemas, tables and columns with transformers that de-identify your data or generate synthetic data. Neosync automatically handles all relational integrity. `,
      image: <ConfigureDash />,
    },
    {
      step: '3',
      title: 'Sync',
      description: `Sync data across systems or generate synthetic data from scratch and send it a downstream system. `,
      image: <SyncDash />,
    },
  ];

  return (
    <div className="px-6">
      <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        Get Up and Running in Minutes
      </div>
      <div className="text-md text-gray-700 font-satoshi font-semibold pt-10 lg:px-60 text-center">
        It shouldn&apos;t take weeks to get up and running. With Neosync, you
        can configure your connections and set up your Jobs in just a few
        minutes.
      </div>
      <div className="pt-20">
        {steps.map((step, index) => (
          <div className="flex flex-col lg:flex-row" key={step.title}>
            <div className="flex flex-row gap-2 lg:gap-10">
              <div className="flex flex-col items-center ">
                <div className="w-8 h-9 bg-black rounded-full flex items-center justify-center text-white text-xl">
                  {step.step}
                </div>
                <div className="h-full w-[2px] bg-gray-900" />
                {index == 2 && (
                  <div>
                    <GoCheckCircleFill className="h-8 w-8 text-green-700" />
                  </div>
                )}
              </div>
              <div className="flex flex-col  gap-2 lg:gap-6 justify-start pr-4">
                <div className="text-gray-900 text-2xl"> {step.title}</div>
                <div className="lg:w-[400px]">{step.description}</div>
              </div>
            </div>
            <div className="my-8 lg:w-full">{step.image}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
