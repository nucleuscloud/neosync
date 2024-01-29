import { GitHubLogoIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { DiMysql } from 'react-icons/di';
import { FaAws, FaDocker } from 'react-icons/fa';
import { SiKubernetes, SiPostgresql } from 'react-icons/si';

export default function Intergrations(): ReactElement {
  const integrations = [
    {
      name: 'Postgres',
      logo: <SiPostgresql className="text-gray-300 bg-transparent w-12 h-12" />,
    },
    {
      name: 'Mysql',
      logo: <DiMysql className="text-gray-300 bg-transparent w-12 h-12" />,
    },
    {
      name: 'AWS S3',
      logo: <FaAws className="text-gray-300 bg-transparent w-12 h-12" />,
    },
    {
      name: 'Github Actions',
      logo: (
        <GitHubLogoIcon className="text-gray-300 bg-transparent w-12 h-12" />
      ),
    },
    {
      name: 'Kubernetes',
      logo: <SiKubernetes className="text-gray-300 bg-transparent w-12 h-12" />,
    },
    {
      name: 'Docker',
      logo: <FaDocker className="text-gray-300 bg-transparent w-12 h-12" />,
    },
  ];
  return (
    <div>
      <div className="text-gray-200 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        Integrations With Your Favorite Tools
      </div>
      <div className="text-md text-gray-400 font-satoshi font-semibold pt-10 lg:px-60 text-center">
        Seamlessly integrate Neosync into your stack with out-of-the box
        integrations
      </div>
      <div className="lg:p-6  mt-10 lg:mx-40">
        <div className="flex justify-center">
          <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
            {integrations.map((item) => (
              <div
                key={item.name}
                className=" p-6 lg:px-16 lg:py-10 border border-gray-600 rounded-xl shadow-xl"
              >
                <div className="flex flex-col gap-4 justify-center items-center">
                  <div>{item.logo}</div>
                  <div className="text-gray-300 text-sm">{item.name}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
