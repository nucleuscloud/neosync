import Image from 'next/image';
import { ReactElement } from 'react';

export default function Personas(): ReactElement {
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

  const tableTypes = [
    {
      name: 'Tabular',
      description:
        'Generate statistically consistent synthetic data for a single data frame or table.',
      image: (
        <Image
          src="https://assets.nucleuscloud.com/neosync/marketingsite/tabularcropped.png"
          alt="pre"
          width="436"
          height="416"
          className="w-[240px] h-auto"
        />
      ),
    },
    {
      name: 'Relational',
      description:
        'Generate statistically consistent synthetic data for a relational database while maintaining referential integrity.',
      image: (
        <Image
          src="https://assets.nucleuscloud.com/neosync/marketingsite/relationaltablesgrid.png"
          alt="pre"
          width="436"
          height="416"
          className="w-[600px] h-auto"
        />
      ),
    },
  ];

  return (
    <div className="bg-[#F5F5F5]">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className="text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi text-center">
          Built for Engineering Teams
        </div>
        <div className="text-lg text-gray-600 font-satoshi font-semibold pt-10 px-60 text-center">
          From Local, to Stage to CI, Neosync has APIs, SDKs and a CLI to fit
          into every workflow.
        </div>
        <div className="border border-gray-400 bg-white p-6 rounded-xl shadow-lg mt-10 mx-80">
          <div className="rounded-xl  flex justify-center">
            <Image
              src="https://assets.nucleuscloud.com/neosync/marketingsite/cicd.png"
              alt="pre"
              width="800"
              height="542"
            />
          </div>
        </div>
      </div>
    </div>
  );
}
