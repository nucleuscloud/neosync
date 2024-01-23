import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { CheckCircledIcon } from '@radix-ui/react-icons';
import Image from 'next/image';
import { ReactElement } from 'react';

export default function Platform(): ReactElement {
  return (
    <div>
      <div className="text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi text-center">
        Data Orchestration and Anonymization at Scale
      </div>
      <div className="p-10 border border-gray-400 rounded-xl mt-10 shadow-lg">
        <Tabs defaultValue="orchestration">
          <div className="flex justify-between flex-row items-center">
            <TabsList className=" justify-center">
              <div className="border border-gray-500 rounded-lg p-2 justify-center flex lg:flex-row flex-col gap-4 ">
                <TabsTrigger
                  value="orchestration"
                  className="data-[state=active]:bg-gray-900 data-[state=active]:text-gray-100"
                >
                  Orchestration
                </TabsTrigger>
                <TabsTrigger
                  value="anonymization"
                  className="data-[state=active]:bg-gray-900 data-[state=active]:text-gray-100"
                >
                  Anonymization
                </TabsTrigger>
                <TabsTrigger
                  value="synthetic_data"
                  className="data-[state=active]:bg-gray-900 data-[state=active]:text-gray-100"
                >
                  Synthetic Data
                </TabsTrigger>
                <TabsTrigger
                  value="subsetting"
                  className="data-[state=active]:bg-gray-900 data-[state=active]:text-gray-100"
                >
                  Subsetting
                </TabsTrigger>
              </div>
            </TabsList>
            <Image
              src="https://assets.nucleuscloud.com/neosync/marketingsite/soc2.png"
              alt="soc2"
              className="object-scale-down"
              width="60"
              height="60"
            />
          </div>
          <TabsContent value="orchestration" className="pt-10">
            <div className="flex flex-row justify-start gap-20">
              <div>
                <div className="flex flex-col gap-6">
                  With Neosync, you can move data from a source system to
                  multiple destination systems ad-hoc or on a schedule. Neosync
                  handles:
                </div>
                <div className="pt-10">
                  <CheckMarkList
                    list={[
                      'Scheduling and executing async jobs to move data',
                      'Retries and back-offs',
                      'Alerting and logging when jobs are taking longer than normal or failing',
                      'Syncing across destination types such as from RDS to S3',
                    ]}
                  />
                </div>
              </div>
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/schemaconnect.png"
                alt="st"
                width="650"
                height="400"
                className="rounded-xl border border-gray-400 shadow-xl"
              />
            </div>
          </TabsContent>
          <TabsContent value="anonymization" className="pt-10">
            <div className="flex flex-row justify-start gap-20">
              <div>
                <div className="flex flex-col gap-6">
                  Neosync ships with over 35+ Transformers that enable you to
                  anonymize sensitive data. This is great for:
                </div>
                <div className="pt-10">
                  <CheckMarkList
                    list={[
                      'Creating privacy-safe data that can be shared across regions and environments',
                      'Certain types of machine learning analytics and use-cases',
                      'Protecting user data privacy',
                      'Following compliance requirements such as HIPAA and GDPR',
                    ]}
                  />
                </div>
              </div>
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/customTransformersNew.png"
                alt="st"
                width="650"
                height="400"
                className="rounded-xl border border-gray-400 shadow-xl"
              />
            </div>
          </TabsContent>
          <TabsContent value="synthetic_data" className="pt-10">
            <div className="flex flex-row justify-start gap-20">
              <div>
                <div className="flex flex-col gap-6">
                  Neosync ships with over 35+ Transformers that enable you to
                  create truly privacy-safe synthetic data. This is great for:
                </div>
                <div className="pt-10">
                  <CheckMarkList
                    list={[
                      'Augmenting an existing database with more data for peformance testing',
                      'Creating privacy-safe data that can be shared across regions and environments',
                      'Seeding demo environments to show realistic use-cases',
                      'Following compliance requirements such as HIPAA and GDPR',
                    ]}
                  />
                </div>
              </div>
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/systemTransformers.png"
                alt="st"
                width="650"
                height="400"
                className="rounded-xl border border-gray-400 shadow-xl"
              />
            </div>
          </TabsContent>
          <TabsContent value="subsetting" className="justify-center flex pt-10">
            <div className="flex flex-row justify-start gap-20">
              <div>
                <div className="flex flex-col gap-6">
                  Neosync allows you to flexibly subset or filter your data
                  using standard SQL syntax. Subsetting is great for:
                </div>
                <div className="pt-10">
                  <CheckMarkList
                    list={[
                      'Shrinking your production database so that it can fit locally',
                      'Filtering data by an ID to replicate a specific view of the data',
                      'Debugging data errors',
                      'Reducing data transfer costs across environments',
                    ]}
                  />
                </div>
              </div>
              <Image
                src="https://assets.nucleuscloud.com/neosync/marketingsite/subsets.png"
                alt="st"
                width="650"
                height="400"
                className="rounded-xl border border-gray-400 shadow-xl"
              />
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

interface Props {
  list: string[];
}

function CheckMarkList(props: Props): ReactElement {
  const { list } = props;
  return (
    <div className="flex flex-col gap-4">
      {list.map((item) => (
        <div className="flex flex-row items-center gap-4" key={item}>
          <CheckCircledIcon className="w-4 h-4" /> {item}
        </div>
      ))}
    </div>
  );
}
