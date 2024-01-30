'use client';
import ApiCodeBlock from '@/components/codeblocks/ApiCode';
import { CheckCircleIcon } from 'lucide-react';
import { ReactElement } from 'react';

export default function APISection(): ReactElement {
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

  const features = [
    'World class developer documentation',
    'CLI to work with Neosync locally',
    'Flexible set up your automation and schedules',
    'SDKs available in most popular languages',
    'Hydrate a local database with privacy-safe data',
  ];

  return (
    <div className="flex flex-col items-center justify-center">
      <div className="pt-10 text-center text-gray-900 font-semibold  text-2xl lg:text-4xl font-satoshi z-10">
        Designed and Built for Developers
      </div>
      <div className="flex flex-col-reverse lg:flex-row lg:gap-10 p-4 mt-10">
        <div className="text-sm shadow-xl">
          <ApiCodeBlock code={apiCode} />
        </div>
        <div className="flex flex-col p-2 lg:p10">
          <div className="flex flex-row pl-2">
            <div className="w-[100%] lg:w-[90%] pl-2 text-lg font-normal text-gray-900">
              Neosync&apos;s APIs, SDKs and CLI help developers easily integrate
              their local and staging environments with safe, production-like
              data.
            </div>
          </div>
          <div className="flex flex-col pt-10 space-y-10 pl-2">
            {features.map((item) => (
              <div className="flex flex-row items-center space-x-4" key={item}>
                <CheckCircleIcon className="w-4 h-4" />
                <div className="text-left text-lg text-gray-900 font-semibold">
                  {item}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
