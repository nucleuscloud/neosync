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

const datatypes = [
  'varchar',
  'bigint',
  'timestamp',
  'int',
  'text',
  'time',
  'binary',
  'boolean',
  'blob',
  'bit',
  'date',
  'decimal',
  'float',
  'datetime',
  'double',
  'tinyint',
];

const schema = [
  'nucleus',
  'neosync',
  'sample',
  'accounts',
  'man_city',
  'chelsea',
  'newcastle',
  'liverpool',
  'tottenham',
  'real_madrid',
  'barcelona',
  'seville',
  'premier_league',
  'la_liga',
  'arsenal',
  'lyon',
];

const tables = [
  'users',
  'accounts',
  'employees',
  'products',
  'stats',
  'dogs',
  'cats',
  'inventory',
  'sales',
  'people',
  'managers',
  'orders',
  'posts',
  'comments',
  'images',
  'tasks',
  'workflows',
  'admins',
  'queues',
  'overtime',
  'teams',
  'players',
  'staff',
  'leagues',
  'referees',
  'var',
];

const newPerson = (): Person => {
  return {
    dataType:
      datatypes[faker.number.int({ min: 0, max: 100 }) % datatypes.length],
    transformer: {
      value: 'passthrough',
      config: {},
    },
    table: tables[faker.number.int({ min: 0, max: 100 }) % tables.length],
    schema: schema[faker.number.int({ min: 0, max: 100 }) % schema.length],
    column: faker.database.column(),
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
