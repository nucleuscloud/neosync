import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import { CheckCheckIcon } from 'lucide-react';
import { TbCircleX } from 'react-icons/tb';

export default function FeatureMatrix() {
  const headers = [
    'Features',
    'Open Source',
    'Individual',
    'Team',
    'Enterprise',
  ];
  const features = [
    {
      feature: 'Records',
      description:
        'The number of individual rows you can insert into a destination',
      Plan: ['Unlimited', '100k/month', 'Unlimited', 'Unlimited'],
    },
    {
      feature: 'Jobs',
      description: 'The number of individual jobs that you can configure.',
      Plan: ['Unlimited', 'Unlimited', 'Unlimited', 'Unlimited'],
    },
    {
      feature: 'Users',
      description: 'The number of users you can have.',
      Plan: ['Unlimited', '1 user', 'Unlimited', 'Unlimited'],
    },
    {
      feature: 'Regions',
      description: 'The region that your Neosync instance is deployed to. ',
      Plan: ['Your Region', 'US', 'US/EU', 'Custom'],
    },
    {
      feature: 'Authentication',
      description:
        'The type of authentication that is available to your Neosync instance.',
      Plan: ['Auth/No Auth', 'Social', 'Social/SSO', 'Social/SSO'],
    },
    {
      feature: 'Deployment',
      description: 'The way that your Neosync instance is deployed.',
      Plan: ['Self-deployed', 'Neosync Cloud', 'Neosync Cloud', 'Managed'],
    },
    {
      feature: 'Support',
      description: 'The main support channel for your Neosync account.',
      Plan: [
        'Community Discord',
        'Community Discord',
        'Private Slack',
        'Private Slack',
      ],
    },
    {
      feature: 'Support Type',
      description: 'The type of support for your Neosync account.',
      Plan: ['Community', 'Community', 'Priority', 'Live Support'],
    },
    {
      feature: 'Log retention',
      description: 'The number of days that your logs are retained.',
      Plan: ['no', '5 days', '30 days', 'Custom'],
    },
    {
      feature: 'RBAC (coming soon!)',
      description:
        'Whether or not your Neosync instance has access to Role based access control.',
      Plan: ['no', 'no', 'no', 'yes'],
    },

    {
      feature: 'Audit Logs (coming soon!)',
      description:
        'Whether or not your Neosync instance retains audit logs of user and machine actions.',
      Plan: ['no', 'no', 'no', 'yes'],
    },
    {
      feature: 'Slack Integration (coming soon!)',
      description:
        'A native integration with a Slack channel for notifications and events.',
      Plan: ['no', 'no', 'yes', 'yes'],
    },
    {
      feature: 'Web Hooks (coming soon!)',
      description: 'Access to web-hooks for events.',
      Plan: ['no', 'no', 'yes', 'yes'],
    },
    {
      feature: 'Streaming Mode (coming soon!)',
      description:
        'Whether or not your account has access to Neoysnc in real-time streaming mode.',
      Plan: ['no', 'no', 'yes', 'yes'],
    },
    {
      feature: 'PII Detection (coming soon!)',
      description:
        'Whether or not your account has access to use Neosync for PII detection.',
      Plan: ['no', 'no', 'yes', 'yes'],
    },
  ];

  const handlePlan = (plan: string) => {
    if (plan == 'yes') {
      return <CheckCheckIcon className="w-5 h-5 text-green-500" />;
    } else if (plan == 'no') {
      return <TbCircleX className="w-5 h-5 text-red-500" />;
    } else {
      return <div>{plan}</div>;
    }
  };

  return (
    <div className="w-full max-w-6xl mx-auto py-12 md:py-16 lg:mt-20">
      <div className="grid gap-6 md:gap-8 lg:gap-10">
        <div className="text-center">
          <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold tracking-tight">
            Features by Plan
          </h2>
          <div className="text-center text-gray-800 font-semibold text-lg font-satoshi bg-white/50 pt-6">
            Choose the plan that fits your needs and budget.
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-gray-900">
                {headers.map((header, index) => (
                  <th
                    key={header}
                    className={`px-6 py-4 font-medium text-gray-100 ${
                      index === 0
                        ? 'rounded-l-lg'
                        : index === headers.length - 1
                          ? 'rounded-r-lg'
                          : ''
                    }`}
                  >
                    {header}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {features.map((feat) => (
                <tr
                  key={feat.feature}
                  className="border-b border-gray-200 dark:border-gray-700 bg-white/80"
                >
                  <td className="px-6 py-4 font-medium text-gray-900 dark:text-gray-100">
                    <div className="flex flex-row items-center gap-2">
                      <div>{feat.feature}</div>
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger>
                            <InfoCircledIcon />
                          </TooltipTrigger>
                          <TooltipContent className="bg-gray-800 text-white">
                            <p>{feat.description}</p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                  </td>

                  {feat.Plan.map((plan) => (
                    <td
                      key={plan}
                      className="px-6 py-4 font-medium text-sm text-gray-900 dark:text-gray-100 bg-white/80"
                    >
                      {handlePlan(plan)}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
