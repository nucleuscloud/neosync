import Image from 'next/image';
import { ReactElement } from 'react';
import { Headshots } from './Team';

export default function InvestorSection(): ReactElement {
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
          Our Investors
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
