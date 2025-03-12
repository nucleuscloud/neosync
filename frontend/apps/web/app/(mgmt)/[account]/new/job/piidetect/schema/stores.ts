import { BaseHookStore } from '@/util/zustand.stores.util';
import { create } from 'zustand';
import { PiiDetectionSchemaFormValues } from '../../job-form-validations';

function getInitialFormState(): PiiDetectionSchemaFormValues {
  return {
    dataSampling: {
      isEnabled: true,
    },
    tableScanFilter: {
      mode: 'include_all',
      patterns: {
        schemas: [],
        tables: [],
      },
    },
    userPrompt: '',
  };
}

interface PiiDetectionSchemaStore
  extends BaseHookStore<PiiDetectionSchemaFormValues> {}

export const usePiiDetectionSchemaStore = create<PiiDetectionSchemaStore>(
  (set) => ({
    formData: getInitialFormState(),
    errors: {},
    isSubmitting: false,
    setFormData: (data) =>
      set((state) => ({ formData: { ...state.formData, ...data } })),
    setErrors: (errors) => set({ errors }),
    setSubmitting: (isSubmitting) => set({ isSubmitting }),
    resetForm: () =>
      set({
        formData: getInitialFormState(),
        errors: {},
        isSubmitting: false,
      }),
  })
);
