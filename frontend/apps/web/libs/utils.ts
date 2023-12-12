import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// interface DtoClass<T>{
//   new (data: any): DtoClass<T>;

//   fromJson(data: any): DtoClass<T>;

//   // fromJson(jsonValue: JsonValue): T;
// }

// export function hookOnData<T>(data: JsonValue | DtoClass<T>, cl: DtoClass<T>): DtoClass<T> {
//   return data instanceof DtoClass<T> ? data :
// }
export function getRefreshIntervalFn<T>(
  fn?: (data: T) => number
): ((data: T | undefined) => number) | undefined {
  if (!fn) {
    return undefined;
  }
  return (data) => {
    if (!data) {
      return 0;
    }
    return fn(data);
  };
}
