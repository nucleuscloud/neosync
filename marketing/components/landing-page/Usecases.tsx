'use client';
import ApiCodeBlock from '@/components/codeblocks/ApiCode';
import { CheckCircledIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import { ReactElement } from 'react';

export default function UseCases(): ReactElement {
  const localTesting = [
    {
      name: 'sync',
      description:
        'APIs to automatically sync local databases with prod to keep them up to date. ',
    },
    {
      name: 'local',
      description:
        'Locally develop against safe,anonymized, production-like data',
    },
    {
      name: 'subset',
      description:
        'Subset your data using a custom query to shrink it to fit locally.',
    },
    {
      name: 'cli',
      description:
        'Use the Neosync CLI to back up to an old copy or pull the latest',
    },
  ];

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

  const apiCode = `...
  schedule = "0 23 * * *"
  haltOnNewColAdd = True
  jobRes, err = jobclient.CreateJob(ctx, connect.NewRequest({
      'AccountId': accountId,
      'JobName': 'prod-to-stage',
      'ConnectionSourceId': prodDbResp['Msg']['Connection']['Id'],
      'DestinationSourceIds': [
          stageDbResp['Msg']['Connection']['Id'],
          s3Resp['Msg']['Connection']['Id'],
      ],
      'CronSchedule': schedule,
      'HaltOnNewColumnAddition': haltOnNewColAdd,
      'Mappings': [
          {
              'Schema': 'public',
              'Table': 'users',
              'Column': 'account_number',
              'Transformer': JobMappingTransformer.custom_account_number,
          },
          {
              'Schema': 'public',
              'Table': 'users',
              'Column': 'address',
              'Transformer': JobMappingTransformer.address_anonymize,
          },
      ],
  }))
  if err:
      raise Exception(err)
...`;

  return (
    <div className="bg-[#F5F5F5] pt-20">
      <div className="flex flex-col pt-5 lg:pt-40  bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className="text-gray-900 font-semibold  text-2xl lg:text-5xl font-satoshi z-10  ">
          Synthetic Data Orchestration Built for Developers
        </div>
        <div className="flex flex-col pt-10 z-10 items-center space-y-10">
          <div className="flex flex-col lg:flex-row gap-4 border border-gray-400 bg-white p-6 rounded-xl shadow-lg">
            <div className="flex flex-col space-y-8 lg:w-[60%]">
              <div className="text-gray-900 font-semibold text-xl font-satoshi z-10">
                Automatically sync all of your data stores, from local databases
                to s3 buckets with anonymized, production data to safely build
                and test your applications and services.
              </div>
              <div className="flex flex-col justify-start">
                {localTesting.map((item) => (
                  <div
                    className="flex flex-row space-x-4 items-center pt-6"
                    key={item.name}
                  >
                    <CheckCircledIcon className=" w-6 h-6" />{' '}
                    <div className="text-sm lg:text-lg text-gray-700 font-satoshi font-semibold">
                      {item.description}
                    </div>
                  </div>
                ))}
              </div>
            </div>
            <div className="flex flex-col lg:p-5 text-xs space-y-5">
              <ApiCodeBlock code={apiCode} />
            </div>
          </div>
          <div className="flex flex-col-reverse lg:flex-row gap-8 border border-gray-400 bg-white p-6 rounded-xl shadow-lg">
            <div className="lg:w-[60%] rounded-lg">
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/cicode.svg"
                alt="pre"
                className="w-full"
                width="549"
                height="554"
              />
            </div>
            <div className="flex flex-col space-y-8 lg:w-[60%]">
              <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi z-10">
                Use GitOps to hydrate your CI databases with synthetic data
              </div>
              <div className="flex flex-col justify-start">
                {citesting.map((item) => (
                  <div
                    className="flex flex-row space-x-4 items-center pt-6"
                    key={item.name}
                  >
                    <CheckCircledIcon className="w-6 h-6" />{' '}
                    <div className="text-sm lg:text-lg text-gray-700  font-satoshi font-semibold">
                      {item.description}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
