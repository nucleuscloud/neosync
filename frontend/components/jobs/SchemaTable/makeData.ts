import { faker } from '@faker-js/faker';

export type Person = {
  table: string;
  transformer: {
    value: string;
    config: {};
  };
  schema: string;
  column: string;
  dataType: string;
  isSelected: boolean;
};

const range = (len: number) => {
  const arr: number[] = [];
  for (let i = 0; i < len; i++) {
    arr.push(i);
  }
  return arr;
};

const newPerson = (index: number): Person => {
  return {
    dataType: faker.database.type(),
    transformer: {
      value: faker.animal.dog(),
      config: {},
    },
    table: faker.person.firstName(),
    schema: faker.person.lastName(),
    column: faker.person.jobTitle(),
    isSelected: false,
  };
};

export function makeData(...lens: number[]) {
  const makeDataLevel = (depth = 0): Person[] => {
    const len = lens[depth]!;
    return range(len).map((d): Person => {
      return {
        ...newPerson(d),
      };
    });
  };

  return makeDataLevel();
}
