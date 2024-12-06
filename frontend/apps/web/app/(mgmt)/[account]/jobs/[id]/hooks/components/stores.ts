import { create } from 'zustand';
import { EditJobHookFormValues, NewJobHookFormValues } from './validation';

type FieldValues = Record<string, any>; // eslint-disable-line @typescript-eslint/no-explicit-any
type FieldErrors = Record<string, string>; // todo: make this type safe

interface BaseHookStore<T extends FieldValues = FieldValues> {
  formData: T;
  errors: FieldErrors;
  isSubmitting: boolean;
  setFormData(data: Partial<T>): void;
  setErrors(errors: Record<string, string>): void;
  setSubmitting(isSubmitting: boolean): void;
  resetForm(): void;
}

function getInitialEditFormState(): EditJobHookFormValues {
  return {
    hookType: 'sql',
    name: '',
    priority: 0,
    config: {
      sql: { query: '', timing: 'preSync', connectionId: '' },
    },
    description: '',
    enabled: true,
  };
}
interface EditHookStore extends BaseHookStore<EditJobHookFormValues> {}

export const useEditHookStore = create<EditHookStore>((set) => ({
  formData: getInitialEditFormState(),
  errors: {},
  isSubmitting: false,
  setFormData: (data) =>
    set((state) => ({ formData: { ...state.formData, ...data } })),
  setErrors: (errors) => set({ errors }),
  setSubmitting: (isSubmitting) => set({ isSubmitting }),
  resetForm: () =>
    set({
      formData: getInitialEditFormState(),
      errors: {},
      isSubmitting: false,
    }),
}));

function getInitialNewFormState(): NewJobHookFormValues {
  return {
    hookType: 'sql',
    name: '',
    priority: 0,
    config: {
      sql: { query: '', timing: 'preSync', connectionId: '' },
    },
    description: '',
    enabled: true,
  };
}

interface NewHookStore extends BaseHookStore<NewJobHookFormValues> {}

export const useNewHookStore = create<NewHookStore>((set) => ({
  formData: getInitialNewFormState(),
  errors: {},
  isSubmitting: false,
  setFormData: (data) =>
    set((state) => ({ formData: { ...state.formData, ...data } })),
  setErrors: (errors) => set({ errors }),
  setSubmitting: (isSubmitting) => set({ isSubmitting }),
  resetForm: () =>
    set({
      formData: getInitialNewFormState(),
      errors: {},
      isSubmitting: false,
    }),
}));
