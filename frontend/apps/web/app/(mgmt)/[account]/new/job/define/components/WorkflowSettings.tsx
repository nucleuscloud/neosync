import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { Input } from '@/components/ui/input';
import { useMemo } from 'react';
import { FieldErrors } from 'react-hook-form';
import {
  ActivityOptionsFormValues,
  NewJobType,
} from '../../job-form-validations';

interface SyncActivityOptionsFormProps {
  value: ActivityOptionsFormValues;
  setValue(value: ActivityOptionsFormValues): void;
  errors?: FieldErrors<ActivityOptionsFormValues>;
  jobtype: NewJobType;
}

export default function SyncActivityOptionsForm({
  value,
  setValue,
  errors,
  jobtype,
}: SyncActivityOptionsFormProps) {
  const { title: startToCloseTitle, description: startToCloseDescription } =
    useStartToCloseTimeoutLabels(jobtype);
  const {
    title: scheduleToCloseTitle,
    description: scheduleToCloseDescription,
  } = useScheduleToCloseTimeoutLabels(jobtype);
  const { title: retryPolicyTitle, description: retryPolicyDescription } =
    useRetryPolicyLabels(jobtype);
  return (
    <div className="flex flex-col gap-6">
      <div className="space-y-2">
        <FormHeader
          htmlFor="startToCloseTimeout"
          title={startToCloseTitle}
          description={startToCloseDescription}
          isRequired={true}
          isErrored={!!errors?.startToCloseTimeout}
        />
        <Input
          type="number"
          value={value.startToCloseTimeout ?? 0}
          onChange={(e) =>
            setValue({
              ...value,
              startToCloseTimeout: e.target.valueAsNumber,
            })
          }
        />
        <FormErrorMessage message={errors?.startToCloseTimeout?.message} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="scheduleToCloseTimeout"
          title={scheduleToCloseTitle}
          description={scheduleToCloseDescription}
          isErrored={!!errors?.scheduleToCloseTimeout}
        />
        <Input
          type="number"
          value={value.scheduleToCloseTimeout ?? 0}
          onChange={(e) =>
            setValue({
              ...value,
              scheduleToCloseTimeout: e.target.valueAsNumber,
            })
          }
        />
        <FormErrorMessage message={errors?.scheduleToCloseTimeout?.message} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="retryPolicy.maximumAttempts"
          title={retryPolicyTitle}
          description={retryPolicyDescription}
          isErrored={!!errors?.retryPolicy?.maximumAttempts}
        />
        <Input
          type="number"
          value={value.retryPolicy?.maximumAttempts ?? 0}
          onChange={(e) =>
            setValue({
              ...value,
              retryPolicy: {
                ...value.retryPolicy,
                maximumAttempts: e.target.valueAsNumber,
              },
            })
          }
        />
        <FormErrorMessage
          message={errors?.retryPolicy?.maximumAttempts?.message}
        />
      </div>
    </div>
  );
}

interface Labels {
  title: string;
  description: string;
}

function useStartToCloseTimeoutLabels(jobtype: NewJobType): Labels {
  return useMemo(() => {
    switch (jobtype) {
      case 'pii-detection':
        return {
          title: 'PII Detection Timeout',
          description:
            'The maximum amount of time (in minutes) a single table detection may run before it times out. This timeout is applied per retry.',
        };
      default:
        return {
          title: 'Table Page Sync Timeout',
          description:
            'The maximum amount of time (in minutes) a single table page synchronization may run before it times out. This may need tuning depending on your datasize and how well optimized the indices on the table. This is applied to every table page sync and generate. This timeout is applied per retry.',
        };
    }
  }, [jobtype]);
}

function useScheduleToCloseTimeoutLabels(jobtype: NewJobType): Labels {
  return useMemo(() => {
    switch (jobtype) {
      case 'pii-detection':
        return {
          title: 'Max PII Detection Timeout including retries',
          description:
            'The total time (in minutes) that a single table detection is allowed to run, including retries. 0 means no timeout.',
        };
      default:
        return {
          title: 'Max Table Timeout including retries',
          description:
            'The total time (in minutes) that a single table sync is allowed to run, including retries. 0 means no timeout.',
        };
    }
  }, [jobtype]);
}

function useRetryPolicyLabels(jobtype: NewJobType): Labels {
  return useMemo(() => {
    switch (jobtype) {
      case 'pii-detection':
        return {
          title: 'Maximum Retry Attempts',
          description:
            'The maximum number of times a table detection may retry before giving up.  If not set or set to 0, it means unlimited retry attempts and the max table timeout including retries will be used to know when to stop.',
        };
      default:
        return {
          title: 'Maximum Retry Attempts',
          description:
            'The maximum number of times a table sync may retry before giving up.  If not set or set to 0, it means unlimited retry attempts and the max table timeout including retries will be used to know when to stop.',
        };
    }
  }, [jobtype]);
}
