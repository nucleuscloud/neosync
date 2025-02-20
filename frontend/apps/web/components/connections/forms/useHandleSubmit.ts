import { BaseStore } from '@/util/zustand.stores.util';
import { FormEvent } from 'react';
import { FieldValues } from 'react-hook-form';
import { ValidationError } from 'yup';

export function useHandleSubmit<T extends FieldValues = FieldValues>(
  store: BaseStore<T>,
  onSubmit: (values: T) => Promise<void>,
  onValidate: (values: T) => Promise<T>
): (e: FormEvent) => Promise<void> {
  const { formData, setErrors, setSubmitting, isSubmitting } = store;

  return async (e: FormEvent) => {
    e.preventDefault();
    if (isSubmitting || !onSubmit) {
      return;
    }

    try {
      setSubmitting(true);
      setErrors({});

      const validatedData = await onValidate(formData);
      await onSubmit(validatedData);
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
  };
}
