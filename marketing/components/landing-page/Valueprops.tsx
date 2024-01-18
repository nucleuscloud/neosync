import Image from 'next/image';
import { ReactElement } from 'react';

export default function ValueProps(): ReactElement {
  const features = [
    {
      title: 'Unblock local development ',
      description:
        'Shift left and give developers the ability to self-serve de-identified and synthetic data locally whenever they need it. ',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/fixlocal.png',
    },
    {
      title: 'Fix broken staging environments',
      description:
        'Catch production bugs and ship faster when you hydrate your staging and QA environments with production-like data that is safe and fast to generate. ',
      image: 'https://assets.nucleuscloud.com/neosync/marketingsite/stage.png',
    },
    {
      title: 'Keep environments up to date',
      description:
        'Speed up your dev and test cycles. Make sure your environments stay in sync with the latest de-identified and synthetic data that you can refresh whenever you need to.',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/environmentsync.png',
    },
    {
      title: `Frictionless security, privacy and compliance`,
      description: `Easily and quickly abide by laws like HIPAA, GDPR, and DPDP with de-identified and synthetic data that structurally and statistically looks just like your production data.  `,
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/security.png',
    },
  ];

  return (
    <div className="bg-[#F5F5F5]">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className=" flex flex-col px-6">
          <div className="text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi text-center">
            Protect Your Sensitive Data, Empower Your Developers
          </div>
          <div className="text-lg text-gray-600 font-satoshi font-semibold pt-10 px-40 text-center">
            Neosync is for teams of all sizes. Whether you&apos;re an enterprise
            who needs an air-gapped deployment or a startup looking to get
            started quickly, Neosync can help.
          </div>
        </div>
        <div className="text-lg text-gray-400 font-satoshi font-light pt-20 flex flex-row gap-6">
          {features.map((item) => (
            <div
              key={item.title}
              className="border border-gray-400 bg-white rounded-xl p-8 shadow-xl items-center flex flex-col gap-6 text-center w-[480px]"
            >
              <div className="text-gray-900">
                {' '}
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
    </div>
  );
}
