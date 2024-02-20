import { NeonTech } from '@/styles/icons/Neon';
import { ExternalLinkIcon, GitHubLogoIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';
import { DiMysql } from 'react-icons/di';
import { FaAws, FaDocker } from 'react-icons/fa';
import { SiKubernetes, SiPostgresql, SiSupabase } from 'react-icons/si';

export default function Intergrations(): ReactElement {
  const integrations = [
    {
      name: 'Postgres',
      logo: <SiPostgresql className="text-gray-300 bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/connections/postgres',
    },
    {
      name: 'Mysql',
      logo: <DiMysql className="text-gray-300 bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/connections/mysql',
    },
    {
      name: 'AWS S3',
      logo: <FaAws className="text-gray-300 bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/connections/s3',
    },
    {
      name: 'Github Actions',
      logo: (
        <GitHubLogoIcon className="text-gray-300 bg-transparent w-12 h-12" />
      ),
      href: 'https://docs.neosync.dev/guides/using-neosync-in-ci',
    },
    {
      name: 'Kubernetes',
      logo: <SiKubernetes className="text-gray-300 bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/deploy/kubernetes',
    },
    {
      name: 'Docker',
      logo: <FaDocker className="text-gray-300 bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/deploy/docker-compose',
    },
    {
      name: 'Supabase',
      logo: <SiSupabase className="text-gray-300 bg-transparent w-12 h-12" />,
      href: 'https://docs.neosync.dev/connections/postgres',
    },
    {
      name: 'Neon',
      logo: <NeonTech />,
      href: 'https://docs.neosync.dev/connections/postgres',
    },
    {
      name: 'AWS RDS',
      logo: <Image src="/images/rds.svg" width="48" height="48" alt="rds" />,
      href: 'https://docs.neosync.dev/deploy/docker-compose',
    },
  ];
  return (
    <div>
      <div className="text-gray-200 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        Integrate Your Stack
      </div>
      <div className="text-md text-gray-400 font-satoshi font-semibold pt-10 lg:px-60 text-center">
        Seamlessly integrate Neosync into your stack with out-of-the box
        integrations
      </div>

      <div className="lg:p-6  mt-10 lg:mx-40">
        <div className="flex justify-center">
          <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
            {integrations.map((item) => (
              <Link key={item.name} href={item.href}>
                <div className="p-6 lg:px-16 lg:py-10 border border-gray-600 bg-gradient-to-tr from-[#1E1E1E] to-[#2c2b2b] rounded-xl shadow-xl transition duration-150 ease-in-out hover:-translate-y-1 relative">
                  <div className="absolute top-0 right-0 p-2">
                    <ExternalLinkIcon className="w-4 h-4 text-gray-500" />
                  </div>
                  <div className="flex flex-col gap-4 justify-center items-center">
                    <div>{item.logo}</div>
                    <div className="text-gray-300 text-sm">{item.name}</div>
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
