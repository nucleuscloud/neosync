'use client';
import { LinkBreak1Icon } from '@radix-ui/react-icons';
import Image from 'next/image';
import { ReactElement } from 'react';
import { BsFunnel } from 'react-icons/bs';
import { PiArrowsSplitLight, PiFlaskLight } from 'react-icons/pi';
import { BentoGrid, BentoGridItem } from '../ui/bento-grid';

export default function Platform(): ReactElement {
  const items = [
    {
      title: 'Reliable Orchestration',
      description:
        'Neosync supports async scheduling, retries, alerting and syncing across multiple destinations.',
      header: (
        <Image
          src={'/images/conn.svg'}
          alt="st"
          width="690"
          height="290"
          className="rounded-xl border border-gray-700 shadow-xl max-h-[290px] "
        />
      ),
      className: 'md:col-span-2',
      icon: <PiArrowsSplitLight className="h-4 w-4 text-gray-100" />,
    },
    {
      title: 'Synthetic Data',
      description: 'Choose from 40+ Synthetic Data Transformers',
      header: (
        <Image
          src={'/images/bento-tf.svg'}
          alt="st"
          width="310"
          height="277"
          className="rounded-xl border border-gray-700 shadow-xl max-h-[290px]"
        />
      ),
      className: 'md:col-span-1',
      icon: <PiFlaskLight className="h-4 w-4 text-gray-100" />,
    },

    {
      title: 'Subsetting',
      description: 'Flexibly subset your database.',
      header: (
        <Image
          src={'/images/bento-subset.svg'}
          alt="st"
          width="310"
          height="277"
          className="rounded-xl border border-gray-700 shadow-xl max-h-[290px]"
        />
      ),
      className: 'md:col-span-1',
      icon: <BsFunnel className="h-4 w-4 text-gray-100" />,
    },
    {
      title: 'Data Anonymization',
      description:
        'Use a Transformer to anonymize your data or create you own transformation in code. ',
      header: (
        <Image
          src={'/images/datanon.svg'}
          alt="st"
          width="690"
          height="290"
          className="rounded-xl border border-gray-700 shadow-xl max-h-[290px]"
        />
      ),
      className: 'md:col-span-2',
      icon: <LinkBreak1Icon className="h-4 w-4 text-gray-100" />,
    },
  ];

  return (
    <div>
      <div className="text-gray-200 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        A Modern Platform Built for Teams Who Care About Data Security
      </div>

      <div className=" p-6 lg:p-10 rounded-xl mt-10 ">
        <BentoGrid className="mx-auto md:auto-rows-[20rem] max-w-[1092px]">
          {items.map((item, i) => (
            <BentoGridItem
              key={i}
              title={item.title}
              description={item.description}
              header={item.header}
              className={item.className}
              icon={item.icon}
            />
          ))}
        </BentoGrid>
      </div>
    </div>
  );
}
