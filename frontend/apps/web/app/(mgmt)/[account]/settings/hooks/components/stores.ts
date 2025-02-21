import { BaseHookStore } from '@/util/zustand.stores.util';
import { create } from 'zustand';
import {
  EditAccountHookFormValues,
  NewAccountHookFormValues,
} from './validation';

function getInitialEditFormState(): EditAccountHookFormValues {
  return {
    hookType: 'webhook',
    name: '',
    config: {
      webhook: { url: '', secret: '', disableSslVerification: false },
    },
    description: '',
    enabled: true,
  };
}
interface EditHookStore extends BaseHookStore<EditAccountHookFormValues> {}

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

function getInitialNewFormState(): NewAccountHookFormValues {
  return {
    hookType: 'webhook',
    name: '',
    config: {
      webhook: { url: '', secret: '', disableSslVerification: false },
    },
    description: '',
    enabled: true,
  };
}

interface NewHookStore extends BaseHookStore<NewAccountHookFormValues> {}

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
