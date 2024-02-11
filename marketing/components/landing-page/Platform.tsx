'use client';
import { ReactElement } from 'react';
import { BentoGridSecondDemo } from './bentoPlatform';

export default function Platform(): ReactElement {
  const tabs = [
    {
      name: 'Orchestration',
      key: 'orchestration',
      description:
        'Neosync handles all of the orchestration heavy lifting for you and allows you to move data from a source system to multiple destination systems on any schedule you want. Neosync handles:',
      usecases: ['Scheduling', 'Retries', 'Alerting', 'Destination(s)'],
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/connect-dark.png',
    },
    {
      name: 'Anonymization',
      key: 'anonymization',
      description:
        'Neosync ships with over 35+ Transformers that enable you to anonymize sensitive data. This is great for:',
      usecases: [
        'Custom Transformers',
        'Certain types of machine learning analytics and use-cases',
        'Protecting user data privacy',
        'Following compliance requirements such as HIPAA and GDPR',
      ],
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/customTransformer-dark.png',
    },
    {
      name: 'Synthetic Data',
      key: 'synthetic_data',
      description:
        'Neosync ships with over 35+ Transformers that enable you to create truly privacy-safe synthetic data. This is great for:',
      usecases: [
        'Augmenting an existing database with more data for peformance testing',
        'Creating privacy-safe data that can be shared across regions and environments',
        'Seeding demo environments to show realistic use-cases',
        'Following compliance requirements such as HIPAA and GDPR',
      ],
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/systemTransformer-dark.png',
    },
    {
      name: 'Subsetting',
      key: 'subsetting',
      description:
        '    Neosync allows you to flexibly subset or filter your data using standard SQL syntax. Subsetting is great for:',
      usecases: [
        'Shrinking your production database so that it can fit locally',
        'Filtering data by an ID to replicate a specific view of the data',
        'Debugging data errors',
        'Reducing data transfer costs across environments',
      ],
      image:
        'https://assets.nucleuscloud.com/neosync/marketingsite/subset-dark.png',
    },
  ];

  return (
    <div>
      <div className="text-gray-200 font-semibold text-2xl lg:text-4xl font-satoshi text-center">
        A Modern Platform Built for Teams Who Care About Data Security
      </div>

      <div className=" p-6 lg:p-10 rounded-xl mt-10 ">
        <BentoGridSecondDemo />
        {/* <Tabs defaultValue="orchestration">
          <TabsList className="w-full bg-transparent  ">
            <div className="border border-gray-600 rounded-lg p-2 justify-center flex lg:flex-row flex-col lg:gap-4 ">
              {tabs.map((tab) => (
                <TabsTrigger
                  value={tab.key}
                  className="data-[state=active]:bg-gray-300 data-[state=active]:text-gray-950"
                  key={tab.key}
                  onClick={() => {
                    posthog.capture('tab click', {
                      page: tab.name,
                    });
                  }}
                >
                  {tab.name}
                </TabsTrigger>
              ))}
            </div>
          </TabsList>
          <div>
            {tabs.map((tab) => (
              <TabsContent value={tab.key} className="pt-10" key={tab.key}>
                <div className="flex flex-col justify-center items-center gap-6 lg:gap-10">
                  <div className="flex flex-col gap-2 justify-center">
                    <div className="flex flex-row gap-4 text-gray-300">
                      {tab.usecases.map((item) => (
                        <div
                          className="flex flex-row items-center gap-2"
                          key={item}
                        >
                          <div>
                            <CheckCircledIcon className="min-w-4 min-h-4" />
                          </div>
                          {item}
                        </div>
                      ))}
                    </div>
                  </div>
                  <Image
                    src={tab.image}
                    alt="st"
                    width="1050"
                    height="800"
                    className="rounded-xl border border-gray-700 shadow-xl"
                  />
                </div>
              </TabsContent>
            ))}
          </div>
        </Tabs> */}
      </div>
    </div>
  );
}
