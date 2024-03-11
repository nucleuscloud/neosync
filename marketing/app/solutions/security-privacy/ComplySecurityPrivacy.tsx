import Image from 'next/image';
import { ReactElement } from 'react';

export default function ComplySecurityPrivacy(): ReactElement {
  const features = [
    {
      title: 'Anonymize sensitive data',
      description:
        'Anonymize sensitive data in order to make it safe for use in lower level environments and to share across teams.',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/anonymize.svg',
    },
    {
      title: 'Easily generate synthetic data',
      description:
        'Create synthetic data that statistically and structurally looks like your production data for use in lower level environments. ',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/synthetic.svg',
    },
    {
      title: 'Enable developers to self-service',
      description:
        'Developers can self-serve de-identified or synthetic data whenever they need to without waiting on other teams. ',
      image: '/images/unblocklocal.svg',
    },
    {
      title: `Comply with HIPAA, GDPR and more`,
      description: `Easily comply with laws like HIPAA, GDPR, and DPDP with de-identified and synthetic data that structurally and statistically looks just like your production data.`,
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/compliance.svg',
    },
  ];

  return (
    <div>
      <div className="px-6">
        <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
          Don&apos;t let Security, Privacy and Compliance Slow You Down
        </div>
      </div>
      <div className="text-lg text-gray-400 font-satoshi font-light pt-10 lg:pt-20 flex flex-col lg:flex-row gap-6 justify-center items-center">
        {features.map((item) => (
          <div
            key={item.title}
            className="border border-gray-300 bg-white rounded-xl p-2 lg:p-4 xl:p-8 shadow-xl flex flex-col justify-between gap-6 text-center w-full  max-w-[300px] lg:max-w-[400px] mx-auto h-[360px] hover:border-gray-400"
          >
            <div className="text-gray-900 ">
              <Image
                src={item.image}
                alt="NeosyncLogo"
                width="178"
                height="131"
                className="lg:max-w-[250px] lg:min-h-[172px] lg:max-h-[172px] sm:max-w-[200px]"
              />
            </div>
            <div className="text-gray-900 text-2xl">{item.title}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
