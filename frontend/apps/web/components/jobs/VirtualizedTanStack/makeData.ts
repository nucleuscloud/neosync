import { faker } from '@faker-js/faker';

export type JobMapRow = {
  table: string;
  transformer: {
    value: string;
    config: {};
    // source: string;
  };
  schema: string;
  column: string;
  dataType: string;
  isSelected: boolean;
  formIdx: number;
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

const newRecord = (
  schemas: string[],
  tables: string[],
  idx: number
): JobMapRow => {
  return {
    dataType:
      datatypes[faker.number.int({ min: 0, max: datatypes.length - 1 })],
    transformer: {
      value: 'passthrough',
      config: {},
      // source: 'generate_full_name',
    },
    table: tables[faker.number.int({ min: 0, max: tables.length - 1 })],
    schema: schemas[faker.number.int({ min: 0, max: schemas.length - 1 })],
    column: faker.database.column(),
    isSelected: false,
    formIdx: idx,
  };
};

export function makeData(num: number, schemaCount: number, tableCount: number) {
  // Generate random schema and table names

  const schemas = Array.from({ length: schemaCount }, () =>
    faker.company.name()
  );
  const tables = Array.from({ length: tableCount }, () =>
    faker.company.buzzPhrase().replaceAll(' ', '_')
  );

  const rows: JobMapRow[] = [];

  for (let i = 0; i < num; i++) {
    rows[i] = newRecord(schemas, tables, i);
  }

  return rows;
}
