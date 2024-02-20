import Image from 'next/image';
import { ReactElement } from 'react';

export default function TeamSection(): ReactElement {
  const team = [
    {
      image:
        'https://assets.nucleuscloud.com/neosync/blog/authorHeadshots/evis.png',
      name: 'Evis Drenova',
      title: 'CEO & Co-Founder',
      prev: 'Lead PM @ Skyflow',
      linkedin: 'https://www.linkedin.com/in/evisdrenova/',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/blog/authorHeadshots/nick_headshot.png',
      name: 'Nick Zelei',
      title: 'CTO & Co-Founder',
      prev: 'Staff Engineer @ Newfront',
      linkedin: 'https://www.linkedin.com/in/nick-zelei/',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/blog/authorHeadshots/alisha.png',
      name: 'Alisha Kawaguchi',
      title: 'Founding Engineer',
      prev: 'Sr. Engineer @ Newfront',
      linkedin: 'https://www.linkedin.com/in/alishakawaguchi/',
    },
  ];
  return (
    <div className="flex px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px]">
      <div className="flex flex-col-reverse lg:flex-row  shadow-lg rounded-xl bg-gradient-to-tr from-[#0F0F0F] to-[#262626] text-gray-200">
        <div className="flex flex-col text-center pb-[15px] lg:pb-[64px] items-center pt-[44px] lg:mx-40px">
          <div className="font-satoshi pt-16 text-lg px-2 lg:px-40">
            Neosync is built by a small, focused team based in San Francisco,
            California working on challenging problems at the intersection of
            data privacy, security and developer experience.
          </div>
          <div className="flex flex-col lg:flex-row gap-4 pt-10">
            {team.map((item) => (
              <TeamHeadshots
                key={item.name}
                image={item.image}
                name={item.name}
                title={item.title}
                prev={item.prev}
                linkedin={item.linkedin}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export interface Headshots {
  image: string;
  name: string;
  title: string;
  prev?: string;
  linkedin?: string;
}

function TeamHeadshots(props: Headshots): ReactElement {
  const { image, name, title, prev, linkedin } = props;
  return (
    <div className="rounded-xl border w-[200px] lg:w-[320px] border-gray-700 bg-gradient-to-tr from-[#0F0F0F] to-[#191919] ">
      <div className="flex flex-col space-y-1 items-center text-center mx-[16px] py-[32px]">
        <Image
          src={image}
          alt="employee_pic"
          height="100"
          width="100"
          className="w-[80px] h-[80px] rounded-xl"
        />
        <div className="flex flex-row items-center">
          <h2 className="font-satoshi text-xl text-gray-200">{name}</h2>
        </div>
        <div className="font-satoshi text-gray-400">{title}</div>
        <div className="font-satoshi text-ld text-gray-600">
          Previously {prev}
        </div>
      </div>
    </div>
  );
}
