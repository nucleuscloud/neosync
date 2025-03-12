import { getConnectionIdFromSource } from '@/app/(mgmt)/[account]/jobs/[id]/source/components/util';
import { BaseHookStore } from '@/util/zustand.stores.util';
import { Job } from '@neosync/sdk';
import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';
import {
  EditPiiDetectionJobFormValues,
  PiiDetectionSchemaFormValues,
  TableScanFilterFormValue,
} from '../../job-form-validations';

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

function getInitialEditFormState(): EditPiiDetectionJobFormValues {
  return {
    ...getInitialFormState(),
    sourceId: '',
  };
}

interface PiiDetectionSchemaStore
  extends BaseHookStore<PiiDetectionSchemaFormValues> {
  sourcedFromRemote: boolean;
  setFromRemoteJob(job: Job): void;
}

const PLACEHOLDER_STORE_PERSIST_KEY = 'pii-detect-schema';

export const usePiiDetectionSchemaStore = create<PiiDetectionSchemaStore>()(
  persist(
    (set, get) => ({
      formData:
        (get()?.formData as PiiDetectionSchemaFormValues) ??
        getInitialFormState(),
      errors: {},
      isSubmitting: false,
      sourcedFromRemote: false,
      setFromRemoteJob: (job) =>
        set({
          formData: getFormStateFromJob(job),
          sourcedFromRemote: true,
          isSubmitting: false,
          errors: {},
        }),
      setFormData: (data) =>
        set((state) => ({ formData: { ...state.formData, ...data } })),
      setErrors: (errors) => set({ errors }),
      setSubmitting: (isSubmitting) => set({ isSubmitting }),
      resetForm: () =>
        set({
          formData: getInitialFormState(),
          errors: {},
          isSubmitting: false,
          sourcedFromRemote: false,
        }),
    }),
    {
      name: PLACEHOLDER_STORE_PERSIST_KEY,
      storage: createJSONStorage(() => sessionStorage),
      partialize: (state) => ({
        formData: state.formData,
      }),
    }
  )
);

// Hack to allow dynamic zustand store persistence keys
// https://github.com/pmndrs/zustand/issues/513
export function setPiiDetectionSchemaStorePersistenceKey(sessionKey: string) {
  usePiiDetectionSchemaStore.persist.setOptions({
    name: sessionKey,
  });
  usePiiDetectionSchemaStore.persist.rehydrate();
  sessionStorage.removeItem(PLACEHOLDER_STORE_PERSIST_KEY);
}

interface EditPiiDetectionSchemaStore
  extends BaseHookStore<EditPiiDetectionJobFormValues> {
  sourcedFromRemote: boolean;
  setFromRemoteJob(job: Job): void;
}

export const useEditPiiDetectionSchemaStore =
  create<EditPiiDetectionSchemaStore>((set) => ({
    formData: getInitialEditFormState(),
    errors: {},
    isSubmitting: false,
    sourcedFromRemote: false,
    setFromRemoteJob: (job) =>
      set({
        formData: getEditFormStateFromJob(job),
        sourcedFromRemote: true,
        isSubmitting: false,
        errors: {},
      }),
    setFormData: (data) =>
      set((state) => ({ formData: { ...state.formData, ...data } })),
    setErrors: (errors) => set({ errors }),
    setSubmitting: (isSubmitting) => set({ isSubmitting }),
    resetForm: () =>
      set({
        formData: getInitialEditFormState(),
        errors: {},
        isSubmitting: false,
        sourcedFromRemote: false,
      }),
  }));

function getEditFormStateFromJob(job: Job): EditPiiDetectionJobFormValues {
  return {
    ...getFormStateFromJob(job),
    sourceId: getConnectionIdFromSource(job.source) ?? '',
  };
}

function getFormStateFromJob(job: Job): PiiDetectionSchemaFormValues {
  if (!job || job.jobType?.jobType.case !== 'piiDetect') {
    return {
      dataSampling: {
        isEnabled: true,
      },
      tableScanFilter: {
        mode: 'include_all',
        patterns: { schemas: [], tables: [] },
      },
      userPrompt: '',
    };
  }

  const jobTypeConfig = job.jobType.jobType.value;

  const tableScanFilter: TableScanFilterFormValue = {
    mode: 'include_all',
    patterns: {
      schemas: [],
      tables: [],
    },
  };

  switch (jobTypeConfig.tableScanFilter?.mode.case) {
    case 'include':
      tableScanFilter.mode = 'include';
      tableScanFilter.patterns = jobTypeConfig.tableScanFilter?.mode.value;
      break;
    case 'exclude':
      tableScanFilter.mode = 'exclude';
      tableScanFilter.patterns = jobTypeConfig.tableScanFilter?.mode.value;
      break;
  }

  return {
    dataSampling: {
      isEnabled: jobTypeConfig.dataSampling?.isEnabled ?? true,
    },
    tableScanFilter: tableScanFilter,
    userPrompt: jobTypeConfig.userPrompt ?? '',
  };
}
