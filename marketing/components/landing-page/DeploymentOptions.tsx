'use client';
import { ArrowTopRightIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';

export default function DeploymentOptions(): ReactElement {
  const deployOptions = [
    {
      name: 'Open Source',
      description:
        'Deploy on your infrastructure using a Helm Chart or Docker Compose file on your infrastructure.',
      link: 'github.com/nucleuscloud/neosync',
      image: '/images/osdeploy.svg',
    },
    {
      name: 'Neosync Cloud',
      description:
        "Use Neosync's fully managed cloud to not worry about infrastructure and sign up now.",
      link: 'https://app.neosync.dev',
      image: '/images/osdeploy.svg',
    },
    {
      name: 'Neosync Hybrid',
      description:
        'Keep your data in your infrastructure but get the ease of a managed control plane.',
      link: 'https://calendly.com/evis1/30min',
      image: '/images/osdeploy.svg',
    },
  ];

  return (
    <div className="px-6 gap-6">
      <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        Flexible Deployment Options
      </div>
      <div className="text-md text-gray-700 font-satoshi font-semibold pt-10 lg:px-60 text-center">
        Choose the deployment option that works best for you and your
        infrastructure.
      </div>
      <div className="flex flex-row gap-4 pt-20 justify-center">
        {deployOptions.map((opt) => (
          <Link
            href={opt.link}
            target="_blank"
            key={opt.name}
            className="flex flex-col gap-2 border rounded-xl shadow-xl p-6 bg-white w-full  max-w-[300px] lg:max-w-[400px]h-[360px] hover:border-gray-400 items-center"
          >
            <Image
              src={opt.image}
              alt="NeosyncLogo"
              width="178"
              height="150"
              className="w-[178px] h-[150px]"
            />
            <div className="flex flex-row items-center gap-4 pt-6">
              <div className="text-xl font-semibold text-al">{opt.name}</div>
              <ArrowTopRightIcon />
            </div>
            <div className="text-sm text-gray-500">{opt.description}</div>
          </Link>
        ))}
      </div>
    </div>

    // <div className="px-6">
    //   <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
    //     Flexible Deployment Options
    //   </div>
    //   <div className="text-md text-gray-700 font-satoshi font-semibold pt-10 lg:px-60 text-center">
    //     Whether you decide to host Neosync yourself or use Neoysnc Cloud,
    //     you&apos;ll get access to the same powerful features.
    //   </div>
    //   <div className="flex flex-col lg:flex-row items-center justify-center pt-20 gap-4 o">
    //     <div className="rounded-xl shadow-xl bg-[#1E1E24] text-white flex-flex-col">
    //       <div className="flex flex-col gap-4 py-10">
    //         <Image
    //           src={'/images/logo_dark_mode.svg'}
    //           alt="NeosyncLogo"
    //           className="w-5 object-scale-down ml-10"
    //           width="64"
    //           height="20"
    //         />
    //         <div className="flex flex-row items-center gap-2 pl-10">
    //           <div className="text-xl font-semibold">Neosync</div>
    //           <div className="text-sm">Cloud</div>
    //         </div>
    //         <div className="text-sm pl-10">
    //           Don&apos;t worry about infrastructure and sign up now
    //         </div>
    //         <Link href="https://app.neosync.dev" target="_blank">
    //           <Button
    //             className="px-6 w-[188px] ml-10"
    //             variant="secondary"
    //             onClick={() =>
    //               posthog.capture('user click', {
    //                 page: 'deployment options app sign up',
    //               })
    //             }
    //           >
    //             Get Started Now <ArrowRightIcon className="ml-2 h-4 w-4" />
    //           </Button>
    //         </Link>
    //       </div>
    //       <Image
    //         src={'/images/ss-dark-new.svg'}
    //         alt="NeosyncLogo"
    //         width="624"
    //         height="280"
    //         className="w-full rounded-b-xl"
    //       />
    //     </div>
    //     <div className="border border-gray-400 rounded-xl shadow-xl  bg-white flex-flex-col overflow-hidden">
    //       <div className="flex flex-col gap-4 py-10">
    //         <Image
    //           src={'/images/logo_light_mode.svg'}
    //           alt="NeosyncLogo"
    //           className="w-5 object-scale-down ml-10"
    //           width="64"
    //           height="20"
    //         />
    //         <div className="flex flex-row items-center gap-2 pl-10">
    //           <div className="text-xl font-semibold">Neosync</div>
    //           <div className="text-sm">Open Source</div>
    //         </div>
    //         <div className="text-sm pl-10">
    //           Deploy using a Helm Chart or Docker Compose file.
    //         </div>
    //         <Button
    //           className="px-6 w-[188px] ml-10"
    //           variant="default"
    //           onClick={() =>
    //             posthog.capture('user click', {
    //               page: 'deployment options open source',
    //             })
    //           }
    //         >
    //           <Link href="https://github.com/nucleuscloud/neosync">
    //             <div className="flex flex-row gap-2">
    //               <GitHubLogoIcon className="mr-2 h-5 w-5" /> Open Source
    //             </div>
    //           </Link>
    //         </Button>
    //       </div>
    //       <Image
    //         src={'/images/ss-light-new.svg'}
    //         alt="NeosyncLogo"
    //         width="624"
    //         height="280"
    //         className="w-full"
    //       />
    //     </div>
    //   </div>
    // </div>
  );
}
