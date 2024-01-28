'use client';
import CodeBlock from '@/components/codeblocks/TerraformCode';
import IconContainer from '@/components/containers/IconContainer';
import { CheckCircledIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { AiOutlineAudit, AiOutlineCloudSync } from 'react-icons/ai';
import { BiGitBranch } from 'react-icons/bi';
import { BsFileCode } from 'react-icons/bs';

export default function GitOpsSection(): ReactElement {
  const jobCode = `resource "neosync_job" "staging-sync-job" {
    name = "prod-to-stage"

    source_id = neosync_postgres_connection.prod_db.id
    destination_ids = [
      neosync_postgres_connection.stage_db.id,
      neosync_s3_connection.stage_backup.id,
    ]

    schedule = "0 23 * * *" # 11pm every night

    halt_on_new_column_addition = false

    mappings = [
      {
        "schema" : "public",
        "table" : "users",
        "column" : "account_number",
        "transformer" : "custom_accout_number",
      },
      {
        "schema" : "public",
        "table" : "users",
        "column" : "address",
        "transformer" : "address_anonymize"
      },
    ]
  }`;

  const vp = [
    {
      name: 'iac',
      icon: (
        <IconContainer
          icon={<BsFileCode className="text-gray-400 w-[30px] h-[30px]" />}
        />
      ),
      description: ' Manage your test infrastructure in code',
    },
    {
      name: 'create',
      icon: (
        <IconContainer
          icon={<BiGitBranch className="text-gray-400 w-[30px] h-[30px]" />}
        />
      ),
      description: '  Easily create jobs, connections, mappings and more',
    },
    {
      name: 'audit',
      icon: (
        <IconContainer
          icon={<AiOutlineAudit className="text-gray-400 w-[30px] h-[30px]" />}
        />
      ),
      description: ' Audit and track changes across teams',
    },
    {
      name: 'centralize',
      icon: (
        <IconContainer
          icon={
            <AiOutlineCloudSync className="text-gray-400 w-[30px] h-[30px]" />
          }
        />
      ),
      description: 'Centralize your configurations in one place',
    },
  ];
  return (
    <div className="bg-[#F5F5F5] pt-20">
      <div className="flex flex-col items-center lg:py-40 bg-[#F5F5F5]  pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className="text-gray-900 font-semibold  text-2xl lg:text-4xl font-satoshi z-10 w-full">
          Synthetic Data meets GitOps
        </div>
        <div className="flex flex-col lg:flex-row p-4 mt-10 border border-gray-400 shadow-lg rounded-xl text-gray-200 bg-white">
          <div className="flex flex-col p-2 lg:p-10">
            <div className=" pl-2 text-xl font-satoshi font-normal text-gray-900">
              Neosync is built with DevOps and infrastructure teams in mind. Use
              frameworks you know like terraform to manage your Neosync
              infrastructure and even create new jobs.
            </div>
            <div className="flex flex-col pt-10 space-y-10 pl-2">
              <div className="flex flex-col justify-start">
                {vp.map((item) => (
                  <div
                    className="flex flex-row space-x-4 items-center pt-6"
                    key={item.name}
                  >
                    <CheckCircledIcon className="w-6 h-6 text-gray-700" />{' '}
                    <div className="text-sm lg:text-lg text-gray-700 font-satoshi font-semibold">
                      {item.description}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
          <div className="z-1 lg:text-md text-sm">
            <CodeBlock code={jobCode} />
          </div>
        </div>
      </div>
    </div>
  );
}
