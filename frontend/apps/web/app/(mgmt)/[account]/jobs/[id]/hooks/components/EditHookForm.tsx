import { FormEvent, useEffect } from 'react';
import * as yup from 'yup';

import { Button } from '@/components/ui/button';
import { JobHook } from '@neosync/sdk';
import {
  Description,
  Enabled,
  HookType,
  JobConfig,
  Name,
  Priority,
} from './FormInputs';
import { useEditHookStore } from './stores';
import {
  editFormDataToJobHook,
  EditJobHookFormValues,
  toEditFormData,
} from './validation';

interface EditHookFormProps {
  hook: JobHook;
  onSubmit: (values: JobHook) => Promise<void>;

  jobConnectionIds: string[];
}

export function EditHookForm({
  hook,
  onSubmit,
  jobConnectionIds,
}: EditHookFormProps) {
  const {
    formData,
    errors,
    setFormData,
    setErrors,
    setSubmitting,
    isSubmitting,
  } = useEditHookStore();

  useEffect(() => {
    // Initialize form with hook data
    const formData = toEditFormData(hook);
    setFormData(formData);
  }, [hook, setFormData]);

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    if (isSubmitting) {
      return;
    }

    try {
      setSubmitting(true);
      setErrors({});

      const validatedData = await EditJobHookFormValues.validate(formData, {
        abortEarly: false,
      });

      await onSubmit(editFormDataToJobHook(hook, validatedData));
    } catch (err) {
      if (err instanceof yup.ValidationError) {
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
          {isSubmitting ? 'Saving...' : 'Save Changes'}
        </Button>
      </div>
    </form>
  );
}
