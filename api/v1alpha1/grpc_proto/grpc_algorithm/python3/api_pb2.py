# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: api.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='api.proto',
  package='api.suggestion',
  syntax='proto3',
  serialized_options=b'Z\024../grpc_algorithm/go',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\tapi.proto\x12\x0e\x61pi.suggestion\"&\n\x08KeyValue\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t\"D\n\x14ParameterAssignments\x12,\n\nkey_values\x18\x01 \x03(\x0b\x32\x18.api.suggestion.KeyValue\"\\\n\x0bTrialResult\x12\x37\n\x15parameter_assignments\x18\x01 \x03(\x0b\x32\x18.api.suggestion.KeyValue\x12\x14\n\x0cobject_value\x18\x02 \x01(\x02\"l\n\rParameterSpec\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x35\n\x0eparameter_type\x18\x02 \x01(\x0e\x32\x1d.api.suggestion.ParameterType\x12\x16\n\x0e\x66\x65\x61sible_space\x18\x03 \x03(\t\"\xbc\x02\n\x0fSamplingRequest\x12\x18\n\x10is_first_request\x18\x01 \x01(\x08\x12\x16\n\x0e\x61lgorithm_name\x18\x02 \x01(\t\x12:\n\x18\x61lgorithm_extra_settings\x18\x03 \x03(\x0b\x32\x18.api.suggestion.KeyValue\x12!\n\x19sampling_number_specified\x18\x04 \x01(\x05\x12\x19\n\x11required_sampling\x18\x06 \x01(\x05\x12\x13\n\x0bis_maximize\x18\x07 \x01(\x08\x12\x35\n\x10\x65xisting_results\x18\x08 \x03(\x0b\x32\x1b.api.suggestion.TrialResult\x12\x31\n\nparameters\x18\t \x03(\x0b\x32\x1d.api.suggestion.ParameterSpec\"Q\n\x10SamplingResponse\x12=\n\x0f\x61ssignments_set\x18\x01 \x03(\x0b\x32$.api.suggestion.ParameterAssignments\"\xda\x01\n\x19SamplingValidationRequest\x12\x16\n\x0e\x61lgorithm_name\x18\x01 \x01(\t\x12:\n\x18\x61lgorithm_extra_settings\x18\x02 \x03(\x0b\x32\x18.api.suggestion.KeyValue\x12!\n\x19sampling_number_specified\x18\x03 \x01(\x05\x12\x13\n\x0bis_maximize\x18\x04 \x01(\x08\x12\x31\n\nparameters\x18\x05 \x03(\x0b\x32\x1d.api.suggestion.ParameterSpec\"\x1c\n\x1aSamplingValidationResponse*U\n\rParameterType\x12\x10\n\x0cUNKNOWN_TYPE\x10\x00\x12\n\n\x06\x44OUBLE\x10\x01\x12\x07\n\x03INT\x10\x02\x12\x0c\n\x08\x44ISCRETE\x10\x03\x12\x0f\n\x0b\x43\x41TEGORICAL\x10\x04\x32\xd5\x01\n\nSuggestion\x12S\n\x0eGetSuggestions\x12\x1f.api.suggestion.SamplingRequest\x1a .api.suggestion.SamplingResponse\x12r\n\x19ValidateAlgorithmSettings\x12).api.suggestion.SamplingValidationRequest\x1a*.api.suggestion.SamplingValidationResponseB\x16Z\x14../grpc_algorithm/gob\x06proto3'
)

_PARAMETERTYPE = _descriptor.EnumDescriptor(
  name='ParameterType',
  full_name='api.suggestion.ParameterType',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='UNKNOWN_TYPE', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='DOUBLE', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='INT', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='DISCRETE', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='CATEGORICAL', index=4, number=4,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=996,
  serialized_end=1081,
)
_sym_db.RegisterEnumDescriptor(_PARAMETERTYPE)

ParameterType = enum_type_wrapper.EnumTypeWrapper(_PARAMETERTYPE)
UNKNOWN_TYPE = 0
DOUBLE = 1
INT = 2
DISCRETE = 3
CATEGORICAL = 4



_KEYVALUE = _descriptor.Descriptor(
  name='KeyValue',
  full_name='api.suggestion.KeyValue',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='api.suggestion.KeyValue.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='value', full_name='api.suggestion.KeyValue.value', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=29,
  serialized_end=67,
)


_PARAMETERASSIGNMENTS = _descriptor.Descriptor(
  name='ParameterAssignments',
  full_name='api.suggestion.ParameterAssignments',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='key_values', full_name='api.suggestion.ParameterAssignments.key_values', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=69,
  serialized_end=137,
)


_TRIALRESULT = _descriptor.Descriptor(
  name='TrialResult',
  full_name='api.suggestion.TrialResult',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='parameter_assignments', full_name='api.suggestion.TrialResult.parameter_assignments', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='object_value', full_name='api.suggestion.TrialResult.object_value', index=1,
      number=2, type=2, cpp_type=6, label=1,
      has_default_value=False, default_value=float(0),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=139,
  serialized_end=231,
)


_PARAMETERSPEC = _descriptor.Descriptor(
  name='ParameterSpec',
  full_name='api.suggestion.ParameterSpec',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='api.suggestion.ParameterSpec.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='parameter_type', full_name='api.suggestion.ParameterSpec.parameter_type', index=1,
      number=2, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='feasible_space', full_name='api.suggestion.ParameterSpec.feasible_space', index=2,
      number=3, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=233,
  serialized_end=341,
)


_SAMPLINGREQUEST = _descriptor.Descriptor(
  name='SamplingRequest',
  full_name='api.suggestion.SamplingRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='is_first_request', full_name='api.suggestion.SamplingRequest.is_first_request', index=0,
      number=1, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='algorithm_name', full_name='api.suggestion.SamplingRequest.algorithm_name', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='algorithm_extra_settings', full_name='api.suggestion.SamplingRequest.algorithm_extra_settings', index=2,
      number=3, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='sampling_number_specified', full_name='api.suggestion.SamplingRequest.sampling_number_specified', index=3,
      number=4, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='required_sampling', full_name='api.suggestion.SamplingRequest.required_sampling', index=4,
      number=6, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='is_maximize', full_name='api.suggestion.SamplingRequest.is_maximize', index=5,
      number=7, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='existing_results', full_name='api.suggestion.SamplingRequest.existing_results', index=6,
      number=8, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='parameters', full_name='api.suggestion.SamplingRequest.parameters', index=7,
      number=9, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=344,
  serialized_end=660,
)


_SAMPLINGRESPONSE = _descriptor.Descriptor(
  name='SamplingResponse',
  full_name='api.suggestion.SamplingResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='assignments_set', full_name='api.suggestion.SamplingResponse.assignments_set', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=662,
  serialized_end=743,
)


_SAMPLINGVALIDATIONREQUEST = _descriptor.Descriptor(
  name='SamplingValidationRequest',
  full_name='api.suggestion.SamplingValidationRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='algorithm_name', full_name='api.suggestion.SamplingValidationRequest.algorithm_name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='algorithm_extra_settings', full_name='api.suggestion.SamplingValidationRequest.algorithm_extra_settings', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='sampling_number_specified', full_name='api.suggestion.SamplingValidationRequest.sampling_number_specified', index=2,
      number=3, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='is_maximize', full_name='api.suggestion.SamplingValidationRequest.is_maximize', index=3,
      number=4, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='parameters', full_name='api.suggestion.SamplingValidationRequest.parameters', index=4,
      number=5, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=746,
  serialized_end=964,
)


_SAMPLINGVALIDATIONRESPONSE = _descriptor.Descriptor(
  name='SamplingValidationResponse',
  full_name='api.suggestion.SamplingValidationResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=966,
  serialized_end=994,
)

_PARAMETERASSIGNMENTS.fields_by_name['key_values'].message_type = _KEYVALUE
_TRIALRESULT.fields_by_name['parameter_assignments'].message_type = _KEYVALUE
_PARAMETERSPEC.fields_by_name['parameter_type'].enum_type = _PARAMETERTYPE
_SAMPLINGREQUEST.fields_by_name['algorithm_extra_settings'].message_type = _KEYVALUE
_SAMPLINGREQUEST.fields_by_name['existing_results'].message_type = _TRIALRESULT
_SAMPLINGREQUEST.fields_by_name['parameters'].message_type = _PARAMETERSPEC
_SAMPLINGRESPONSE.fields_by_name['assignments_set'].message_type = _PARAMETERASSIGNMENTS
_SAMPLINGVALIDATIONREQUEST.fields_by_name['algorithm_extra_settings'].message_type = _KEYVALUE
_SAMPLINGVALIDATIONREQUEST.fields_by_name['parameters'].message_type = _PARAMETERSPEC
DESCRIPTOR.message_types_by_name['KeyValue'] = _KEYVALUE
DESCRIPTOR.message_types_by_name['ParameterAssignments'] = _PARAMETERASSIGNMENTS
DESCRIPTOR.message_types_by_name['TrialResult'] = _TRIALRESULT
DESCRIPTOR.message_types_by_name['ParameterSpec'] = _PARAMETERSPEC
DESCRIPTOR.message_types_by_name['SamplingRequest'] = _SAMPLINGREQUEST
DESCRIPTOR.message_types_by_name['SamplingResponse'] = _SAMPLINGRESPONSE
DESCRIPTOR.message_types_by_name['SamplingValidationRequest'] = _SAMPLINGVALIDATIONREQUEST
DESCRIPTOR.message_types_by_name['SamplingValidationResponse'] = _SAMPLINGVALIDATIONRESPONSE
DESCRIPTOR.enum_types_by_name['ParameterType'] = _PARAMETERTYPE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

KeyValue = _reflection.GeneratedProtocolMessageType('KeyValue', (_message.Message,), {
  'DESCRIPTOR' : _KEYVALUE,
  '__module__' : 'api_pb2'
  # @@protoc_insertion_point(class_scope:api.suggestion.KeyValue)
  })
_sym_db.RegisterMessage(KeyValue)

ParameterAssignments = _reflection.GeneratedProtocolMessageType('ParameterAssignments', (_message.Message,), {
  'DESCRIPTOR' : _PARAMETERASSIGNMENTS,
  '__module__' : 'api_pb2'
  # @@protoc_insertion_point(class_scope:api.suggestion.ParameterAssignments)
  })
_sym_db.RegisterMessage(ParameterAssignments)

TrialResult = _reflection.GeneratedProtocolMessageType('TrialResult', (_message.Message,), {
  'DESCRIPTOR' : _TRIALRESULT,
  '__module__' : 'api_pb2'
  # @@protoc_insertion_point(class_scope:api.suggestion.TrialResult)
  })
_sym_db.RegisterMessage(TrialResult)

ParameterSpec = _reflection.GeneratedProtocolMessageType('ParameterSpec', (_message.Message,), {
  'DESCRIPTOR' : _PARAMETERSPEC,
  '__module__' : 'api_pb2'
  # @@protoc_insertion_point(class_scope:api.suggestion.ParameterSpec)
  })
_sym_db.RegisterMessage(ParameterSpec)

SamplingRequest = _reflection.GeneratedProtocolMessageType('SamplingRequest', (_message.Message,), {
  'DESCRIPTOR' : _SAMPLINGREQUEST,
  '__module__' : 'api_pb2'
  # @@protoc_insertion_point(class_scope:api.suggestion.SamplingRequest)
  })
_sym_db.RegisterMessage(SamplingRequest)

SamplingResponse = _reflection.GeneratedProtocolMessageType('SamplingResponse', (_message.Message,), {
  'DESCRIPTOR' : _SAMPLINGRESPONSE,
  '__module__' : 'api_pb2'
  # @@protoc_insertion_point(class_scope:api.suggestion.SamplingResponse)
  })
_sym_db.RegisterMessage(SamplingResponse)

SamplingValidationRequest = _reflection.GeneratedProtocolMessageType('SamplingValidationRequest', (_message.Message,), {
  'DESCRIPTOR' : _SAMPLINGVALIDATIONREQUEST,
  '__module__' : 'api_pb2'
  # @@protoc_insertion_point(class_scope:api.suggestion.SamplingValidationRequest)
  })
_sym_db.RegisterMessage(SamplingValidationRequest)

SamplingValidationResponse = _reflection.GeneratedProtocolMessageType('SamplingValidationResponse', (_message.Message,), {
  'DESCRIPTOR' : _SAMPLINGVALIDATIONRESPONSE,
  '__module__' : 'api_pb2'
  # @@protoc_insertion_point(class_scope:api.suggestion.SamplingValidationResponse)
  })
_sym_db.RegisterMessage(SamplingValidationResponse)


DESCRIPTOR._options = None

_SUGGESTION = _descriptor.ServiceDescriptor(
  name='Suggestion',
  full_name='api.suggestion.Suggestion',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_start=1084,
  serialized_end=1297,
  methods=[
  _descriptor.MethodDescriptor(
    name='GetSuggestions',
    full_name='api.suggestion.Suggestion.GetSuggestions',
    index=0,
    containing_service=None,
    input_type=_SAMPLINGREQUEST,
    output_type=_SAMPLINGRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='ValidateAlgorithmSettings',
    full_name='api.suggestion.Suggestion.ValidateAlgorithmSettings',
    index=1,
    containing_service=None,
    input_type=_SAMPLINGVALIDATIONREQUEST,
    output_type=_SAMPLINGVALIDATIONRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
])
_sym_db.RegisterServiceDescriptor(_SUGGESTION)

DESCRIPTOR.services_by_name['Suggestion'] = _SUGGESTION

# @@protoc_insertion_point(module_scope)
