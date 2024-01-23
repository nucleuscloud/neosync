import Image from 'next/image';
import { ReactElement } from 'react';

export default function ValueProps(): ReactElement {
  const features = [
    {
      title: 'Unblock local development ',
      description:
        'Shift left and give developers the ability to self-serve de-identified and synthetic data locally whenever they need it without having to worry about sensitive data privacy or security. ',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/localdev.svg',
    },
    {
      title: 'Fix broken staging environments',
      description:
        'Catch production bugs and ship faster when you hydrate your staging and QA environments with production-like data that is safe and fast to generate. ',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/stagingsvg.svg',
    },
    {
      title: 'Keep environments up to date',
      description:
        'Speed up your dev and test cycles. Make sure your environments stay in sync with the latest de-identified and synthetic data that you can refresh whenever you need to.',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/syncenv.svg',
    },
    {
      title: `Frictionless security, privacy and compliance`,
      description: `Easily and quickly abide by laws like HIPAA, GDPR, and DPDP with de-identified and synthetic data that structurally and statistically looks just like your production data.  `,
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/compliance.svg',
    },
  ];

  return (
    <div>
      <div className="px-6">
        <div className="text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi text-center">
          Protect Your Sensitive Data, Build Faster with Neosync
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
