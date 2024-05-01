'use client';
import { ArrowTopRightIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import Link from 'next/link';
import posthog from 'posthog-js';
import { ReactElement } from 'react';

export default function DeploymentOptions(): ReactElement {
  const deployOptions = [
    {
      name: 'Open Source',
      description:
        'Deploy Neosync on your infrastructure using Helm or Docker Compose.',
      link: 'https://github.com/nucleuscloud/neosync',
      image: '/images/osdeploy.svg',
    },
    {
      name: 'Neosync Cloud',
      description:
        "Use Neosync's fully managed cloud to not worry about infrastructure and sign up now.",
      link: 'https://app.neosync.dev',
      image: '/images/nss.svg',
    },
    {
      name: 'Neosync Managed',
      description:
        'Deploy the Neosync data plane in your infrastructure and keep your data in your VPC.',
      link: 'https://calendly.com/evis1/30min',
      image: '/images/manag.svg',
    },
  ];

  return (
    <div className="px-6 gap-6">
      <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        Flexible Deployment Options
      </div>
      <div className="text-md text-gray-700 font-satoshi font-semibold pt-10 lg:px-60 text-center">
        Choose the deployment option that works best for you and your needs.
      </div>
      <div className="flex flex-col items-center lg:flex-row gap-4 pt-20 justify-center">
        {deployOptions.map((opt) => (
          <Link
            href={opt.link}
            target="_blank"
            key={opt.name}
            className="flex flex-col gap-2 border rounded-xl shadow-xl p-6 bg-white w-full  max-w-[300px] lg:max-w-[500px]h-[360px] hover:border-gray-400 items-center"
            onClick={() =>
              posthog.capture('user click', {
                page: opt.name,
              })
            }
          >
            <Image
              src={opt.image}
              alt="NeosyncLogo"
              width="178"
              height="150"
              className="w-[178px] h-[150px]"
            />
            <div className="flex flex-row items-center gap-4 pt-6">
              <div className="text-xl font-semibold text-al">{opt.name}</div>
              <ArrowTopRightIcon />
            </div>
            <div className="text-sm text-gray-500">{opt.description}</div>
          </Link>
        ))}
      </div>
    </div>
  );
}
