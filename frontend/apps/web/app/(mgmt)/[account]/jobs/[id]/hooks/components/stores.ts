import { create } from 'zustand';
import { EditJobHookFormValues, NewJobHookFormValues } from './validation';

interface BaseHookStore<T> {
  formData: EditJobHookFormValues;
  errors: Record<string, string>;
  isSubmitting: boolean;
  setFormData(data: Partial<T>): void;
  setErrors(errors: Record<string, string>): void;
  setSubmitting(isSubmitting: boolean): void;
  resetForm(): void;
}

interface EditHookStore extends BaseHookStore<EditJobHookFormValues> {}

export const useEditHookStore = create<EditHookStore>((set) => ({
  formData: {
    hookType: 'sql',
    name: 'my-initial-job-hook',
    priority: 0,
    sql: { query: 'INITIAL FORM VALUE', timing: 'preSync', connectionId: '' },
    description: '',
    enabled: true,
  },
  errors: {},
  isSubmitting: false,
  setFormData: (data) =>
    set((state) => ({ formData: { ...state.formData, ...data } })),
  setErrors: (errors) => set({ errors }),
  setSubmitting: (isSubmitting) => set({ isSubmitting }),
  resetForm: () =>
    set({
      formData: {
        hookType: 'sql',
        name: 'my-initial-job-hook',
        priority: 0,
        sql: { query: 'RESET FORM VALUE', timing: 'preSync', connectionId: '' },
        description: '',
        enabled: true,
      },
      errors: {},
      isSubmitting: false,
    }),
}));

interface NewHookStore extends BaseHookStore<NewJobHookFormValues> {}

export const useNewHookStore = create<NewHookStore>((set) => ({
  formData: {
    hookType: 'sql',
    name: 'my-initial-job-hook',
    priority: 0,
    sql: { query: 'INITIAL FORM VALUE', timing: 'preSync', connectionId: '' },
    description: '',
    enabled: true,
  },
  errors: {},
  isSubmitting: false,
  setFormData: (data) =>
    set((state) => ({ formData: { ...state.formData, ...data } })),
  setErrors: (errors) => set({ errors }),
  setSubmitting: (isSubmitting) => set({ isSubmitting }),
  resetForm: () =>
    set({
      formData: {
        hookType: 'sql',
        name: 'my-initial-job-hook',
        priority: 0,
        sql: { query: 'RESET FORM VALUE', timing: 'preSync', connectionId: '' },
        description: '',
        enabled: true,
      },
      errors: {},
      isSubmitting: false,
    }),
}));
