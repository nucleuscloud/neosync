import { createEcmaScriptPlugin } from '@bufbuild/protoplugin';
import { version } from '../package.json';
// import { generateDts } from './declaration.js';
import { generateJs } from './javascript';
import { generateTs } from './typescript';

export const protocGenEs = createEcmaScriptPlugin({
  name: 'protoc-gen-es-zod',
  version: `v${String(version)}`,
  generateTs,
  generateJs,
  // generateDts,
});
