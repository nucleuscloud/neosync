import type { Config } from 'jest';
import nextJest from 'next/jest.js';

import { getPackageAliases } from '@storybook/nextjs/export-mocks';
// import { pathsToModuleNameMapper } from 'ts-jest';
// import { compilerOptions } from './tsconfig.json';

import path from 'path';

const createJestConfig = nextJest({
  // Provide the path to your Next.js app to load next.config.js and .env files in your test environment
  dir: './',
});

// console.log(
//   pathsToModuleNameMapper(compilerOptions.paths, {
//     prefix: '<rootDir>/',
//   })
// );
// console.log(__dirname);

const config: Config = {
  preset: 'ts-jest',
  // coverageProvider: 'v8',
  testEnvironment: 'jsdom',
  verbose: true,
  moduleNameMapper: {
    ...getPackageAliases(),
    // ...pathsToModuleNameMapper(compilerOptions.paths, {
    //   prefix: '<rootDir>/',
    // }),
    '^@/(.*)$': '<rootDir>/$1',
    '^@neosync/sdk$': '<rootDir>/../../packages/sdk/$2',
  },
  setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],
  moduleDirectories: ['node_modules'],
  roots: ['<rootDir>', path.join(__dirname, '..', '..', 'packages', 'sdk')],
  moduleFileExtensions: [
    'js',
    'mjs',
    'cjs',
    'jsx',
    'ts',
    'tsx',
    'json',
    'node',
  ],
  modulePathIgnorePatterns: ['<rootDir>/.next/'],
  testPathIgnorePatterns: ['<rootDir>/.next/'],
};

export default createJestConfig(config);
