'use client';
import GradientTag from '@/components/GradientTag';
import CTA from '@/components/cta/CTA';
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
import { ReactElement } from 'react';
import { BiLogoPostgresql } from 'react-icons/bi';
import { BsGithub } from 'react-icons/bs';
import { SiMongodb } from 'react-icons/si';

export default function IntegrationSection(): ReactElement {
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
