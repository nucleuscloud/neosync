import Image from 'next/image';
import { ReactElement } from 'react';

export default function KeepEnvironmentsInSync(): ReactElement {
  const features = [
    {
      title: 'No more environment mismatches',
      description:
        'Keep dev, staging and CI all in sync with the latest data. No more "well, it works locally!"',
      image: '/images/brokenenv.svg',
    },
    {
      title: 'Enable developers to self-service',
      description:
        'Developers can self-serve de-identified or synthetic data whenever they need to without waiting on other teams. ',
      image: '/images/unblocklocal.svg',
    },
    {
      title: 'Run syncs ad-hoc or on a schedule',
      description:
        'Configure your data syncs to run on a schedule or trigger then ad-hoc to get the latest data. ',
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
          Never Work with Stale Data Again
        </div>
      </div>
      <div className="text-lg text-gray-400 font-satoshi font-light  pt-10 lg:pt-20 flex flex-col lg:flex-row gap-6 justify-center items-center">
        {features.map((item) => (
          <div
            key={item.title}
            className="border border-gray-400 bg-white rounded-xl p-8 shadow-xl flex flex-col gap-6 text-center w-full lg:w-[480px] max-w-xs mx-auto lg:h-[520px]"
          >
            <div className="text-gray-900 ">
              <Image
                src={item.image}
                alt="NeosyncLogo"
                width="250"
                height="172"
                className="lg:max-w-[250px] lg:max-h-[172px] sm:max-w-[200px]"
              />
            </div>
            <div className="text-gray-900 text-2xl">{item.title}</div>
            <div className=" text-gray-500 text-[16px] ">
              {item.description}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
