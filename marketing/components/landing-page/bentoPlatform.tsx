import { LinkBreak1Icon } from '@radix-ui/react-icons';
import Image from 'next/image';
import { BsFunnel } from 'react-icons/bs';
import { PiArrowsSplitLight, PiFlaskLight } from 'react-icons/pi';
import { BentoGrid, BentoGridItem } from '../ui/bento-grid';

export function BentoGridSecondDemo() {
  return (
    <BentoGrid className="mx-auto md:auto-rows-[20rem]">
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
  );
}

const items = [
  {
    title: 'Reliable Orchestration',
    description:
      'Neosync supports async scheduling, retries, alerting and syncing across multiple destinations.',
    header: (
      <Image
        src={'/images/bento-orchestration.svg'}
        alt="st"
        width="690"
        height="290"
        className="rounded-xl border border-gray-700 shadow-xl"
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
        className="rounded-xl border border-gray-700 shadow-xl"
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
        className="rounded-xl border border-gray-700 shadow-xl"
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
        src={'/images/bento-data-anon.svg'}
        alt="st"
        width="690"
        height="290"
        className="rounded-xl border border-gray-700 shadow-xl"
      />
    ),
    className: 'md:col-span-2',
    icon: <LinkBreak1Icon className="h-4 w-4 text-gray-100" />,
  },
];
