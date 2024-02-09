'use client';
import { ArrowRight } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { posthog } from 'posthog-js';
import { ReactElement } from 'react';

export default function ValueProps(): ReactElement {
  const features = [
    {
      title: 'Unblock local development ',
      description:
        'Give developers the ability to self-serve de-identified and synthetic data locally whenever they need it without having to worry about sensitive data privacy or security. ',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/localdev.svg',
      link: '/solutions/unblock-local-development',
    },
    {
      title: 'Fix broken staging environments',
      description:
        'Catch production bugs and ship faster when you hydrate your staging and QA environments with production-like data that is safe and fast to generate. ',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/stagingsvg.svg',
      link: '/solutions/fix-staging-environments',
    },
    {
      title: 'Keep environments in sync',
      description:
        'Speed up your dev and test cycles. Make sure your environments stay in sync with the latest de-identified and synthetic data that you can refresh whenever you need to.',
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/syncenv.svg',
      link: '/solutions/keep-environments-in-sync',
    },
    {
      title: `Frictionless security, privacy and compliance`,
      description: `Easily comply with laws like HIPAA, GDPR, and DPDP with de-identified and synthetic data that structurally and statistically looks just like your production data.`,
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/compliance.svg',
      link: '/solutions/security-privacy',
    },
  ];

  const router = useRouter();

  return (
    <div>
      <div className="px-6">
        <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
          Protect Your Sensitive Data, Build Faster with Neosync
        </div>
      </div>
      <div className="text-lg text-gray-400 font-satoshi font-light  pt-10 lg:pt-20 flex flex-col lg:flex-row gap-6 justify-center items-center">
        {features.map((item) => (
          <div
            key={item.title}
            className="border border-gray-400 bg-white rounded-xl p-8 shadow-xl flex flex-col justify-between gap-6 text-center w-full lg:w-[480px] max-w-xs mx-auto lg:h-[560px] hover:shadow-gray-400"
          >
            <div>
              <div className="text-gray-900 ">
                <Image
                  src={item.image}
                  alt="NeosyncLogo"
                  width="250"
                  height="172"
                />
              </div>
              <div className="text-gray-900 text-2xl pt-10">{item.title}</div>
              <div className=" text-gray-500 text-[16px] pt-6 text-left">
                {item.description}
              </div>
            </div>
            <div>
              <Link
                href={item.link}
                target="_blank"
                className="flex flex-row justify-end text-sm items-center gap-2"
                onClick={() => {
                  posthog.capture('user click', {
                    page: item.title,
                  });
                }}
              >
                <div className="text-gray-900">Learn more</div>
                <div>
                  <ArrowRight className="text-gray-900 w-4 h-4" />
                </div>
              </Link>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
