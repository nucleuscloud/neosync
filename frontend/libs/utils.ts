import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function isNil<T>(item: T | null | undefined): item is null | undefined {
  return item == null;
}

export function isNotNil<T>(item: T | null | undefined): item is T {
  return !isNil(item);
}

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
