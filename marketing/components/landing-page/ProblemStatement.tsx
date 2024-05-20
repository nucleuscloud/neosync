'use client';
import { ReactElement } from 'react';

export default function ProblemStatement(): ReactElement {
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
      <div className="flex flex-row items-center lg:flex-row gap-4 pt-20 justify-center">
        <div className=" w-1/2">
          <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-left">
            Safely test your code against production data
          </div>
          <div className="text-md text-gray-700 font-satoshi font-semibold pt-10  text-left">
            Mock data is never representative of production data. Neosync allows
            you to anonymizes production data for the best local testing
            experience. Test against all of the edge cases before you to get
            production.
          </div>
          <div></div>
        </div>
        <div className=" w-1/2"></div>
      </div>
      <div className="flex flex-row items-center lg:flex-row gap-4 pt-20 justify-center">
        <div className=" w-1/2">
          <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-left">
            Reproduce production bugs locally
          </div>
          <div className="text-md text-gray-700 font-satoshi font-semibold pt-10  text-left">
            Seeing a data bug in production but can&apos;t reproduce it locally?
            Use Neosync to anonymize and subset your production database by a
            customer&apos;s ID to safely reproduce their state locally.
          </div>
          <div></div>
        </div>
        <div className=" w-1/2"></div>
      </div>
      <div className="flex flex-row items-center lg:flex-row gap-4 pt-20 justify-center">
        <div className=" w-1/2">
          <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-left">
            Automate your seeding scripts
          </div>
          <div className="text-md text-gray-700 font-satoshi font-semibold pt-10  text-left">
            Neosync is built to orchestrate, anonymize and automate data across
            environments. No more scripts that you need to manually update every
            time a schema changes.
          </div>
          <div></div>
        </div>
        <div className=" w-1/2"></div>
      </div>
      <div className="flex flex-row items-center lg:flex-row gap-4 pt-20 justify-center">
        <div className=" w-1/2">
          <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-left">
            Reduce your compliance scope
          </div>
          <div className="text-md text-gray-700 font-satoshi font-semibold pt-10  text-left">
            Neosync helps you isolate your sensitive data to a single
            environment and anonymizes it for everywhere else. This reduces your
            compliance scope and makes it easier to comply with GDPR, DPDP,
            HIPAA and other regulations.
          </div>
          <div></div>
        </div>
        <div className=" w-1/2"></div>
      </div>
    </div>
  );
}
