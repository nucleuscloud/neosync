import { ReactElement } from 'react';
import { FieldValues, UseFormReturn } from 'react-hook-form';
import useFormPersist from './useFormPersist';

interface FormPersistProps<T extends FieldValues> {
  form: UseFormReturn<T>;
  formKey: string;
}
const isBrowser = () => typeof window !== 'undefined';

export default function FormPersist<T extends FieldValues>(
  props: FormPersistProps<T>
): ReactElement {
  const { form, formKey } = props;
  useFormPersist(formKey, {
    control: form.control,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });
  return <></>;
}
