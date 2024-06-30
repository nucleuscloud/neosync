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
  codegenInfo,
  DescEnumValue,
  DescExtension,
  DescField,
  FieldDescriptorProto_Label,
  LongType,
  ScalarType,
  ScalarValue,
} from '@bufbuild/protobuf';
import type { Printable } from '@bufbuild/protoplugin/ecmascript';
import { localName } from '@bufbuild/protoplugin/ecmascript';

export function getFieldTypeInfo(field: DescField | DescExtension): {
  typing: Printable;
  optional: boolean;
  typingInferrableFromZeroValue: boolean;
} {
  const typing: Printable = [];
  let typingInferrableFromZeroValue: boolean;
  let optional = false;
  switch (field.fieldKind) {
    case 'scalar':
      typing.push(scalarTypeScriptType(field.scalar, field.longType));
      optional =
        field.optional ||
        field.proto.label === FieldDescriptorProto_Label.REQUIRED;
      typingInferrableFromZeroValue = true;
      break;
    case 'message': {
      const baseType = codegenInfo.getUnwrappedFieldType(field);
      if (baseType !== undefined) {
        typing.push(scalarTypeScriptType(baseType, LongType.BIGINT));
      } else {
        typing.push({
          kind: 'es_ref_message',
          type: field.message,
          typeOnly: true,
        });
      }
      optional = true;
      typingInferrableFromZeroValue = true;
      break;
    }
    case 'enum':
      typing.push({
        kind: 'es_ref_enum',
        type: field.enum,
        typeOnly: true,
      });
      optional =
        field.optional ||
        field.proto.label === FieldDescriptorProto_Label.REQUIRED;
      typingInferrableFromZeroValue = true;
      break;
    case 'map': {
      let keyType: string;
      switch (field.mapKey) {
        case ScalarType.INT32:
        case ScalarType.FIXED32:
        case ScalarType.UINT32:
        case ScalarType.SFIXED32:
        case ScalarType.SINT32:
          keyType = 'number';
          break;
        default:
          keyType = 'string';
          break;
      }
      let valueType: Printable;
      switch (field.mapValue.kind) {
        case 'scalar':
          valueType = scalarTypeScriptType(
            field.mapValue.scalar,
            LongType.BIGINT
          );
          break;
        case 'message':
          valueType = {
            kind: 'es_ref_message',
            type: field.mapValue.message,
            typeOnly: true,
          };
          break;
        case 'enum':
          valueType = {
            kind: 'es_ref_enum',
            type: field.mapValue.enum,
            typeOnly: true,
          };
          break;
      }
      typing.push('{ [key: ', keyType, ']: ', valueType, ' }');
      typingInferrableFromZeroValue = false;
      optional = false;
      break;
    }
  }
  if (field.repeated) {
    typing.push('[]');
    optional = false;
    typingInferrableFromZeroValue = false;
  }
  return { typing, optional, typingInferrableFromZeroValue };
}

/**
 * Return a printable expression for the default value of a field.
 * Only applicable for singular scalar and enum fields.
 */
export function getFieldDefaultValueExpression(
  field: DescField | DescExtension,
  enumAs:
    | 'enum_value_as_is'
    | 'enum_value_as_integer'
    | 'enum_value_as_cast_integer' = 'enum_value_as_is'
): Printable | undefined {
  if (field.repeated) {
    return undefined;
  }
  if (field.fieldKind !== 'enum' && field.fieldKind !== 'scalar') {
    return undefined;
  }
  const defaultValue = field.getDefaultValue();
  if (defaultValue === undefined) {
    return undefined;
  }
  switch (field.fieldKind) {
    case 'enum': {
      const enumValue = field.enum.values.find(
        (value) => value.number === defaultValue
      );
      if (enumValue === undefined) {
        throw new Error(
          `invalid enum default value: ${String(defaultValue)} for ${enumValue}`
        );
      }
      return literalEnumValue(enumValue, enumAs);
    }
    case 'scalar':
      return literalScalarValue(defaultValue, field);
  }
}

/**
 * Return a printable expression for the zero value of a field.
 *
 * Returns either:
 * - empty array literal for repeated fields
 * - empty object literal for maps
 * - undefined for message fields
 * - an enums first value
 * - scalar zero value
 */
export function getFieldZeroValueExpression(
  field: DescField | DescExtension,
  enumAs:
    | 'enum_value_as_is'
    | 'enum_value_as_integer'
    | 'enum_value_as_cast_integer' = 'enum_value_as_is'
): Printable | undefined {
  if (field.repeated) {
    return '[]';
  }
  switch (field.fieldKind) {
    case 'message':
      return undefined;
    case 'map':
      return '{}';
    case 'enum': {
      // In proto3, the first enum value must be zero.
      // In proto2, protobuf-go returns the first value as the default.
      if (field.enum.values.length < 1) {
        throw new Error('invalid enum: missing at least one value');
      }
      const zeroValue = field.enum.values[0];
      return literalEnumValue(zeroValue, enumAs);
    }
    case 'scalar': {
      const defaultValue = codegenInfo.scalarZeroValue(
        field.scalar,
        field.longType
      );
      return literalScalarValue(defaultValue, field);
    }
  }
}

function literalScalarValue(
  value: ScalarValue,
  field: (DescField | DescExtension) & { fieldKind: 'scalar' }
): Printable {
  switch (field.scalar) {
    case ScalarType.DOUBLE:
    case ScalarType.FLOAT:
    case ScalarType.INT32:
    case ScalarType.FIXED32:
    case ScalarType.UINT32:
    case ScalarType.SFIXED32:
    case ScalarType.SINT32:
      if (typeof value != 'number') {
        throw new Error(
          `Unexpected value for ${ScalarType[field.scalar]} ${field.toString()}: ${String(value)}`
        );
      }
      return value;
    case ScalarType.BOOL:
      if (typeof value != 'boolean') {
        throw new Error(
          `Unexpected value for ${ScalarType[field.scalar]} ${field.toString()}: ${String(value)}`
        );
      }
      return value;
    case ScalarType.STRING:
      if (typeof value != 'string') {
        throw new Error(
          `Unexpected value for ${ScalarType[field.scalar]} ${field.toString()}: ${String(value)}`
        );
      }
      return { kind: 'es_string', value };
    case ScalarType.BYTES:
      if (!(value instanceof Uint8Array)) {
        throw new Error(
          `Unexpected value for ${ScalarType[field.scalar]} ${field.toString()}: ${String(value)}`
        );
      }
      return value;
    case ScalarType.INT64:
    case ScalarType.SINT64:
    case ScalarType.SFIXED64:
    case ScalarType.UINT64:
    case ScalarType.FIXED64:
      if (typeof value != 'bigint' && typeof value != 'string') {
        throw new Error(
          `Unexpected value for ${ScalarType[field.scalar]} ${field.toString()}: ${String(value)}`
        );
      }
      return {
        kind: 'es_proto_int64',
        type: field.scalar,
        longType: field.longType,
        value,
      };
  }
}

function literalEnumValue(
  value: DescEnumValue,
  enumAs:
    | 'enum_value_as_is'
    | 'enum_value_as_integer'
    | 'enum_value_as_cast_integer'
): Printable {
  switch (enumAs) {
    case 'enum_value_as_is':
      return [
        { kind: 'es_ref_enum', type: value.parent, typeOnly: false },
        '.',
        localName(value),
      ];
    case 'enum_value_as_integer':
      return [
        value.number,
        ' /* ',
        value.parent.typeName,
        '.',
        value.name,
        ' */',
      ];
    case 'enum_value_as_cast_integer':
      return [
        value.number,
        ' as ',
        { kind: 'es_ref_enum', type: value.parent, typeOnly: true },
        '.',
        localName(value),
      ];
  }
}

function scalarTypeScriptType(type: ScalarType, longType: LongType): Printable {
  switch (type) {
    case ScalarType.STRING:
      return 'string';
    case ScalarType.BOOL:
      return 'boolean';
    case ScalarType.UINT64:
    case ScalarType.SFIXED64:
    case ScalarType.FIXED64:
    case ScalarType.SINT64:
    case ScalarType.INT64:
      if (longType === LongType.STRING) {
        return 'string';
      }
      return 'bigint';
    case ScalarType.BYTES:
      return 'Uint8Array';
    default:
      return 'number';
  }
}
