import { NeonTech } from '@/styles/icons/Neon';
import { Supabase } from '@/styles/icons/Supabase';
import { ExternalLinkIcon, GitHubLogoIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';
import { DiMysql } from 'react-icons/di';
import { FaAws, FaDocker } from 'react-icons/fa';
import { SiKubernetes, SiPostgresql } from 'react-icons/si';

export default function Intergrations(): ReactElement {
  const integrations = [
    {
      name: 'Postgres',
      logo: <SiPostgresql className="text-blue-600 bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/connections/postgres',
    },
    {
      name: 'Mysql',
      logo: <DiMysql className="text-blue-700 bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/connections/mysql',
    },
    {
      name: 'AWS S3',
      logo: <FaAws className="text-orange-400 bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/connections/s3',
    },
    {
      name: 'Github Actions',
      logo: (
        <GitHubLogoIcon className="text-gray-900 bg-transparent w-12 h-12" />
      ),
      href: 'https://docs.neosync.dev/guides/using-neosync-in-ci',
    },
    {
      name: 'Kubernetes',
      logo: <SiKubernetes className="text-blue-500 bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/deploy/kubernetes',
    },
    {
      name: 'Docker',
      logo: <FaDocker className="text-blue-400  bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/deploy/docker-compose',
    },
    {
      name: 'Supabase',
      logo: <Supabase />,
      href: 'https://docs.neosync.dev/connections/postgres',
    },
    {
      name: 'Neon',
      logo: <NeonTech />,
      href: 'https://www.neosync.dev/blog/neosync-neon-data-gen-job',
    },
    {
      name: 'AWS RDS',
      logo: (
        <Image
          src="/images/rds.svg"
          width="40"
          height="48"
          alt="rds"
          className="min-w-[40px] min-h-[48px]"
        />
      ),
      href: 'https://docs.neosync.dev/connections/postgres',
    },
  ];
  return (
    <div>
      <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center z-30">
        Seamlessly Integrate With Your Stack
      </div>
      <div className="lg:p-6 mt-10">
        <div className="flex justify-center">
          <div className="grid grid-cols-2 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {integrations.map((item) => (
              <Link key={item.name} href={item.href}>
                <div className="p-6 lg:px-16 lg:py-10 border border-gray-300 bg-gradient-to-tr from-[#ffffff] to-[#f2f1f1] rounded-xl shadow-xl transition duration-150 ease-in-out hover:-translate-y-1 relative">
                  <div className="absolute top-0 right-0 p-2">
                    <ExternalLinkIcon className="w-4 h-4 text-gray-500" />
                  </div>
                  <div className="flex flex-col gap-4 justify-center items-center ">
                    <div>{item.logo}</div>
                    <div className="text-gray-900 text-sm">{item.name}</div>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
