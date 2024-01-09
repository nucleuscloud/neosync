'use client';
import GradientTag from '@/components/GradientTag';
import ApiCodeBlock from '@/components/codeblocks/ApiCode';
import CodeBlock from '@/components/codeblocks/TerraformCode';
import IconContainer from '@/components/containers/IconContainer';
import CTA from '@/components/cta/CTA';
import { Button } from '@/components/ui/button';
import { HeroandGrid } from '@/public/heroandgrid';
import { SubsetAnimation } from '@/public/subsetanimation';
import { ApacheSpark } from '@/styles/icons/ApacheSpark';
import { AwsRedshift } from '@/styles/icons/AwsRedshift';
import { AWSS3 } from '@/styles/icons/AwsS3';
import { AzureBlob } from '@/styles/icons/AzureBlob';
import { Databricks } from '@/styles/icons/Databricks';
import { AWSDynamoDB } from '@/styles/icons/DynamoDB';
import { Emr } from '@/styles/icons/EMR';
import { GcpIconColored } from '@/styles/icons/GCP';
import { GCPBigQuery } from '@/styles/icons/GCPBigQuery';
import { GithubActionsSVG } from '@/styles/icons/GithubActions';
import { Kubernetes } from '@/styles/icons/Kubernetes';
import { Snowflake } from '@/styles/icons/Snowflake';
import { MysqlIcon } from '@/styles/icons/mysql';
import {
  ArrowRightIcon,
  CalendarIcon,
  CheckCircledIcon,
  GitHubLogoIcon,
  LockClosedIcon,
} from '@radix-ui/react-icons';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import {
  AiOutlineAudit,
  AiOutlineCloudSync,
  AiOutlineCode,
} from 'react-icons/ai';
import {
  BiExpand,
  BiGitBranch,
  BiLogoPostgresql,
  BiSolidQuoteLeft,
  BiSolidQuoteRight,
} from 'react-icons/bi';
import { BsFileCode, BsGithub, BsRecycle } from 'react-icons/bs';
import { FiPackage } from 'react-icons/fi';
import { MdLibraryBooks } from 'react-icons/md';
import { PiLink } from 'react-icons/pi';
import { SiMongodb } from 'react-icons/si';

export default function Home(): ReactElement {
  return (
    <div>
      <Hero />
      <TableTypes />
      <FeaturesGrid />
      {/* <ProblemStmt /> */}
      <UseCases />
      <Transformers />
      <Subset />
      <GitOpsSection />
      {/* <IntegrationSection /> */}
    </div>
  );
}

function Hero(): ReactElement {
  return (
    <div className="flex flex-col bg-[#FFFFFF] border-b border-b-gray-200 items-center space-y-10 mt-20">
      <div className="text-gray-900 font-semibold lg:text-6xl text-4xl leading-tight text-center z-20 px-2 relative">
        Open Source Synthetic Data Orchestration
      </div>
      <h3 className="text-[#606060]  text-md lg:text-lg font-light relative text-center z-20 ">
        A developer-first way to create anonymized or synthetic data and sync it
        across all environments for high-quality local, stage and CI testing
      </h3>
      <div className="flex flex-col lg:flex-row lg:space-y-0 space-y-2 lg:space-x-4 z-30">
        <Button className="px-4">
          <Link href="https://docs.neosync.dev">
            <div className="flex flex-row">
              Documentation <ArrowRightIcon className="ml-2 h-5 w-5" />
            </div>
          </Link>
        </Button>
        <Button variant="secondary" className="px-6">
          <Link href="https://github.com/nucleuscloud/neosync">
            <div className="flex flex-row">
              <GitHubLogoIcon className="mr-2 h-5 w-5" /> Get started
            </div>
          </Link>
        </Button>
      </div>
      <div className="mt-10 rounded-xl z-10 overflow-hidden">
        <div className="hidden lg:block">
          <HeroandGrid />
        </div>
        <div className="block md:hidden lg:hidden">
          <Image
            src="https://assets.nucleuscloud.com/neosync/marketingsite/mainheromobile.svg"
            alt="pre"
            width="436"
            height="416"
            className="w-full"
          />
        </div>
      </div>
    </div>
  );
}

function TableTypes(): ReactElement {
  const tableTypes = [
    {
      name: 'Tabular',
      description:
        'Generate statistically consistent synthetic data for a single data frame or table.',
      image: (
        <Image
          src="https://assets.nucleuscloud.com/neosync/marketingsite/tabularcropped.png"
          alt="pre"
          width="436"
          height="416"
          className="w-[280px] h-auto"
        />
      ),
    },
    {
      name: 'Relational',
      description:
        'Generate statistically consistent synthetic data for a relational database while maintaining referential integrity.',
      image: (
        <Image
          src="https://assets.nucleuscloud.com/neosync/marketingsite/relationaltablesgrid.png"
          alt="pre"
          width="436"
          height="416"
          className="w-[600px] h-auto"
        />
      ),
    },
  ];
  return (
    <div className="bg-[#F5F5F5] pt-20">
      <div className="flex flex-col  pt-5 lg:pt-40 bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-80 max-w-[1800px] mx-auto ">
        <div className="text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi z-10  ">
          Generate Data for Any Schema
        </div>
        <div className="flex flex-col lg:flex-row space-y-4 lg:space-y-0 items-center justify-between gap-x-20 pt-10">
          {tableTypes.map((item) => (
            <div
              key={item.name}
              className="justify-center border border-gray-400 bg-white shadow-lg rounded-xl p-6 lg:w-[800px] lg:h-[400px]"
            >
              <div className="justify-center flex">{item.image}</div>
              <div className="text-xl text-gray-800 font-satoshi font-semibold pt-10">
                {item.name}
              </div>
              <div className=" text-gray-600 font-satoshi pt-4 ">
                {item.description}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function FeaturesGrid(): ReactElement {
  const features = [
    {
      title: 'Complete referential integrity',
      description: `Neosync automatically handles and preserves your data's referential integrity.`,
      icon: <PiLink />,
    },
    {
      title: 'Retries',
      description:
        'Neosync handles automatically handles retries in the case of an error during the sync process.',
      icon: <BsRecycle />,
    },
    {
      title: 'Scheduling',
      description:
        'Run jobs ad-hoc, pause existing jobs or schedule jobs to run on any cadence you decide.',
      icon: <CalendarIcon />,
    },
    {
      title: 'Scalable',
      description:
        "Neosync's async pipeline is horizontally scalable making it a great choice for bigger data sets.",
      icon: <BiExpand />,
    },
    {
      title: 'Security',
      description:
        'Neosync comes with RBAC out of the box and is designed to be multi-tenant for teams deploying on-prem.',
      icon: <LockClosedIcon />,
    },
    {
      title: 'Audit',
      description:
        'Neosync ships with audit controls so security and compliance teams have a full audit trail over the system.',
      icon: <AiOutlineAudit />,
    },
  ];

  return (
    <div className="bg-[#F5F5F5] pt-20">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className=" flex flex-col space-y-20 bg-gradient-to-tr from-[#0F0F0F] to-[#262626] rounded-2xl py-12 px-6 shadow-lg">
          <div>
            <div className="text-gray-100 font-semibold text-2xl lg:text-5xl font-satoshi pt-10">
              Built for Security and Scalability
            </div>
            <div className="text-lg text-gray-400 font-satoshi font-light pt-8 lg:w-[70%]">
              Neosync is built for teams of all sizes and ships with features
              that put security and compliance teams of all sizes at ease.
              Whether you&apos;re an enterprise who needs an air-gapped
              deployment or a startup looking to get started quickly, Neosync
              can help.
            </div>
            <div className="pt-10">
              <Button variant="secondary" className="px-4">
                <Link href="https://docs.neosync.dev">
                  <div className="flex flex-row">
                    Documentation <ArrowRightIcon className="ml-2 h-5 w-5" />
                  </div>
                </Link>
              </Button>
            </div>
            <div className="text-lg text-gray-400 font-satoshi font-light py-20 grid grid-cols-1 lg:grid-cols-3 gap-6">
              {features.map((item) => (
                <div
                  key={item.title}
                  className="bg-[#282828] border border-[#484848] rounded-lg p-4 hover:-translate-y-2 duration-150 shadow-xl"
                >
                  <div className="flex flex-row items-center space-x-3">
                    <div className="text-gray-100">{item.icon}</div>
                    <div className="text-gray-100 text-xl">{item.title}</div>
                  </div>
                  <div className="text-sm">{item.description}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function ProblemStmt(): ReactElement {
  const router = useRouter();
  return (
    <div className="bg-[#F5F5F5] mt-20">
      <div className="flex flex-col justify-center px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <Button variant="link" className="pt-20 text-gray-400">
          <Link href="/about">
            <div className="flex flex-row items-center space-x-2 w-full justify-center z-20 hover:underline hover:text-gray-800">
              <div className="font-normal text-gray-800 py-2">
                About us & our Investors
              </div>
              <ArrowRightIcon className="w-3 h-3 text-gray-800" />
            </div>
          </Link>
        </Button>
        <div className="p-4 lg:p-10 lg:m-10 mt-10 lg:mt-20 border-2 border-gray-700 rounded-xl shadow-lg z-1 bg-[#262626] z-10 ">
          <div className="flex flex-col">
            <BiSolidQuoteLeft className="w-10 h-10 text-blue-200" />
            <div className="font-normal text-lg text-gray-300 font-satoshi font-20px">
              Test data management is inconsistently adopted across
              organizations, with most teams copying production data for use in
              test environments...this traditional approach is increasingly at
              odds with requirements for efficiency, privacy and security.
            </div>
            <div className="flex justify-end">
              <BiSolidQuoteRight className="w-10 h-10 text-blue-200" />
            </div>
            <div className="flex flex-col pt-10">
              <div className="text-lg font-normal text-gray-200 font-satoshi">
                Gartner
              </div>
              <div className="text-md font-light text-gray-400 font-satoshi">
                Hype Cycle for Agile & DevOps, 2022
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function UseCases(): ReactElement {
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

function Transformers(): ReactElement {
  const transformers = [
    {
      name: 'Pre-built Transformers',
      description: 'Choose from our library of pre-built transformers',
      image: (
        <Image
          src="https://assets.nucleuscloud.com/neosync/marketingsite/pbsvg.svg"
          alt="pre"
          width="436"
          height="416"
          className="w-full"
        />
      ),
    },
    {
      name: 'Custom Transformers',
      description:
        'Create your own transformer for when you need bespoke logic ',
      image: (
        <Image
          src="https://assets.nucleuscloud.com/neosync/marketingsite/customtran.svg"
          alt="pre"
          width="436"
          height="416"
          className="w-full"
        />
      ),
    },
  ];
  return (
    <div className="bg-[#F5F5F5] pt-20">
      <div className="flex flex-col  pt-5 lg:pt-40 bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto ">
        <div className="text-gray-900 font-semibold  text-2xl lg:text-5xl font-satoshi">
          Fully customizable Transformers
        </div>
        <div className="flex flex-col mt-10 gap-4 border border-gray-400 bg-white p-6 rounded-xl shadow-lg">
          <div className="text-xl text-gray-700  font-satoshi font-semibold lg:w-[60%]">
            Anonymize, mask or generate data using a transformer from our
            library of pre-built tranformers or write your own transformation
            logic in code.
          </div>
          <div className="flex flex-col lg:flex-row space-y-4 lg:space-y-0 items-center justify-center gap-x-20 pt-10">
            {transformers.map((item) => (
              <div
                key={item.name}
                className="justify-center border border-gray-400 bg-white shadow-lg rounded-xl p-6"
              >
                <div className="border border-gray-900 overflow-hidden rounded-lg shadow-lg lg:w-[436px]">
                  {item.image}
                </div>
                <div className="text-xl text-gray-800 font-satoshi font-semibold pt-10">
                  {item.name}
                </div>
                <div className=" text-gray-600 font-satoshi pt-4 ">
                  {item.description}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function Subset(): ReactElement {
  return (
    <div className="bg-[#F5F5F5] pt-20">
      <div className="pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className=" flex flex-col space-y-20 bg-gradient-to-tr from-[#0F0F0F] to-[#262626] rounded-2xl py-12 px-6 shadow-lg">
          <div className="text-gray-100 font-semibold  text-2xl lg:text-5xl font-satoshi pt-10">
            Powerful subsetting
          </div>
          <div className="text-lg text-gray-400 font-satoshi font-light pt-8 lg:w-[70%]">
            Subsetting allows you to filter your source database so that you can
            easily reproduce bugs and data errors and shrink your production
            database so that it fits locally. Neosync maintains full referential
            integrity of your data automatically.
          </div>
          <div className="hidden lg:justify-center lg:flex">
            <SubsetAnimation />
          </div>
          <div className="block md:hidden lg:hidden ">
            <Image
              src="https://assets.nucleuscloud.com/neosync/marketingsite/customfilters.png"
              alt="pre"
              width="436"
              height="416"
              className="w-full"
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function GitOpsSection(): ReactElement {
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
    <div className="flex flex-col items-center lg:py-40 bg-[#F5F5F5]  pt-5 lg:pt-40 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
      <div className="text-gray-900 font-semibold  text-2xl lg:text-5xl font-satoshi z-10 w-full">
        Synthetic Data meets GitOps
      </div>
      <div className="flex flex-col lg:flex-row p-4 mt-10 border border-gray-400 shadow-lg rounded-xl text-gray-200">
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
  );
}

function APISection(): ReactElement {
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

function AuditSection(): ReactElement {
  return (
    <div className="flex flex-col pt-20 lg:pt-40 items-center">
      <GradientTag tagValue={'Security and Compliance'} />
      <div className="pt-10 text-center z-3 text-gray-900 font-semibold  text-2xl lg:text-5xlfont-satoshi z-10">
        Access Controls and Audit Logs for Compliance and Security
      </div>
      <div className="flex flex-col lg:flex-row px-1 lg:px-5 pt-10 space-y-5 lg:space-x-5 lg:space-y-0 z-1 items-center z-10">
        <div className="bg-gradient-to-tr from-[#0F0F0F] to-[#191919] border-[1px] border-gray-700  rounded-xl text-gray-200 lg:w-[60%]">
          <div className="flex flex-col p-5">
            <div className="flex flex-col text-left font-semibold text-gray-200 text-xl">
              Fine grained Access Controls
            </div>
            <div className="text-left text-lg text-gray-400 font-semibold">
              Access control govern which members have access to create jobs,
              transformers, invite members and more.
            </div>
            <div className="border border-gray-700 rounded-xl mt-5 overflow-hidden">
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/rbac.png"
                alt="pre"
                width="549"
                height="554"
                className="w-full"
              />
            </div>
          </div>
        </div>
        <div className="bg-gradient-to-tr from-[#0F0F0F] to-[#191919] border-[1px] border-gray-700 rounded-xl text-gray-200 lg:w-[60%]">
          <div className="flex flex-col p-5">
            <div className="flex flex-col text-left font-semibold text-gray-200 text-xl">
              Audit Logs for Compliance
            </div>
            <div className="text-left text-lg text-gray-400 font-semibold">
              Stay in control of your data and track every event in the Audit
              log. Export it to stay up to date with compliance.
            </div>
            <div className="border border-gray-700 rounded-xl overflow-hidden mt-5">
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/audittrimmed.png"
                alt="pre"
                className="w-full"
                width="549"
                height="554"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function IntegrationSection(): ReactElement {
  const developerFeatures = [
    {
      title: 'Github',
      icon: <BsGithub className="w-10 h-10" />,
    },
    {
      title: 'Postgres',
      icon: <BiLogoPostgresql className="w-10 h-10 text-blue-300" />,
    },
    {
      title: 'MySQL',
      icon: <MysqlIcon />,
    },
    {
      title: 'MongoDB',
      icon: <SiMongodb className="w-10 h-10 text-green-400" />,
    },
    {
      title: 'AWS S3',
      icon: <AWSS3 />,
    },
    {
      title: 'Github Actions',
      icon: <GithubActionsSVG />,
    },
    {
      title: 'AWS DynamoDb',
      icon: <AWSDynamoDB />,
    },
    {
      title: 'GCP Cloud Storage',
      icon: <GcpIconColored />,
    },
    {
      title: 'Kuberentes',
      icon: <Kubernetes />,
    },
    {
      title: 'GCP Big Query',
      icon: <GCPBigQuery />,
    },
    {
      title: 'Apache Spark',
      icon: <ApacheSpark />,
    },
    {
      title: 'Snowflake',
      icon: <Snowflake />,
    },
    {
      title: 'Databricks',
      icon: <Databricks />,
    },
    {
      title: 'AWS Redshift',
      icon: <AwsRedshift />,
    },
    {
      title: 'AWS EMR',
      icon: <Emr />,
    },
    {
      title: 'Azure Blob Storage',
      icon: <AzureBlob />,
    },
  ];
  return (
    <div className="flex flex-col justify-center items-center pt-10 lg:pt-20">
      <GradientTag tagValue={'Integrations'} />
      <div className="pt-10 text-center z-3 text-gray-900 font-semibold  text-2xl lg:text-5xl font-satoshi">
        Easily integrate your existing tools and workflows
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-12 pt-20 z-10">
        {developerFeatures.map((item) => (
          <IntegrationIcon
            key={item.title}
            icon={item.icon}
            title={item.title}
          />
        ))}
      </div>
      <div className="w-full z-10">
        <CTA />
      </div>
    </div>
  );
}
interface IntegrationIconProps {
  title: string;
  icon: ReactElement;
}

function IntegrationIcon(props: IntegrationIconProps): ReactElement {
  const { title, icon } = props;

  return (
    <div className="flex flex-row space-x-5 p-5 align-middle items-center bg-gradient-to-tr from-[#0F0F0F] to-[#191919]  rounded-xl w-full w-60-lg z-1 text-gray-300 border-gray-700 border">
      <div>{icon}</div>
      <div className="text-gray-200">{title}</div>
    </div>
  );
}
