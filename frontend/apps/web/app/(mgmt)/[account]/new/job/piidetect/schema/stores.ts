import { BaseHookStore } from '@/util/zustand.stores.util';
import { Job } from '@neosync/sdk';
import { create } from 'zustand';
import {
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

interface PiiDetectionSchemaStore
  extends BaseHookStore<PiiDetectionSchemaFormValues> {
  sourcedFromRemote: boolean;
  setFromRemoteJob(job: Job): void;
}

export const usePiiDetectionSchemaStore = create<PiiDetectionSchemaStore>(
  (set) => ({
    formData: getInitialFormState(),
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
  })
);

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
