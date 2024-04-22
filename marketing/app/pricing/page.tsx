'use client';
import ShimmeringButton from '@/components/buttons/ShimmeringButton';
import CTA from '@/components/cta/CTA';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { CheckCircle2Icon } from 'lucide-react';
import Link from 'next/link';
import posthog from 'posthog-js';
import { ReactElement } from 'react';

export default function Pricing() {
  return (
    <div className="flex flex-col gap-6 justify-center z-40 py-20  mx-auto">
      <div className="text-center text-gray-900 font-semibold text-3xl lg:text-5xl font-satoshi pt-10 bg-white/50">
        Simple, Transparent Pricing
      </div>
      <div className="text-center text-gray-800 font-semibold text-lg font-satoshi bg-white/50 pt-6">
        Pricing shouldn&apos;t be complicated, so we made it easy.
      </div>
      <div className="flex flex-col lg:flex-row items-center justify-center gap-6 pt-10 ">
        <FreePlan />
        <TeamPlan />
        <CustomPlan />
      </div>
      <FAQs />
      <div className="px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto py-10 lg:py-20">
        <CTA />
      </div>
    </div>
  );
}

function FreePlan(): ReactElement {
  const features = [
    '100k records/month',
    'Unlimited Jobs',
    '1 user',
    'US Region',
    'Social Login',
    'Shared Infrastructure',
    'Community Discord',
  ];

  return (
    <div className="border-2 border-gray-400 rounded-xl py-8 px-12 lg:mt-28 lg:w-[350px] bg-gradient-to-b from-[#ffffff] to-[#f3f3f3]">
      <div className="flex flex-col gap-6">
        <div className="flex justify-center">
          <Badge variant="outline" className="border-gray-400 border">
            Individual
          </Badge>
        </div>
        <div className="flex justify-center flex-row gap-2">
          <div className="text-3xl ">Free</div>
          <div className="text-sm self-end pb-1">/mo</div>
        </div>
        <Separator className="mt-6" />
        <div className="flex flex-col gap-2 pt-6">
          {features.map((item) => (
            <div key={item} className="flex flex-row items center gap-2">
              <CheckCircle2Icon className="w-4 h-4 text-green-800 bg-green-200 rounded-full" />
              <div>{item}</div>
            </div>
          ))}
        </div>
        <div className="flex justify-center pt-6">
          <Button
            variant="default"
            className="w-full"
            onClick={() =>
              posthog.capture('user click', {
                page: 'sign up for free',
              })
            }
          >
            <Link href="https://app.neosync.dev" target="_blank">
              Start for free
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
}

function TeamPlan(): ReactElement {
  const features = [
    '5M records/mo ($60/1M after)',
    'Unlimited Jobs',
    '5 users ($10/user after)',
    'US or EU Region',
    'Social, SSO',
    'Shared Infrastructure',
    'Private Discord/Slack',
  ];

  return (
    <div className="lg:w-[350px] bg-gradient-to-b from-[#ffffff] to-[#f3f3f3]">
      <div className="flex justify-center bg-gradient-to-t from-[#191919] to-[#484848] text-white p-4 shadow-lg rounded-t-xl">
        Most Popular
      </div>
      <div className="border-4 border-gray-800 rounded-b-xl p-8 gap-6 shadow-xl">
        <div className="flex flex-col gap-6">
          <div className="flex justify-center">
            <Badge variant="outline" className="border-gray-400 border">
              Team
            </Badge>
          </div>
          <div className="flex justify-center flex-row gap-2">
            <div className="text-3xl ">$299</div>
            <div className="text-sm self-end pb-1">/mo</div>
          </div>
        </div>
        <Separator className="mt-6" />
        <div className="flex flex-col gap-2 pt-6">
          {features.map((item) => (
            <div key={item} className="flex flex-row items center gap-2">
              <CheckCircle2Icon className="w-4 h-4 text-green-800 bg-green-200 rounded-full" />
              <div>{item}</div>
            </div>
          ))}
        </div>
        <div className="flex justify-center pt-10 w-full">
          <ShimmeringButton
            onClick={() =>
              posthog.capture('user click', {
                page: 'sign up for pro plan',
              })
            }
          >
            <Link
              href="https://calendly.com/evis1/30min"
              className="w-[242px]"
              target="_blank"
            >
              <div className="text-white w-full">Get in touch</div>
            </Link>
          </ShimmeringButton>
        </div>
      </div>
    </div>
  );
}

function CustomPlan(): ReactElement {
  const features = [
    'Unlimited Records',
    'Unlimited Jobs',
    'Unlimited Users',
    'Dedicated Infrastructure',
    'Hybrid Deployment',
    'Social, SSO',
    'Private Discord/Slack',
  ];

  return (
    <div className="border-2 border-gray-400 rounded-xl p-8 lg:mt-28 lg:w-[350px] bg-gradient-to-b from-[#ffffff] to-[#f3f3f3]">
      <div className="flex flex-col gap-6">
        <div className="flex justify-center">
          <Badge variant="outline" className="border-gray-400 border">
            Enterprise
          </Badge>
        </div>
        <div className="flex justify-center flex-row gap-2">
          <div className="text-3xl ">Custom</div>
        </div>
        <Separator className="mt-6" />
        <div className="flex flex-col gap-2 pt-6">
          {features.map((item) => (
            <div key={item} className="flex flex-row items center gap-2">
              <CheckCircle2Icon className="w-4 h-4 text-green-800 bg-green-200 rounded-full" />
              <div>{item}</div>
            </div>
          ))}
        </div>
        <div className="flex justify-center pt-6">
          <Button variant="default" className="w-full">
            <Link
              href="https://calendly.com/evis1/30min"
              className="w-[242px]"
              target="_blank"
            >
              Contact us
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
}

// function PricingCalc(): ReactElement {
//   const basePrice = 400;
//   const bucketTop = 5000000;
//   const ratePerM = 100; // $100 per additional 1M records
//   const [inputRecord, setInputRecord] = useState<number>(1000000);
//   const [syncFreq, setSyncFreq] = useState<number>(4);
//   const [finalEstimate, setFinalEstimate] = useState<number>(400);

//   useEffect(() => {
//     if (inputRecord < 100001) {
//       setFinalEstimate(0);
//     } else {
//       const additionalRecords: number = inputRecord - bucketTop;
//       const additionalPrice: number =
//         additionalRecords > 0 ? (additionalRecords / 1000000) * ratePerM : 0;
//       const totalPrice: number = basePrice + additionalPrice;
//       setFinalEstimate(parseFloat(totalPrice.toFixed(0)));
//     }
//   }, [inputRecord, syncFreq]);

//   const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
//     const value = e.target.value.replace(/,/g, '');
//     if (value === '') {
//       setInputRecord(0);
//     } else {
//       const parsedValue = parseFloat(value);
//       if (!isNaN(parsedValue)) {
//         setInputRecord(parsedValue);
//       }
//     }
//   };
//   return (
//     <div className=" bg-[#1E1E1E]  pt-20 rounded-3xl mt-20 lg:mt-40 py-10 justify-center flex flex-col lg:mx-40">
//       <div className="mx-10 lg:mx-40">
//         <div className="text-center font-semibold text-2xl lg:text-5xl font-satoshi text-white ">
//           Estimate your monthly price
//         </div>
//         <div className="text-center text-white font-semibold text-md font-satoshi pt-6">
//           Complete the form below to get an estimate of your monthly cost.
//         </div>

//         <div className="flex flex-col xl:flex-row items-center justify-center p-4 gap-8 rounded-xl mt-10">
//           <div className="gap-6 flex flex-col">
//             <div className="flex flex-col gap-2">
//               <div className="text-white text-sm">
//                 Number of records to Sync or Generate
//               </div>
//               <div className="flex flex-col items-center rounded-lg gap-4 ">
//                 <Input
//                   id="record_count"
//                   type="string"
//                   value={
//                     inputRecord > 0 ? formatNumberWithCommas(inputRecord) : ''
//                   }
//                   className="bg-transparent border border-gray-700 text-white sm:w-[150px] lg:w-[300px]"
//                   onChange={handleInputChange}
//                 />
//               </div>
//             </div>
//             <div className="flex flex-col gap-2">
//               <div className="text-white text-sm">Sync frequency</div>
//               <div className="flex flex-col items-start rounded-lg gap-4">
//                 <ToggleGroup
//                   type="single"
//                   id="sync_count"
//                   className="text-white outline"
//                   onValueChange={(e) => setSyncFreq(parseFloat(e))}
//                   defaultValue="4"
//                 >
//                   <ToggleGroupItem
//                     value="30"
//                     className={cn(
//                       syncFreq == 30
//                         ? 'bg-gray-700 text-white'
//                         : 'border border-gray-700 hover:bg-gray-800'
//                     )}
//                   >
//                     Daily
//                   </ToggleGroupItem>
//                   <ToggleGroupItem
//                     value="4"
//                     className={cn(
//                       syncFreq == 4
//                         ? 'bg-gray-700 text-white'
//                         : 'border border-gray-700 hover:bg-gray-800'
//                     )}
//                   >
//                     Weekly
//                   </ToggleGroupItem>
//                   <ToggleGroupItem
//                     value="1"
//                     className={cn(
//                       syncFreq == 1
//                         ? 'bg-gray-700 text-white'
//                         : 'border border-gray-700 hover:bg-gray-800'
//                     )}
//                   >
//                     Monthly
//                   </ToggleGroupItem>
//                 </ToggleGroup>
//               </div>
//             </div>
//           </div>
//           <div className="block xl:hidden">
//             <Separator className="bg-gray-600 h-[1px] w-60" />
//           </div>
//           <div className="hidden xl:block">
//             <Separator className="bg-gray-600 w-[1px] h-32" />
//           </div>
//           <div className="border border-gray-600 p-6 lg:p-14 rounded-xl">
//             <div className="flex justify-center flex-row gap-2">
//               <div className="text-3xl text-white ">
//                 {formatMoney(finalEstimate)}
//               </div>
//               <div className="text-sm text-white self-end pb-1">/mo</div>
//             </div>
//           </div>
//         </div>
//       </div>
//     </div>
//   );
// }

function FAQs(): ReactElement {
  return (
    <div className="xl:mx-40 mt-20">
      <div className="text-center text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi pt-10 bg-white/50">
        Frequently Asked Questions
      </div>
      <div className="text-center text-gray-800 font-semibold text-lg font-satoshi bg-white/50 pt-6">
        <span>
          Answers to our most common pricing questions. If you still have
          questions, chat with us on{' '}
        </span>
        <a
          href="https://discord.com/invite/MFAMgnp4HF"
          target="_blank"
          rel="noopener noreferrer"
          className="text-gray-900 underline"
        >
          Discord
        </a>
        <span> or schedule time to talk to us </span>
        <a
          href="https://calendly.com/evis1/30min"
          target="_blank"
          rel="noopener noreferrer"
          className="text-gray-900 underline "
        >
          here.
        </a>
      </div>
      <Accordion type="single" collapsible className="w-full pt-10">
        <AccordionItem value="item-1">
          <AccordionTrigger>What is considered a record?</AccordionTrigger>
          <AccordionContent>
            A record is a row of a table that is successfully inserted into the
            final destination. For ex, if you had 10 rows and then subsetted the
            dataset to 5, you would only be charged for 5 rows. We do not charge
            extra for multiple destinations.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-2">
          <AccordionTrigger>Do you charge by destination? </AccordionTrigger>
          <AccordionContent>
            No. We only charge by record. If you decide to send the same record
            to 3 different databases, we only charge for one record.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-3">
          <AccordionTrigger>When do the record limits reset?</AccordionTrigger>
          <AccordionContent>
            Billing resets on the first of every month. If you sign up after the
            first of the month, your bill and usage will be prorated
            accordingly.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-4">
          <AccordionTrigger>What is a hybrid deployment?</AccordionTrigger>
          <AccordionContent>
            A hybrid deployment is when we you deploy the data plane in your
            infrastructure and we manage the control plane. This is great for
            customers who don&apos;t want their data to leave their
            infrastructure but also want the ease of use of a hosted platform
            and don&apos;t want to host the open source product.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-5">
          <AccordionTrigger>
            How can I get set up in the EU Region?
          </AccordionTrigger>
          <AccordionContent>
            Right now, we&apos;re onboarding folks into the EU region manually.
            Please reach out to us!
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  );
}

function formatMoney(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
  }).format(amount);
}

function formatNumberWithCommas(number: number): string {
  return new Intl.NumberFormat('en-US').format(number);
}
