import GradientTag from '@/components/GradientTag';
import ShimmeringButton from '@/components/buttons/ShimmeringButton';
import { ArrowRightIcon } from '@radix-ui/react-icons';
import { Metadata } from 'next';
import Image from 'next/image';
import Link from 'next/link';
import { ReactElement } from 'react';
import { DotBackground } from './DotBackground';

export const metadata: Metadata = {
  title: 'About | Neosync',
};

export default function About() {
  return (
    <div className="bg-[#FFFFFF] border-b border-b-gray-200 pt-5">
      <DotBackground>
        <Hero />
      </DotBackground>
      <div className=" bg-[#F5F5F5] px-4">
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 lg:pt-40">
          <TeamSection />
        </div>
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 lg:pt-40">
          <InvestorSection />
        </div>
      </div>
    </div>
  );
}

function Hero(): ReactElement {
  return (
    <div className="flex flex-col gap-6 justify-center z-50 py-20">
      <div className="flex justify-center">
        <Image src="/images/nlogo.svg" alt="logo" width={350} height={350} />
      </div>
      <div className="text-center text-gray-900 font-semibold text-5xl font-satoshi pt-10 bg-white/50">
        The Future is Synthetic Data Engineering
      </div>
      <div className="text-center text-gray-800 font-semibold text-lg font-satoshi mx-60 bg-white/50">
        Synthetic Data Engineering represents the next step in customer data
        security and privacy. Imagine having endless data, at your fingertips,
        without the security, privacy and compliance risk. When we can create
        synthetic data that is structurally and statistically exactly like your
        production data, it opens up a world of possibilities.
      </div>
      <div className="z-50 justify-center flex pt-10">
        <ShimmeringButton>
          <Link
            href="https://www.neosync.dev/blog/synthetic-data-engineering"
            target="_blank"
            className="flex flex-row items-center text-gray-300"
          >
            Read more
            <ArrowRightIcon className="ml-2 w-4 h-4" />
          </Link>
        </ShimmeringButton>
      </div>
    </div>
  );
}

function TeamSection(): ReactElement {
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
          <GradientTag tagValue="Our team" />
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
interface Headshots {
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

function InvestorSection(): ReactElement {
  const investors: Headshots[] = [
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/massimoS.png',
      name: 'Massimo Sgrelli',
      title: 'GP @ LombardStreet Ventures',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/aneel.png',
      name: 'Aneel Ranadive',
      title: 'Managing Partner @ Soma Capital',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/samClayman.png',
      name: 'Sam Clayman',
      title: 'Partner @ Asymmetric Capital Partners',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/paulG.png',
      name: 'Paul Grossinger',
      title: 'GP @ Gaingels',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/namek.png',
      name: `Namek Zu'bi`,
      title: 'GP @ Silicon Valley Badia',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/benLynett.png',
      name: `Ben Lynett`,
      title: 'GP @ Lynett Capital Partners',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/austenAllred.png',
      name: `Austen Allred`,
      title: 'Founder @ Bloomtech',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/joshLewis.png',
      name: `Josh Lewis`,
      title: 'Founder @ Sensible',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/amanKhan.png',
      name: 'Aman Khan',
      title: 'Product @ Arize AI',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/akshat.png',
      name: 'Akshat Agarwal',
      title: 'PM @ Skyflow',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/zachG.png',
      name: 'Zach Ginsburg',
      title: 'GP @ Calm VC',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/adityaN.png',
      name: 'Aditya Naganth',
      title: 'Angel Investor',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/mattS.png',
      name: 'Matt Schulman',
      title: 'CEO/Founder @ Pave',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/shaneF.png',
      name: 'Shane Frykholm',
      title: 'CTO/Co-Founder @ Interview Query',
    },
    {
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/investors/vlad.png',
      name: 'Vlad Blumen',
      title: 'Co-founder @ Alto Pharmacy',
    },
  ];
  return (
    <div className="flex px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto ">
      <div className="flex flex-col text-center pb-[64px] align-middle  mx-2 lg:mx-4">
        <div className="font-satoshi text-4xl text-gray-900 lg:w-8/10 ">
          Our investors
        </div>
        <div className="min-w-full">
          <div className="grid grid-cols-2 lg:grid-cols-5 gap-6 pt-[64px]">
            {investors.map((item) => (
              <InvestorHeadshot
                image={item.image}
                name={item.name}
                title={item.title}
                key={item.name}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function InvestorHeadshot(props: Headshots): ReactElement {
  const { image, name, title } = props;
  return (
    <div className="flex flex-col items-center text-center">
      <Image
        src={image}
        alt="employee_pic"
        className="w-[80px] rounded-xl"
        width="200"
        height="200"
      />
      <div className="text-lg font-satoshi text-gray-900">{name}</div>
      <div className="text-sm font-satoshi text-gray-700 font-normal">
        {title}
      </div>
    </div>
  );
}
