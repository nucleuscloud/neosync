import { useEffect } from 'react';
import { Control, SetFieldValue, useWatch } from 'react-hook-form';

interface FormPersistConfig {
  storage?: Storage;
  control: Control<any>; // eslint-disable-line @typescript-eslint/no-explicit-any
  setValue: SetFieldValue<any>; // eslint-disable-line @typescript-eslint/no-explicit-any
  exclude?: string[];
  onDataRestored?: (data: any) => void; // eslint-disable-line @typescript-eslint/no-explicit-any
  validate?: boolean;
  dirty?: boolean;
  touch?: boolean;
  onTimeout?: () => void;
  timeout?: number;
}

interface UseFormPersistResult {
  clear(): void;
}

// I copied this from the original react-hook-form-persist npm module so that we could take advantage of the useWatch() hook
// This also lets use nest it in the component tree instead of being a pure hook.
// This allows us to persist values without triggering a wholesale re-render of the entire component hierarchy
export default function useFormPersist(
  name: string,
  {
    storage,
    control,
    setValue,
    exclude = [],
    onDataRestored,
    validate = false,
    dirty = false,
    touch = false,
    onTimeout,
    timeout,
  }: FormPersistConfig
): UseFormPersistResult {
  const watchedValues = useWatch({
    control,
  });

  const getStorage = () => storage || window.sessionStorage;

  const clearStorage = () => getStorage().removeItem(name);

  useEffect(() => {
    const str = getStorage().getItem(name);

    if (str) {
      const { _timestamp = null, ...values } = JSON.parse(str);
      const dataRestored: { [key: string]: any } = {}; // eslint-disable-line @typescript-eslint/no-explicit-any
      const currTimestamp = Date.now();

      if (timeout && currTimestamp - _timestamp > timeout) {
        onTimeout && onTimeout();
        clearStorage();
        return;
      }

      Object.keys(values).forEach((key) => {
        const shouldSet = !exclude.includes(key);
        if (shouldSet) {
          dataRestored[key] = values[key];
          setValue(key, values[key], {
            shouldValidate: validate,
            shouldDirty: dirty,
            shouldTouch: touch,
          });
        }
      });

      if (onDataRestored) {
        onDataRestored(dataRestored);
      }
    }
  }, [storage, name, onDataRestored, setValue]);

  useEffect(() => {
    const values: Record<string, unknown> = exclude.length
      ? Object.entries(watchedValues)
          .filter(([key]) => !exclude.includes(key))
          .reduce((obj, [key, val]) => Object.assign(obj, { [key]: val }), {})
      : Object.assign({}, watchedValues);

    if (Object.entries(values).length) {
      if (timeout !== undefined) {
        values._timestamp = Date.now();
      }
      getStorage().setItem(name, JSON.stringify(values));
    }
  }, [watchedValues, timeout]);

  return {
    clear: () => getStorage().removeItem(name),
  };
}
