// Copyright 2021-2024 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {
  DescEnum,
  DescExtension,
  DescField,
  DescMessage,
  FieldDescriptorProto_Label,
  FieldDescriptorProto_Type,
  LongType,
  ScalarType,
  proto2,
  proto3,
} from '@bufbuild/protobuf';
import type {
  GeneratedFile,
  Printable,
  Schema,
} from '@bufbuild/protoplugin/ecmascript';
import { localName, reifyWkt } from '@bufbuild/protoplugin/ecmascript';
import { getNonEditionRuntime } from './editions.js';
import { getFieldDefaultValueExpression } from './utils';

export function generateJs(schema: Schema) {
  for (const file of schema.files) {
    const f = schema.generateFile(file.name + '_zod.js');
    f.preamble(file);
    for (const enumeration of file.enums) {
      generateEnum(schema, f, enumeration);
    }
    for (const message of file.messages) {
      generateMessage(schema, f, message);
    }
    for (const extension of file.extensions) {
      generateExtension(schema, f, extension);
    }
    // We do not generate anything for services
  }
}

// prettier-ignore
function generateEnum(schema: Schema, f: GeneratedFile, enumeration: DescEnum) {
  const protoN = getNonEditionRuntime(schema, enumeration.file);
  f.print(f.jsDoc(enumeration));
  f.print(f.exportDecl("const", enumeration), " = /*@__PURE__*/ ", protoN, ".makeEnum(")
  f.print(`  "`, enumeration.typeName, `",`)
  f.print(`  [`)
  for (const value of enumeration.values) {
    if (localName(value) === value.name) {
      f.print("    {no: ", value.number, ", name: ", f.string(value.name), "},")
    } else {
      f.print("    {no: ", value.number, ", name: ", f.string(value.name), ", localName: ", f.string(localName(value)), "},")
    }
  }
  f.print(`  ],`)
  f.print(");")
  f.print()
}

// prettier-ignore
function generateMessage(schema: Schema, f: GeneratedFile, message: DescMessage) {
  const protoN = getNonEditionRuntime(schema, message.file);
  f.print(f.jsDoc(message));
  f.print(f.exportDecl("const", message), " = /*@__PURE__*/ ", protoN, ".makeMessageType(")
  f.print(`  `, f.string(message.typeName), `,`)
  if (message.fields.length == 0) {
    f.print("  [],")
  } else {
    f.print("  () => [")
    for (const field of message.fields) {
      generateFieldInfo(schema, f, field);
    }
    f.print("  ],")
  }
  const needsLocalName = localName(message) !==
      (message.file.syntax == "proto3" ? proto2 : proto3)
      .makeMessageType(message.typeName, []).name;
  if (needsLocalName) {
    // local name is not inferrable from the type name, we need to provide it
    f.print(`  {localName: `, f.string(localName(message)), `},`)
  }
  f.print(");")
  f.print()
  generateWktMethods(schema, f, message)
  generateWktStaticMethods(schema, f, message)
  for (const nestedEnum of message.nestedEnums) {
    generateEnum(schema, f, nestedEnum);
  }
  for (const nestedMessage of message.nestedMessages) {
    generateMessage(schema, f, nestedMessage);
  }
  for (const nestedExtension of message.nestedExtensions) {
    generateExtension(schema, f, nestedExtension);
  }
}

// prettier-ignore
export function generateFieldInfo(schema: Schema, f: GeneratedFile, field: DescField | DescExtension) {
  f.print("    ", getFieldInfoLiteral(schema, field), ",");
}

// prettier-ignore
export function getFieldInfoLiteral(schema: Schema, field: DescField | DescExtension): Printable {
  const protoN = getNonEditionRuntime(schema, field.kind == "extension" ? field.file : field.parent.file);
  const e: Printable = [];
  e.push("{ no: ", field.number, `, `);
  if (field.kind == "field") {
    e.push(`name: "`, field.name, `", `);
    if (field.jsonName !== undefined) {
      e.push(`jsonName: "`, field.jsonName, `", `);
    }
  }
  switch (field.fieldKind) {
    case "scalar":
      e.push(`kind: "scalar", T: `, field.scalar, ` /* ScalarType.`, ScalarType[field.scalar], ` */, `);
      if (field.longType != LongType.BIGINT) {
        e.push(`L: `, field.longType, ` /* LongType.`, LongType[field.longType], ` */, `);
      }
      break;
    case "map":
      e.push(`kind: "map", K: `, field.mapKey, ` /* ScalarType.`, ScalarType[field.mapKey], ` */, `);
      switch (field.mapValue.kind) {
        case "scalar":
          e.push(`V: {kind: "scalar", T: `, field.mapValue.scalar, ` /* ScalarType.`, ScalarType[field.mapValue.scalar], ` */}, `);
          break;
        case "message":
          e.push(`V: {kind: "message", T: `, field.mapValue.message, `}, `);
          break;
        case "enum":
          e.push(`V: {kind: "enum", T: `, protoN, `.getEnumType(`, field.mapValue.enum, `)}, `);
          break;
      }
      break;
    case "message":
      e.push(`kind: "message", T: `, field.message, `, `);
      if (field.proto.type === FieldDescriptorProto_Type.GROUP) {
        e.push(`delimited: true, `);
      }
      break;
    case "enum":
      e.push(`kind: "enum", T: `, protoN, `.getEnumType(`, field.enum, `), `);
      break;
  }
  if (field.repeated) {
    e.push(`repeated: true, `);
    if (field.packed !== field.packedByDefault) {
      e.push(`packed: `, field.packed, `, `);
    }
  }
  if (field.optional) {
    e.push(`opt: true, `);
  } else if (field.proto.label === FieldDescriptorProto_Label.REQUIRED) {
    e.push(`req: true, `);
  }
  const defaultValue = getFieldDefaultValueExpression(field);
  if (defaultValue !== undefined) {
    e.push(`default: `, defaultValue, `, `);
  }
  if (field.oneof) {
    e.push(`oneof: "`, field.oneof.name, `", `);
  }
  const lastE = e[e.length - 1];
  if (typeof lastE == "string" && lastE.endsWith(", ")) {
    e[e.length - 1] = lastE.substring(0, lastE.length - 2);
  }
  e.push(" }");
  return e;
}

// prettier-ignore
function generateExtension(
  schema: Schema,
  f: GeneratedFile,
  ext: DescExtension,
) {
  const protoN = getNonEditionRuntime(schema, ext.file);
  f.print(f.jsDoc(ext));
  f.print(f.exportDecl("const", ext), " = ", protoN, ".makeExtension(");
  f.print("  ", f.string(ext.typeName), ", ");
  f.print("  ", ext.extendee, ", ");
  if (ext.fieldKind == "scalar") {
    f.print("  ", getFieldInfoLiteral(schema, ext), ",");
  } else {
    f.print("  () => (", getFieldInfoLiteral(schema, ext), "),");
  }
  f.print(");");
  f.print();
}

// prettier-ignore
function generateWktMethods(schema: Schema, f: GeneratedFile, message: DescMessage) {
  const ref = reifyWkt(message);
  if (ref === undefined) {
    return;
  }
  const {
    ScalarType: rtScalarType,
    protoInt64,
  } = schema.runtime;
  const protoN = getNonEditionRuntime(schema, message.file);
  switch (ref.typeName) {
    case "google.protobuf.Any":
      f.print(message, ".prototype.toJson = function toJson(options) {")
      f.print(`  if (this.`, localName(ref.typeUrl), ` === "") {`);
      f.print("    return {};");
      f.print("  }");
      f.print("  const typeName = this.typeUrlToName(this.", localName(ref.typeUrl), ");");
      f.print("  const messageType = options?.typeRegistry?.findMessage(typeName);");
      f.print("  if (!messageType) {");
      f.print("    throw new Error(`cannot encode message ", message.typeName, ' to JSON: "${this.', localName(ref.typeUrl), '}" is not in the type registry`);');
      f.print("  }");
      f.print("  const message = messageType.fromBinary(this.", localName(ref.value), ");");
      f.print("  let json = message.toJson(options);");
      f.print(`  if (typeName.startsWith("google.protobuf.") || (json === null || Array.isArray(json) || typeof json !== "object")) {`);
      f.print("    json = {value: json};");
      f.print("  }");
      f.print(`  json["@type"] = this.`, localName(ref.typeUrl), `;`);
      f.print("  return json;");
      f.print("};");
      f.print();
      f.print(message, ".prototype.fromJson = function fromJson(json, options) {")
      f.print(`  if (json === null || Array.isArray(json) || typeof json != "object") {`);
      f.print("    throw new Error(`cannot decode message ", message.typeName, ' from JSON: expected object but got ${json === null ? "null" : Array.isArray(json) ? "array" : typeof json}`);');
      f.print("  }");
      f.print(`  if (Object.keys(json).length == 0) {`);
      f.print(`    return this;`);
      f.print(`  }`);
      f.print(`  const typeUrl = json["@type"];`);
      f.print(`  if (typeof typeUrl != "string" || typeUrl == "") {`);
      f.print("    throw new Error(`cannot decode message ", message.typeName, ' from JSON: "@type" is empty`);');
      f.print("  }");
      f.print("  const typeName = this.typeUrlToName(typeUrl), messageType = options?.typeRegistry?.findMessage(typeName);");
      f.print("  if (!messageType) {");
      f.print("    throw new Error(`cannot decode message ", message.typeName, " from JSON: ${typeUrl} is not in the type registry`);");
      f.print("  }");
      f.print("  let message;");
      f.print(`  if (typeName.startsWith("google.protobuf.") &&  Object.prototype.hasOwnProperty.call(json, "value")) {`);
      f.print(`    message = messageType.fromJson(json["value"], options);`);
      f.print("  } else {");
      f.print("    const copy = Object.assign({}, json);");
      f.print(`    delete copy["@type"];`);
      f.print("    message = messageType.fromJson(copy, options);");
      f.print("  }");
      f.print("  this.packFrom(message);");
      f.print("  return this;");
      f.print("};");
      f.print();
      f.print(message, ".prototype.packFrom = function packFrom(message) {")
      f.print("  this.", localName(ref.value), " = message.toBinary();");
      f.print("  this.", localName(ref.typeUrl), " = this.typeNameToUrl(message.getType().typeName);");
      f.print("};");
      f.print();
      f.print(message, ".prototype.unpackTo = function unpackTo(target) {")
      f.print("  if (!this.is(target.getType())) {");
      f.print("    return false;");
      f.print("  }");
      f.print("  target.fromBinary(this.", localName(ref.value), ");");
      f.print("  return true;");
      f.print("};");
      f.print();
      f.print(message, ".prototype.unpack = function unpack(registry) {")
      f.print("    if (this.", localName(ref.typeUrl), ` === "") {`);
      f.print("      return undefined;");
      f.print("    }");
      f.print("    const messageType = registry.findMessage(this.typeUrlToName(this.", localName(ref.typeUrl), "));");
      f.print("    if (!messageType) {");
      f.print("      return undefined;");
      f.print("    }");
      f.print("    return messageType.fromBinary(this.", localName(ref.value), ");");
      f.print("  }");
      f.print();
      f.print(message, ".prototype.is = function is(type) {")
      f.print("  if (this.typeUrl === '') {");
      f.print("    return false;");
      f.print("  }");
      f.print("  const name = this.typeUrlToName(this.", localName(ref.typeUrl), ");");
      f.print("  let typeName = '';");
      f.print("  if (typeof type === 'string') {");
      f.print("      typeName = type;");
      f.print("  } else {");
      f.print("      typeName = type.typeName;");
      f.print("  }");
      f.print("  return name === typeName;");
      f.print("};");
      f.print();
      f.print(message, ".prototype.typeNameToUrl = function typeNameToUrl(name) {")
      f.print("  return `type.googleapis.com/${name}`;");
      f.print("};");
      f.print();
      f.print(message, ".prototype.typeUrlToName = function typeUrlToName(url) {")
      f.print("  if (!url.length) {");
      f.print("    throw new Error(`invalid type url: ${url}`);");
      f.print("  }");
      f.print(`  const slash = url.lastIndexOf("/");`);
      f.print("  const name = slash >= 0 ? url.substring(slash + 1) : url;");
      f.print("  if (!name.length) {");
      f.print("    throw new Error(`invalid type url: ${url}`);");
      f.print("  }");
      f.print("  return name;");
      f.print("};");
      f.print();
      break;
    case "google.protobuf.Timestamp":
      f.print(message, ".prototype.fromJson = function fromJson(json, options) {")
      f.print(`  if (typeof json !== "string") {`);
      f.print("    throw new Error(`cannot decode ", message.typeName, " from JSON: ${", protoN, ".json.debug(json)}`);");
      f.print("  }");
      f.print(`  const matches = json.match(/^([0-9]{4})-([0-9]{2})-([0-9]{2})T([0-9]{2}):([0-9]{2}):([0-9]{2})(?:Z|\\.([0-9]{3,9})Z|([+-][0-9][0-9]:[0-9][0-9]))$/);`);
      f.print("  if (!matches) {");
      f.print("    throw new Error(`cannot decode ", message.typeName, " from JSON: invalid RFC 3339 string`);");
      f.print("  }");
      f.print(`  const ms = Date.parse(matches[1] + "-" + matches[2] + "-" + matches[3] + "T" + matches[4] + ":" + matches[5] + ":" + matches[6] + (matches[8] ? matches[8] : "Z"));`);
      f.print("  if (Number.isNaN(ms)) {");
      f.print("    throw new Error(`cannot decode ", message.typeName, " from JSON: invalid RFC 3339 string`);");
      f.print("  }");
      f.print(`  if (ms < Date.parse("0001-01-01T00:00:00Z") || ms > Date.parse("9999-12-31T23:59:59Z")) {`);
      f.print("    throw new Error(`cannot decode message ", message.typeName, " from JSON: must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive`);");
      f.print("  }");
      f.print("  this.", localName(ref.seconds), " = ", protoInt64, ".parse(ms / 1000);");
      f.print("  this.", localName(ref.nanos), " = 0;");
      f.print("  if (matches[7]) {");
      f.print(`    this.`, localName(ref.nanos), ` = (parseInt("1" + matches[7] + "0".repeat(9 - matches[7].length)) - 1000000000);` );
      f.print("  }");
      f.print("  return this;");
      f.print("};");
      f.print();
      f.print(message, ".prototype.toJson = function toJson(options) {")
      f.print("  const ms = Number(this.", localName(ref.seconds), ") * 1000;");
      f.print(`  if (ms < Date.parse("0001-01-01T00:00:00Z") || ms > Date.parse("9999-12-31T23:59:59Z")) {`);
      f.print("    throw new Error(`cannot encode ", message.typeName, " to JSON: must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive`);");
      f.print("  }");
      f.print("  if (this.", localName(ref.nanos), " < 0) {");
      f.print("    throw new Error(`cannot encode ", message.typeName, " to JSON: nanos must not be negative`);");
      f.print("  }");
      f.print(`  let z = "Z";`);
      f.print("  if (this.", localName(ref.nanos), " > 0) {");
      f.print("    const nanosStr = (this.", localName(ref.nanos), " + 1000000000).toString().substring(1);");
      f.print(`    if (nanosStr.substring(3) === "000000") {`);
      f.print(`      z = "." + nanosStr.substring(0, 3) + "Z";`);
      f.print(`    } else if (nanosStr.substring(6) === "000") {`);
      f.print(`      z = "." + nanosStr.substring(0, 6) + "Z";`);
      f.print("    } else {");
      f.print(`      z = "." + nanosStr + "Z";`);
      f.print("    }");
      f.print("  }");
      f.print(`  return new Date(ms).toISOString().replace(".000Z", z);`);
      f.print("};");
      f.print();
      f.print(message, ".prototype.toDate = function toDate() {")
      f.print("  return new Date(Number(this.", localName(ref.seconds), ") * 1000 + Math.ceil(this.", localName(ref.nanos), " / 1000000));");
      f.print("};");
      f.print();
      break;
    case "google.protobuf.Duration":
      f.print(message, ".prototype.fromJson = function fromJson(json, options) {")
      f.print(`  if (typeof json !== "string") {`)
      f.print("    throw new Error(`cannot decode ", message.typeName, " from JSON: ${proto3.json.debug(json)}`);")
      f.print("  }")
      f.print(`  const match = json.match(/^(-?[0-9]+)(?:\\.([0-9]+))?s/);`)
      f.print("  if (match === null) {")
      f.print("    throw new Error(`cannot decode ", message.typeName, " from JSON: ${", protoN, ".json.debug(json)}`);")
      f.print("  }")
      f.print("  const longSeconds = Number(match[1]);")
      f.print("  if (longSeconds > 315576000000 || longSeconds < -315576000000) {")
      f.print("    throw new Error(`cannot decode ", message.typeName, " from JSON: ${", protoN, ".json.debug(json)}`);")
      f.print("  }")
      f.print("  this.", localName(ref.seconds), " = ", protoInt64, ".parse(longSeconds);")
      f.print(`  if (typeof match[2] == "string") {`)
      f.print(`    const nanosStr = match[2] + "0".repeat(9 - match[2].length);`)
      f.print("    this.", localName(ref.nanos), " = parseInt(nanosStr);")
      f.print("    if (longSeconds < 0 || Object.is(longSeconds, -0)) {");
      f.print("      this.", localName(ref.nanos), " = -this.", localName(ref.nanos), ";")
      f.print("    }")
      f.print("  }")
      f.print("  return this;")
      f.print("};");
      f.print()
      f.print(message, ".prototype.toJson = function toJson(options) {")
      f.print("  if (Number(this.", localName(ref.seconds), ") > 315576000000 || Number(this.", localName(ref.seconds), ") < -315576000000) {")
      f.print("    throw new Error(`cannot encode ", message.typeName, " to JSON: value out of range`);")
      f.print("  }")
      f.print("  let text = this.", localName(ref.seconds), ".toString();")
      f.print("  if (this.", localName(ref.nanos), " !== 0) {")
      f.print("    let nanosStr = Math.abs(this.", localName(ref.nanos), ").toString();")
      f.print(`    nanosStr = "0".repeat(9 - nanosStr.length) + nanosStr;`)
      f.print(`    if (nanosStr.substring(3) === "000000") {`)
      f.print("      nanosStr = nanosStr.substring(0, 3);")
      f.print(`    } else if (nanosStr.substring(6) === "000") {`)
      f.print("      nanosStr = nanosStr.substring(0, 6);")
      f.print(`    }`)
      f.print(`    text += "." + nanosStr;`)
      f.print("    if (this.", localName(ref.nanos), " < 0 && this.", localName(ref.seconds), " === ", protoInt64, ".zero) {");
      f.print(`        text = "-" + text;`);
      f.print(`    }`);
      f.print("  }")
      f.print(`  return text + "s";`)
      f.print("};");
      f.print()
      break;
    case "google.protobuf.Struct":
      f.print(message, ".prototype.toJson = function toJson(options) {")
      f.print("  const json = {}")
      f.print("  for (const [k, v] of Object.entries(this.", localName(ref.fields), ")) {")
      f.print("    json[k] = v.toJson(options);")
      f.print("  }")
      f.print("  return json;")
      f.print("};")
      f.print()
      f.print(message, ".prototype.fromJson = function fromJson(json, options) {")
      f.print(`  if (typeof json != "object" || json == null || Array.isArray(json)) {`)
      f.print(`    throw new Error("cannot decode `, message.typeName, ` from JSON " + `, protoN, `.json.debug(json));`)
      f.print("  }")
      f.print("  for (const [k, v] of Object.entries(json)) {")
      f.print("    this.", localName(ref.fields), "[k] = ", ref.fields.mapValue.message ?? "", ".fromJson(v);")
      f.print("  }")
      f.print("  return this;")
      f.print("};");
      f.print()
      break;
    case "google.protobuf.Value":
      f.print(message, ".prototype.toJson = function toJson(options) {")
      f.print("  switch (this.", localName(ref.kind), ".case) {")
      f.print(`    case "`, localName(ref.nullValue), `":`)
      f.print("      return null;")
      f.print(`    case "`, localName(ref.numberValue), `":`)
      f.print(`      if (!Number.isFinite(this.`, localName(ref.kind), `.value)) {`);
      f.print(`          throw new Error("google.protobuf.Value cannot be NaN or Infinity");`);
      f.print(`      }`);
      f.print(`    case "`, localName(ref.boolValue), `":`)
      f.print(`    case "`, localName(ref.stringValue), `":`)
      f.print("      return this.", localName(ref.kind), ".value;")
      f.print(`    case "`, localName(ref.structValue), `":`)
      f.print(`    case "`, localName(ref.listValue), `":`)
      f.print(`      return this.`, localName(ref.kind), `.value.toJson({...options, emitDefaultValues: true});`)
      f.print("  }")
      f.print(`  throw new Error("`, message.typeName, ` must have a value");`)
      f.print("};");
      f.print()
      f.print(message, ".prototype.fromJson = function fromJson(json, options) {")
      f.print("  switch (typeof json) {")
      f.print(`    case "number":`)
      f.print(`      this.kind = { case: "`, localName(ref.numberValue), `", value: json };`)
      f.print("      break;")
      f.print(`    case "string":`)
      f.print(`      this.kind = { case: "`, localName(ref.stringValue), `", value: json };`)
      f.print("      break;")
      f.print(`    case "boolean":`)
      f.print(`      this.kind = { case: "`, localName(ref.boolValue), `", value: json };`)
      f.print("      break;")
      f.print(`    case "object":`)
      f.print("      if (json === null) {")
      f.print(`        this.kind = { case: "`, localName(ref.nullValue), `", value: `, ref.nullValue.enum, `.`, localName(ref.nullValue.enum.values[0]), ` };`)
      f.print("      } else if (Array.isArray(json)) {")
      f.print(`        this.kind = { case: "`, localName(ref.listValue), `", value: `, ref.listValue.message, `.fromJson(json) };`)
      f.print("      } else {")
      f.print(`        this.kind = { case: "`, localName(ref.structValue), `", value: `, ref.structValue.message, `.fromJson(json) };`)
      f.print("      }")
      f.print("      break;")
      f.print("    default:")
      f.print(`      throw new Error("cannot decode `, message.typeName, ` from JSON " + `, protoN, `.json.debug(json));`)
      f.print("  }")
      f.print("  return this;")
      f.print("};");
      f.print()
      break;
    case "google.protobuf.ListValue":
      f.print(message, ".prototype.toJson = function toJson(options) {")
      f.print(`  return this.`, localName(ref.values), `.map(v => v.toJson());`)
      f.print(`}`)
      f.print()
      f.print(message, ".prototype.fromJson = function fromJson(json, options) {")
      f.print(`  if (!Array.isArray(json)) {`)
      f.print(`    throw new Error("cannot decode `, message.typeName, ` from JSON " + `, protoN, `.json.debug(json));`)
      f.print(`  }`)
      f.print(`  for (let e of json) {`)
      f.print(`    this.`, localName(ref.values), `.push(`, ref.values.message, `.fromJson(e));`)
      f.print(`  }`)
      f.print(`  return this;`)
      f.print("};");
      f.print()
      break;
    case "google.protobuf.FieldMask":
      f.print(message, ".prototype.toJson = function toJson(options) {")
      f.print(`  // Converts snake_case to protoCamelCase according to the convention`)
      f.print(`  // used by protoc to convert a field name to a JSON name.`)
      f.print(`  function protoCamelCase(snakeCase) {`)
      f.print(`    let capNext = false;`)
      f.print(`    const b = [];`)
      f.print(`    for (let i = 0; i < snakeCase.length; i++) {`)
      f.print(`      let c = snakeCase.charAt(i);`)
      f.print(`      switch (c) {`)
      f.print(`        case '_':`)
      f.print(`          capNext = true;`)
      f.print(`          break;`)
      f.print(`        case '0':`)
      f.print(`        case '1':`)
      f.print(`        case '2':`)
      f.print(`        case '3':`)
      f.print(`        case '4':`)
      f.print(`        case '5':`)
      f.print(`        case '6':`)
      f.print(`        case '7':`)
      f.print(`        case '8':`)
      f.print(`        case '9':`)
      f.print(`          b.push(c);`)
      f.print(`          capNext = false;`)
      f.print(`          break;`)
      f.print(`        default:`)
      f.print(`          if (capNext) {`)
      f.print(`            capNext = false;`)
      f.print(`            c = c.toUpperCase();`)
      f.print(`          }`)
      f.print(`          b.push(c);`)
      f.print(`          break;`)
      f.print(`      }`)
      f.print(`    }`)
      f.print(`    return b.join('');`)
      f.print(`  }`)
      f.print(`  return this.`, localName(ref.paths), `.map(p => {`)
      f.print(`    if (p.match(/_[0-9]?_/g) || p.match(/[A-Z]/g)) {`)
      f.print(`      throw new Error("cannot encode `, message.typeName, ` to JSON: lowerCamelCase of path name \\"" + p + "\\" is irreversible");`)
      f.print(`    }`)
      f.print(`    return protoCamelCase(p);`)
      f.print(`  }).join(",");`)
      f.print("};");
      f.print()
      f.print(message, ".prototype.fromJson = function fromJson(json, options) {")
      f.print(`  if (typeof json !== "string") {`)
      f.print(`    throw new Error("cannot decode `, message.typeName, ` from JSON: " + proto3.json.debug(json));`)
      f.print(`  }`)
      f.print(`  if (json === "") {`)
      f.print(`    return this;`)
      f.print(`  }`)
      f.print(`  function camelToSnake (str) {`)
      f.print(`    if (str.includes("_")) {`)
      f.print(`      throw new Error("cannot decode `, message.typeName, ` from JSON: path names must be lowerCamelCase");`)
      f.print(`    }`)
      f.print(`    const sc = str.replace(/[A-Z]/g, letter => "_" + letter.toLowerCase());`)
      f.print(`    return (sc[0] === "_") ? sc.substring(1) : sc;`)
      f.print(`  }`)
      f.print(`  this.`, localName(ref.paths), ` = json.split(",").map(camelToSnake);`)
      f.print(`  return this;`)
      f.print("};");
      f.print()
      break;
    case "google.protobuf.DoubleValue":
    case "google.protobuf.FloatValue":
    case "google.protobuf.Int64Value":
    case "google.protobuf.UInt64Value":
    case "google.protobuf.Int32Value":
    case "google.protobuf.UInt32Value":
    case "google.protobuf.BoolValue":
    case "google.protobuf.StringValue":
    case "google.protobuf.BytesValue":
      f.print(message, ".prototype.toJson = function toJson(options) {")
      f.print("  return proto3.json.writeScalar(", rtScalarType, ".", ScalarType[ref.value.scalar], ", this.value, true);")
      f.print("}")
      f.print()
      f.print(message, ".prototype.fromJson = function fromJson(json, options) {")
      f.print("  try {")
      f.print("    this.value = ", protoN, ".json.readScalar(", rtScalarType, ".", ScalarType[ref.value.scalar], ", json);")
      f.print("  } catch (e) {")
      f.print("    let m = `cannot decode message ", message.typeName, " from JSON\"`;")
      f.print("    if (e instanceof Error && e.message.length > 0) {")
      f.print("      m += `: ${e.message}`")
      f.print("    }")
      f.print("    throw new Error(m);")
      f.print("  }")
      f.print("  return this;")
      f.print("};");
      f.print()
      break;
  }
}

// prettier-ignore
function generateWktStaticMethods(schema: Schema, f: GeneratedFile, message: DescMessage) {
  const ref = reifyWkt(message);
  if (ref === undefined) {
    return;
  }
  const {
    protoInt64,
  } = schema.runtime;
  switch (ref.typeName) {
    case "google.protobuf.Any":
      f.print(message, ".pack = function pack(message) {")
      f.print("  const any = new ", message, "();")
      f.print("  any.packFrom(message);")
      f.print("  return any;")
      f.print("};")
      f.print()
      break;
    case "google.protobuf.Timestamp":
      f.print(message, ".now = function now() {")
      f.print("  return ", message, ".fromDate(new Date())")
      f.print("};")
      f.print()
      f.print(message, ".fromDate = function fromDate(date) {")
      f.print("  const ms = date.getTime();")
      f.print("  return new ", message, "({")
      f.print("    ", localName(ref.seconds), ": ", protoInt64, ".parse(Math.floor(ms / 1000)),")
      f.print("    ", localName(ref.nanos), ": (ms % 1000) * 1000000,")
      f.print("  });")
      f.print("};")
      f.print()
      break;
    case "google.protobuf.DoubleValue":
    case "google.protobuf.FloatValue":
    case "google.protobuf.Int64Value":
    case "google.protobuf.UInt64Value":
    case "google.protobuf.Int32Value":
    case "google.protobuf.UInt32Value":
    case "google.protobuf.BoolValue":
    case "google.protobuf.StringValue":
    case "google.protobuf.BytesValue": {
      f.print(message, ".fieldWrapper = {")
      f.print("  wrapField(value) {")
      f.print("    return new ", message, "({value});")
      f.print("  },")
      f.print("  unwrapField(value) {")
      f.print("    return value.", localName(ref.value), ";")
      f.print("  }")
      f.print("};")
      f.print()
      break;
    }
    case "google.protobuf.Duration":
    case "google.protobuf.Struct":
    case "google.protobuf.Value":
    case "google.protobuf.ListValue":
    case "google.protobuf.FieldMask":
      break;
  }
}
