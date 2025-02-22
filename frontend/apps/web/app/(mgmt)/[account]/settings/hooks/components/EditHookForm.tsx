import { FormEvent, useEffect } from 'react';
import * as yup from 'yup';

import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import { Button } from '@/components/ui/button';
import { AccountHook } from '@neosync/sdk';
import {
  AccountHookConfig,
  AccountHookEvents,
  Description,
  Enabled,
  HookType,
  Name,
} from './FormInputs';
import { useEditHookStore } from './stores';
import {
  EditAccountHookFormValues,
  editFormDataToAccountHook,
  toEditFormData,
} from './validation';

interface EditHookFormProps {
  hook: AccountHook;
  onSubmit: (values: AccountHook) => Promise<void>;
}

export function EditHookForm({ hook, onSubmit }: EditHookFormProps) {
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

      const validatedData = await EditAccountHookFormValues.validate(formData, {
        abortEarly: false,
      });

      await onSubmit(editFormDataToAccountHook(hook, validatedData));
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
      <Enabled
        error={errors.enabled}
        value={formData.enabled}
        onChange={(value) => setFormData({ enabled: value })}
      />
      <AccountHookEvents
        error={errors.events}
        value={formData.events}
        onChange={(value) => setFormData({ events: value })}
      />
      <HookType
        error={errors.hookType}
        value={formData.hookType}
        onChange={(value) => setFormData({ hookType: value })}
      />
      <AccountHookConfig
        errors={errors}
        value={formData.config}
        hookType={formData.hookType}
        onChange={(value) => setFormData({ config: value })}
      />

      <div className="flex justify-end gap-3">
        <Button
          type="submit"
          disabled={isSubmitting}
          className="w-full sm:w-auto"
        >
          <ButtonText
            leftIcon={isSubmitting ? <Spinner /> : undefined}
            text="Update"
          />
        </Button>
      </div>
    </form>
  );
}
