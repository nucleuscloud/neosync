import { RocketIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { IoTrophyOutline } from 'react-icons/io5';
import { MdLightbulbOutline } from 'react-icons/md';
import { TbBrandOpenSource } from 'react-icons/tb';

export default function Values(): ReactElement {
  const values = [
    {
      title: 'Be Transparent',
      description: `As an open source company, we strongly believe in transparency. You can see pretty much everything on our Github, from code to issues, pull requests, and more.`,
      icon: <TbBrandOpenSource className="w-8 h-8" />,
    },
    {
      title: 'Solve Hard Problems',
      description: `We want to build something meaningful and nothing meaningful comes easy. We tackle hard problems and think of innovative ways to approach problems. `,
      icon: <MdLightbulbOutline className="w-8 h-8" />,
    },
    {
      title: 'Ship Faster',
      description: `Speed is everything. We prioritize ruthlessly and ship code every single day. We're always asking ourselves 'what is the most important thing we can be working on right now'.`,
      icon: <RocketIcon className="w-8 h-8" />,
    },
    {
      title: 'Have Pride',
      description: `We take pride in what we do. Whether it's code we write, blogs we publish, the way we talk to customers or how we talk to each other. We're proud of our work.`,
      icon: <IoTrophyOutline className="w-8 h-8" />,
    },
  ];
  return (
    <div>
      <div className="px-6">
        <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
          Our Values
        </div>
      </div>
      <div className="text-lg text-gray-400 font-satoshi font-light  pt-10 lg:pt-20 flex flex-col lg:flex-row gap-6 justify-center items-center">
        {values.map((item) => (
          <div
            key={item.title}
            className="border border-gray-400 bg-white rounded-xl p-8 shadow-xl flex flex-col justify-between gap-6 text-center w-full lg:w-[480px] max-w-xs mx-auto lg:h-auto"
          >
            <div>
              <div className="text-gray-900 justify-center flex">
                {item.icon}
              </div>
              <div className="text-gray-900 text-2xl pt-10">{item.title}</div>
              <div className=" text-gray-500 text-[16px] pt-6 text-center">
                {item.description}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
