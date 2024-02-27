'use client';

import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { CheckCircle2Icon } from 'lucide-react';

const plans = [
  {
    name: 'Individual',
    description: 'Best for small teams.',
    features: [
      '1 Job',
      'Pre-built Transformers',
      'Unlimited Integrations',
      '3 users',
      'Real time Run Logs',
      'Community Slack Support',
    ],
    lockedFeatures: ['Audit logging', 'Custom Transformers', 'Private Support'],
    price: 'Free',
  },
  {
    name: 'Team',
    description: 'Best for growing teams.',
    features: [
      'All Basic features',
      '3 Jobs',
      'Custom Transformers',
      '10 users',
      'Audit Logging',
      'Private Slack Support',
    ],
    lockedFeatures: ['Unlimited Jobs, Dedicated infrastructure '],
    price: '$400/month',
  },
  {
    name: 'Enterprise',
    description: 'Best for sophisticated teams.',
    features: [
      'All Professional features',
      'Unlimited Jobs',
      'Unlimited Users',
      'SSO',
      'Custom audit requirements',
      'Data residency',
      'Dedicated infrastructure',
    ],
    lockedFeatures: [],
    price: 'Custom',
  },
];

export default function Pricing() {
  return (
    <div className="flex flex-col gap-6 justify-center z-40 py-20">
      <div className="text-center text-gray-900 font-semibold text-3xl lg:text-5xl font-satoshi pt-10 bg-white/50">
        Actually Straightforward Pricing
      </div>
      <div className="text-center text-gray-800 font-semibold text-lg font-satoshi mx-10 md:mx-40 lg:mx-60 xl:mx-80 bg-white/50 max-w-4xl pt-6">
        Simple, transparent pricing that you grows with you. Start for free
        today.
      </div>
      <div className="flex flex-row items-center justify-center gap-6 pt-10">
        {plans.map((item) => (
          <div
            key={item.name}
            className={cn(
              {
                'border-4 ': item.name === 'Team',
                'border-2': item.name !== 'Team',
              },
              'border-gray-800 rounded-xl p-8 transition duration-150 ease-in-out hover:-translate-y-2'
            )}
          >
            <div className="flex flex-col gap-2">
              <div className="flex justify-center">{item.price}</div>
              <div className="flex justify-center">{item.name}</div>
              <div>{item.description}</div>
              <div>
                {item.features.map((feats) => (
                  <div key={feats} className="flex flex-row items center gap-2">
                    <CheckCircle2Icon className="w-4 h-4 bg-green-300" />
                    <div>{feats}</div>
                  </div>
                ))}
              </div>
              <div className="flex justify-center">
                <Button variant="default">Get Started</Button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
