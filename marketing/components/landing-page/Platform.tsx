'use client';
import { LinkBreak1Icon } from '@radix-ui/react-icons';
import Image from 'next/image';
import { ReactElement } from 'react';
import { BsFunnel } from 'react-icons/bs';
import { GoWorkflow } from 'react-icons/go';
import { PiArrowsSplitLight, PiFlaskLight } from 'react-icons/pi';

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
      <div className="text-gray-900 font-semibold text-2xl lg:text-4xl font-satoshi text-center bg-white/60">
        Built for Teams Who Care About Data Security
      </div>
      <div className="flex flex-col items-center justify-center border border-bg-gray-300 rounded-xl mt-20 z-30 bg-white max-w-[1105px] overflow-hidden">
        <GenerateAndAnonymize />
        <ReferentialIntegrity />
        <OrchestrationAndSubset />
      </div>
      <div className=" p-6 lg:p-10 rounded-xl mt-10 "></div>
    </div>
  );
}

function GenerateAndAnonymize(): ReactElement {
  return (
    <div className="flex flex-row items-center bg-white w-full">
      <div className="flex flex-col gap-6 p-10 w-full">
        <div className="flex flex-col gap-2">
          <div className="flex flex-row items-center gap-2">
            <div>
              <PiFlaskLight className="h-4 w-4" />
            </div>
            <div className="text-2xl font-semibold text-left">
              Generate Synthetic Data
            </div>
          </div>
          <div className="text-md font-normal text-left pl-6">
            Choose from 40+ pre-built Synthetic Data Transformers.
          </div>
        </div>
        <Image
          src={'/images/syste.svg'}
          alt="st"
          width="443"
          height="333"
          className="border border-gray-300 shadow-lg rounded-xl"
        />
      </div>
      <div className="flex flex-col gap-4 p-10 border-gray-300 border-l-2 w-full">
        <div className="flex flex-row items-center gap-2">
          <div>
            <LinkBreak1Icon className="h-4 w-4" />
          </div>
          <div className="text-2xl font-semibold text-left">
            Anonymize Sensitive Data
          </div>
        </div>
        <div className="text-md font-normal text-left pl-6">
          Mask, redact, scramble or obfuscate sensitive data in code.
        </div>
        <Image
          src={'/images/customTra.svg'}
          alt="st"
          width="443"
          height="333"
          className="border border-gray-300 shadow-lg rounded-xl"
        />
      </div>
    </div>
  );
}

function ReferentialIntegrity(): ReactElement {
  return (
    <div className="flex flex-col gap-4 p-10 w-full border-t-2 border-gray-300 bg-gradient-to-tr from-[#f8f8f8] to-[#eaeaea]">
      <div className="flex flex-row items-center gap-2">
        <div>
          <GoWorkflow className="h-4 w-4" />
        </div>
        <div className="text-2xl font-semibold text-left">
          Full Referential Integrity
        </div>
      </div>
      <div className="text-md font-normal text-left pl-6">
        Neosync perfectly preserves your data&apos;s referential integrity and
        can handle the most complex schemas.
      </div>
      <Image
        src={'/images/schemaRef.svg'}
        alt="st"
        width="1021"
        height="392"
        className="border border-gray-300 shadow-xl rounded-xl"
      />
    </div>
  );
}

function OrchestrationAndSubset(): ReactElement {
  return (
    <div className="flex flex-row gap-4 w-full border-t-2 border-gray-300">
      <div className="flex flex-col gap-4 p-10 w-full">
        <div className="flex flex-row items-center gap-2">
          <div>
            <PiArrowsSplitLight className="h-4 w-4" />
          </div>
          <div className="text-2xl font-semibold text-left">
            Reliable Orchestration
          </div>
        </div>
        <div className="text-md font-normal text-left pl-6">
          Async scheduling, retries, alerting and multuple destinations.
        </div>
        <Image
          src={'/images/orch.svg'}
          alt="st"
          width="443"
          height="333"
          className="border border-gray-300 shadow-lg rounded-xl"
        />
      </div>
      <div className="flex flex-col gap-4 p-10 border-gray-300 border-l-2 w-full">
        <div className="flex flex-row items-center gap-2">
          <div>
            <BsFunnel className="h-4 w-4 " />
          </div>
          <div className="text-2xl font-semibold text-left">
            Powerful Subsetting
          </div>
        </div>
        <div className="text-md font-normal text-left pl-6">
          Subset your DB and Neosync handles the referential integrity.
        </div>
        <Image
          src={'/images/subsetSVG.svg'}
          alt="st"
          width="443"
          height="333"
          className="border border-gray-300 shadow-lg rounded-xl"
        />
      </div>
    </div>
  );
}
