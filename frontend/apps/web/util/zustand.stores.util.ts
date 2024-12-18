type FieldValues = Record<string, any>; // eslint-disable-line @typescript-eslint/no-explicit-any
type FieldErrors = Record<string, string>; // todo: make this type safe

export interface BaseHookStore<T extends FieldValues = FieldValues> {
  formData: T;
  errors: FieldErrors;
  isSubmitting: boolean;
  setFormData(data: Partial<T>): void;
  setErrors(errors: Record<string, string>): void;
  setSubmitting(isSubmitting: boolean): void;
  resetForm(): void;
}
