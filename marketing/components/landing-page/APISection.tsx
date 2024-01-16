'use client';
import GradientTag from '@/components/GradientTag';
import ApiCodeBlock from '@/components/codeblocks/ApiCode';
import IconContainer from '@/components/containers/IconContainer';
import { ReactElement } from 'react';
import { AiOutlineCloudSync, AiOutlineCode } from 'react-icons/ai';
import { FiPackage } from 'react-icons/fi';
import { MdLibraryBooks } from 'react-icons/md';

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
  return (
    <div className="flex flex-col items-center justify-center pt-20 lg:pt-40">
      <GradientTag tagValue={'Developers'} />
      <div className="pt-10 text-center text-gray-900 font-semibold  text-2xl lg:text-5xl font-satoshi z-10">
        Designed and Built for Developers
      </div>
      <div className="flex flex-col-reverse lg:flex-row p-4 mt-10 border-2 border-gray-700 bg-gradient-to-tr from-[#0F0F0F] to-[#191919] shadow-lg z-1 rounded-xl text-gray-200 z-10">
        <div className="z-1 text-sm">
          <ApiCodeBlock code={apiCode} />
        </div>
        <div className="flex flex-col p-2 lg:p10">
          <div className="flex flex-row pl-2">
            <div className="w-1 bg-purple-400 rounded-xl" />
            <div className="w-[100%] lg:w-[90%] pl-2 text-lg font-normal text-gray-200">
              Neosync&apos;s APIs, SDKs and CLI help developers easily integrate
              their local and staging environments with safe, production-like
              data.
            </div>
          </div>
          <div className="flex flex-col pt-10 space-y-10 pl-2">
            <div className="flex flex-row items-center space-x-4">
              <IconContainer
                icon={
                  <MdLibraryBooks className="text-gray-400 w-[30px] h-[30px]" />
                }
              />
              <div className="text-left text-lg text-gray-400 font-semibold">
                World class developer documentation
              </div>
            </div>
            <div className="flex flex-row items-center space-x-4">
              <IconContainer
                icon={
                  <AiOutlineCode className="text-gray-400 w-[30px] h-[30px]" />
                }
              />
              <div className="text-left text-lg text-gray-400 font-semibold">
                A CLI to work with Neosync locally
              </div>
            </div>
            <div className="flex flex-row items-center space-x-4">
              <IconContainer
                icon={
                  <AiOutlineCloudSync className="text-gray-400 w-[30px] h-[30px]" />
                }
              />
              <div className="text-left text-lg text-gray-400 font-semibold">
                Centralize your configurations in one place
              </div>
            </div>
            <div className="flex flex-row items-center space-x-4">
              <IconContainer
                icon={<FiPackage className="text-gray-400 w-[30px] h-[30px]" />}
              />
              <div className="text-left text-lg text-gray-400 font-semibold">
                SDKs available in Go, Java, Python, Typescript and Ruby
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
