import { Button } from '@/components/ui/button';
import { NewJobHook } from '@neosync/sdk';
import { FormEvent, ReactElement } from 'react';
import { ValidationError } from 'yup';
import {
  Description,
  Enabled,
  HookType,
  JobConfig,
  Name,
  Priority,
} from './FormInputs';
import { useNewHookStore } from './stores';
import { newFormDataToNewJobHook, NewJobHookFormValues } from './validation';

interface Props {
  onSubmit(values: NewJobHook): Promise<void>;
  jobConnectionIds: string[];
}

export default function NewHookForm(props: Props): ReactElement {
  const { onSubmit, jobConnectionIds } = props;
  const {
    formData,
    errors,
    setFormData,
    setErrors,
    setSubmitting,
    isSubmitting,
  } = useNewHookStore();

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    if (isSubmitting) {
      return;
    }

    try {
      setSubmitting(true);
      setErrors({});

      const validatedData = await NewJobHookFormValues.validate(formData, {
        abortEarly: false,
      });
      await onSubmit(newFormDataToNewJobHook(validatedData));
    } catch (err) {
      if (err instanceof ValidationError) {
        const validationErrors: Record<string, string> = {};
        err.inner.forEach((error) => {
          if (error.path) {
            validationErrors[error.path] = error.message;
          }
        });
        setErrors(validationErrors);
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <Name
        error={errors.name}
        value={formData.name ?? ''}
        onChange={(value) => setFormData({ name: value })}
      />
      <Description
        error={errors.description}
        value={formData.description ?? ''}
        onChange={(value) => setFormData({ description: value })}
      />
      <Priority
        error={errors.priority}
        value={formData.priority}
        onChange={(value) => setFormData({ priority: value })}
      />
      <Enabled
        error={errors.enabled}
        value={formData.enabled}
        onChange={(value) => setFormData({ enabled: value })}
      />
      <HookType
        error={errors.hookType}
        value={formData.hookType}
        onChange={(value) => setFormData({ hookType: value })}
      />
      <JobConfig
        errors={errors}
        value={formData.config}
        hookType={formData.hookType}
        jobConnectionIds={jobConnectionIds}
        onChange={(value) => setFormData({ config: value })}
      />

      <div className="flex justify-end gap-3">
        <Button
          type="submit"
          disabled={isSubmitting}
          className="w-full sm:w-auto"
        >
          {isSubmitting ? 'Submitting...' : 'Submit'}
        </Button>
      </div>
    </form>
  );
}
