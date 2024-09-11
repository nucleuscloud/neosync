import * as Yup from 'yup';

export const TemporalFormValues = Yup.object({
  namespace: Yup.string().required('The Namespace is required.'),
  syncJobName: Yup.string().required('The Sync Job Name is required.'),
  temporalUrl: Yup.string().required('The Temporal URL is required'),
});

export type TemporalFormValues = Yup.InferType<typeof TemporalFormValues>;
