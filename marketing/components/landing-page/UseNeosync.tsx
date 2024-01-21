import { ConnectDash } from '@/public/images/Connect';
import { SyncDash } from '@/public/images/Sync';
import { ArrowRightIcon } from 'lucide-react';
import Link from 'next/link';
import { ReactElement } from 'react';
import { GoCheckCircleFill } from 'react-icons/go';
import { PiBookOpenText } from 'react-icons/pi';
import { Button } from '../ui/button';

export default function UseNeosync(): ReactElement {
  const steps = [
    {
      step: '1',
      title: 'Deploy',
      description: `Start Neosync locally using Tilt or Docker compose. When
you're ready, deploy Neosync using a Helm Chart or Docker
    Compose.`,
      image: 'https://assets.nucleuscloud.com/neosync/marketingsite/deploy.svg',
    },
    {
      step: '2',
      title: 'Connect',
      description: `Connect your source and destinations. Neosync supports Postgres, Mysql, S3 and we're always building more integrations.`,
      image: <ConnectDash />,
    },
    {
      step: '3',
      title: 'Configure',
      description: `Configure your schemas, tables and columns with transformers that de-identify your data or generate synthetic data. Neosync automatically handles all relational integrity. `,
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/configurefinal.svg',
    },
    {
      step: '4',
      title: 'Sync',
      description: `Sync data across systems or generate synthetic data from scratch and send it a downstream system. `,
      // image:
      //   'https://assets.nucleuscloud.com/neosync/marketingsite/syncfinal.svg',
      image: <SyncDash />,
    },
  ];
  return (
    <div>
      <div className="px-6">
        <div className="text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi text-center">
          Get Up and Running in Minutes
        </div>
        <div className="text-lg text-gray-600 font-satoshi font-semibold pt-10 lg:px-60 text-center">
          Whether you want to run Neosync locally, on VMs or in Kubernetes,
          Neosync is easy to deploy using Docker or Helm.
        </div>
        <div className="justify-center flex pt-10">
          <Button className="px-4">
            <Link href="https://docs.neosync.dev">
              <div className="flex flex-row items-center gap-2">
                <PiBookOpenText className="h-5 w-5" />
                Documentation <ArrowRightIcon className="h-5 w-5" />
              </div>
            </Link>
          </Button>
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
                  {index == 3 && (
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
              <div className="my-8">
                {/* <Image src={step.image} alt="pre" width="800" height="317" /> */}
                {step.image}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
