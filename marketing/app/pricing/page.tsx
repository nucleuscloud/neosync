'use client';

import CTA from '@/components/cta/CTA';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { cn } from '@/lib/utils';
import { CheckCircle2Icon } from 'lucide-react';
import { ReactElement, useEffect, useState } from 'react';

export default function Pricing() {
  return (
    <div className="flex flex-col gap-6 justify-center z-40 py-20 px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
      <div className="text-center text-gray-900 font-semibold text-3xl lg:text-5xl font-satoshi pt-10 bg-white/50">
        Actually Straightforward Pricing
      </div>
      <div className="text-center text-gray-800 font-semibold text-lg font-satoshi mx-10 md:mx-40 lg:mx-60 xl:mx-80 bg-white/50 pt-6">
        Simple, transparent pricing that is generous and easy to understand.
      </div>
      <div className="flex flex-row items-center justify-center gap-6 pt-10 ">
        <FreePlan />
        <TeamPlan />
        <CustomPlan />
      </div>
      <PricingCalc />
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
    <div className="border-2 border-gray-400 rounded-xl py-8 px-12 mt-28 w-[350px]">
      <div className="flex flex-col gap-6">
        <div className="flex justify-center">
          <Badge variant="outline">Individual</Badge>
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
          <Button variant="default" className="w-full">
            Start for free
          </Button>
        </div>
      </div>
    </div>
  );
}

function TeamPlan(): ReactElement {
  const features = [
    '5M records/mo ($100/1M after)',
    'Unlimited Jobs',
    '5 users ($10/user after)',
    'US Region (EU coming soon!)',
    'Social, SSO (seriously)',
    'Shared Infrastructure',
    'Private Discord/Slack',
  ];

  return (
    <div className="w-[350px]">
      <div className="flex justify-center bg-gradient-to-t from-[#191919] to-[#484848] text-white p-4 shadow-lg rounded-t-xl">
        Most Popular
      </div>
      <div className="border-4 border-gray-800 rounded-b-xl p-8 gap-6 shadow-xl">
        <div className="flex flex-col gap-6">
          <div className="flex justify-center">
            <Badge variant="outline">Team</Badge>
          </div>
          <div className="flex justify-center flex-row gap-2">
            <div className="text-3xl ">$400</div>
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
        <div className="flex justify-center pt-10">
          <Button variant="default" className="w-full">
            Start today
          </Button>
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
    <div className="border-2 border-gray-400 rounded-xl p-8 mt-28 w-[350px]">
      <div className="flex flex-col gap-6">
        <div className="flex justify-center">
          <Badge variant="outline">Enterprise</Badge>
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
            Contact us
          </Button>
        </div>
      </div>
    </div>
  );
}

function PricingCalc(): ReactElement {
  const basePrice = 400;
  const bucketTop = 5000000;
  const ratePerM = 100; // $100 per additional 1M records
  const [inputRecord, setInputRecord] = useState<number>(1000000);
  const [syncFreq, setSyncFreq] = useState<number>(4);
  const [finalEstimate, setFinalEstimate] = useState<number>(400);

  useEffect(() => {
    const additionalRecords: number = inputRecord - bucketTop;
    const additionalPrice: number =
      additionalRecords > 0 ? (additionalRecords / 1000000) * ratePerM : 0;
    const totalPrice: number = basePrice + additionalPrice;
    setFinalEstimate(parseFloat(totalPrice.toFixed(0)));
  }, [inputRecord, syncFreq]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.replace(/,/g, '');
    if (value === '') {
      setInputRecord(0);
    } else {
      const parsedValue = parseFloat(value);
      if (!isNaN(parsedValue)) {
        setInputRecord(parsedValue);
      }
    }
  };
  return (
    <div className=" bg-[#1E1E1E] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 rounded-3xl mt-20 lg:mt-40 py-10 justify-center flex flex-col">
      <div className="mx-40">
        <div className="text-centerfont-semibold text-2xl lg:text-5xl font-satoshi text-white text-center">
          Estimate your monthly price
        </div>
        <div className="text-center text-white font-semibold text-md font-satoshi  pt-6">
          Complete the form below to get an estimate of your monthly cost.
        </div>
        <div>
          <div className="flex flex-col lg:flex-row items-center  p-4 gap-8 rounded-xl mt-10">
            <div className="gap-6 flex flex-col">
              <div className="flex flex-col gap-2">
                <div className="text-white text-sm">
                  Number of records to Sync or Generate
                </div>
                <div className="flex flex-col items-center rounded-lg gap-4 ">
                  <Input
                    id="record_count"
                    type="string"
                    value={
                      inputRecord > 0 ? formatNumberWithCommas(inputRecord) : ''
                    }
                    className="bg-transparent border border-gray-700 text-white w-[300px]"
                    onChange={handleInputChange}
                  />
                </div>
              </div>
              <div className="flex flex-col gap-2">
                <div className="text-white text-sm">Sync frequency</div>
                <div className="flex flex-col items-start rounded-lg gap-4">
                  <ToggleGroup
                    type="single"
                    id="sync_count"
                    className="text-white outline"
                    onValueChange={(e) => setSyncFreq(parseFloat(e))}
                    defaultValue="4"
                  >
                    <ToggleGroupItem
                      value="30"
                      className={cn(
                        syncFreq == 30
                          ? 'bg-gray-700 text-white'
                          : 'border border-gray-700 hover:bg-gray-800'
                      )}
                    >
                      Daily
                    </ToggleGroupItem>
                    <ToggleGroupItem
                      value="4"
                      className={cn(
                        syncFreq == 4
                          ? 'bg-gray-700 text-white'
                          : 'border border-gray-700 hover:bg-gray-800'
                      )}
                    >
                      Weekly
                    </ToggleGroupItem>
                    <ToggleGroupItem
                      value="1"
                      className={cn(
                        syncFreq == 1
                          ? 'bg-gray-700 text-white'
                          : 'border border-gray-700 hover:bg-gray-800'
                      )}
                    >
                      Monthly
                    </ToggleGroupItem>
                  </ToggleGroup>
                </div>
              </div>
            </div>
            <Separator
              orientation="vertical"
              className="bg-gray-600 w-[1px] h-28"
            />
            <div className="border border-gray-600 p-10 rounded-xl">
              <div>
                {' '}
                <div className="flex justify-center flex-row gap-2">
                  <div className="text-3xl text-white ">
                    {formatMoney(finalEstimate)}
                  </div>
                  <div className="text-sm text-white self-end pb-1">/mo</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function FAQs(): ReactElement {
  return (
    <div className="mx-40 mt-20">
      <div className="text-center text-gray-900 font-semibold text-2xl lg:text-5xl font-satoshi pt-10 bg-white/50">
        Frequently Asked Questions
      </div>
      <div className="text-center text-gray-800 font-semibold text-lg font-satoshi  md:mx-10 lg:mx-20 xl:mx-40 bg-white/50 pt-6">
        Answers to our most common pricing questions. If you still have
        questions, chat with us on Discord or schedule time to talk to us here.
      </div>
      <Accordion type="single" collapsible className="w-full pt-10">
        <AccordionItem value="item-1">
          <AccordionTrigger>What is considered a record?</AccordionTrigger>
          <AccordionContent>
            A record is a row of a table that is successfully inserted into the
            final destination. For ex, if you had 10 rows and then subsetted the
            dataset such that only 5 were sent to the destination, you would
            only be charged for 5 rows.
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
            They reset exactly one month after you signed up.{' '}
            <AccordionItem value="item-3">
              <AccordionTrigger>
                When will the EU region be available?
              </AccordionTrigger>
              <AccordionContent>
                We&apos;re working on it and expect to have it in early H2!
              </AccordionContent>
            </AccordionItem>
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-3">
          <AccordionTrigger>
            When will the EU region be available?
          </AccordionTrigger>
          <AccordionContent>
            We&apos;re working on it and expect to have it in early H2!
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
