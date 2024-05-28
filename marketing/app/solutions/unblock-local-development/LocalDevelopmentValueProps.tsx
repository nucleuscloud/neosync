import Image from 'next/image';
import { ReactElement } from 'react';

export default function LocalDevelopmentValueProps(): ReactElement {
  const features = [
    {
      title: 'Reproduce production errors locally',
      description:
        'Subset or filter anonymized production data to easily and quickly identify or reproduce errors.',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/errorsNew.svg',
    },
    {
      title: 'Enable developers to self-service',
      description:
        'Developers can self-serve de-identified or synthetic data whenever they need to without waiting on other teams. ',
      image: '/images/unblocklocal.svg',
    },
    {
      title: 'Hydrate local databases',
      description:
        'Automate hydrating local databases with the latest anonymized production data. Use the Neosync CLI to run ad-hoc jobs. ',
      image: '/images/envsync3.svg',
    },
    {
      title: `Frictionless security, privacy and compliance`,
      description: `Easily comply with laws like HIPAA, GDPR, and DPDP with de-identified and synthetic data that structurally and statistically looks just like your production data.`,
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/compliance.svg',
    },
  ];

  return (
    <div>
      <div className="px-6">
        <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
          Empower your Developers to Self-Serve and Build Faster than Ever
        </div>
      </div>
      <div className="text-lg text-gray-400 font-satoshi font-light  pt-10 lg:pt-20 flex flex-col lg:flex-row gap-6 justify-center items-center">
        {features.map((item) => (
          <div
            key={item.title}
            className="border border-gray-300 bg-white rounded-xl p-2 lg:p-4 xl:p-8 shadow-xl flex flex-col justify-between gap-6 text-center w-full max-w-[300px] lg:w-[270px] lg:max-w-[400px] mx-auto sm:h-[100px] md:h-[360px] hover:border-gray-400"
          >
            <div>
              <div className="flex justify-center">
                <Image
                  src={item.image}
                  alt="NeosyncLogo"
                  width="178"
                  height="131"
                  className="w-[178px] h-[131px]"
                />
              </div>
              <div className="text-gray-900 text-2xl pt-6">{item.title}</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
