import type { Config } from 'jest';
import nextJest from 'next/jest.js';

// import { getPackageAliases } from '@storybook/nextjs/export-mocks';

import path from 'path';

const createJestConfig = nextJest({
  // Provide the path to your Next.js app to load next.config.js and .env files in your test environment
  dir: './',
});

// const config: Config = {
//   preset: 'ts-jest',
//   // coverageProvider: 'v8',
//   testEnvironment: 'jsdom',
//   verbose: true,
//   // moduleNameMapper: {
//   //   ...pathsToModuleNameMapper(
//   //     {
//   //       '@/*': compilerOptions.paths['@/*'],
//   //     },
//   //     { prefix: '<rootDir>/' }
//   //   ),
//   //   ...pathsToModuleNameMapper(
//   //     {
//   //       '@neosync/sdk': compilerOptions.paths['@neosync/sdk'],
//   //     }
//   //     // { prefix: '../' }
//   //   ),
//   // },
//   // moduleNameMapper: {
//   //   // '^@neosync/sdk/(.*)$': '<rootDir>/../../packages/sdk/$1',
//   //   // '@neosync/sdk': '<rootDir>/../../packages/sdk/$1',
//   //   '@neosync/sdk': '../sdk/ts-client/*',
//   //   '^@neosync/sdk/(.*)$': '../../packages/sdk/$1',
//   //   '^@/(.*)$': '<rootDir>/$1',
//   //   // ...getPackageAliases(),
//   // },
//   // moduleNameMapper: {
//   //   ...pathsToModuleNameMapper(compilerOptions.paths, { prefix: '<rootDir>/' }),
//   //   // ...getPackageAliases,
//   // },
//   moduleNameMapper: {
//     ...pathsToModuleNameMapper(compilerOptions.paths, {
//       prefix: path.join('<rootDir>', '..'),
//     }),
//   },
//   setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],
//   // moduleDirectories: ['node_modules', '<rootDir>'],
//   // roots: ['<rootDir>'],
//   // modulePaths: [compilerOptions.baseUrl],
//   moduleDirectories: ['node_modules', path.join(__dirname, '..')],
//   roots: [
//     '<rootDir>',
//     path.join(__dirname, '..', 'sdk'), // Add this line
//   ],
// };

// const config: Config = {
//   preset: 'ts-jest',
//   coverageProvider: 'v8',
//   testEnvironment: 'jsdom',
//   verbose: true,
//   moduleNameMapper: {
//     ...pathsToModuleNameMapper(compilerOptions.paths, {
//       prefix: path.join('<rootDir>', '..'),
//     }),
//   },
//   setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],
//   moduleDirectories: ['node_modules', path.join(__dirname, '..')],
//   roots: ['<rootDir>', path.join(__dirname, '..', 'sdk', 'ts-client')],
// };

const config: Config = {
  preset: 'ts-jest',
  coverageProvider: 'v8',
  testEnvironment: 'jsdom',
  verbose: true,
  moduleNameMapper: {
    '^@neosync/sdk$': path.join(__dirname, '../../../packages/sdk/src'),
    '^@neosync/sdk/(.*)$': path.join(__dirname, '../../../packages/sdk/src/$1'),
    '^@/(.*)$': '<rootDir>/$1',
  },
  setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],
  moduleDirectories: ['node_modules', path.join(__dirname, '..')],
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
};

export default createJestConfig(config);
