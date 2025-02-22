// @generated by protoc-gen-es v2.2.3 with parameter "target=ts,import_extension=.js"
// @generated from file mgmt/v1alpha1/account_hook.proto (package mgmt.v1alpha1, syntax proto3)
/* eslint-disable */

import type { GenEnum, GenFile, GenMessage, GenService } from "@bufbuild/protobuf/codegenv1";
import { enumDesc, fileDesc, messageDesc, serviceDesc } from "@bufbuild/protobuf/codegenv1";
import { file_buf_validate_validate } from "../../buf/validate/validate_pb.js";
import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { file_google_protobuf_timestamp } from "@bufbuild/protobuf/wkt";
import type { Message } from "@bufbuild/protobuf";

/**
 * Describes the file mgmt/v1alpha1/account_hook.proto.
 */
export const file_mgmt_v1alpha1_account_hook: GenFile = /*@__PURE__*/
  fileDesc("CiBtZ210L3YxYWxwaGExL2FjY291bnRfaG9vay5wcm90bxINbWdtdC52MWFscGhhMSLcAgoLQWNjb3VudEhvb2sSCgoCaWQYASABKAkSDAoEbmFtZRgCIAEoCRITCgtkZXNjcmlwdGlvbhgDIAEoCRISCgphY2NvdW50X2lkGAQgASgJEi8KBmV2ZW50cxgFIAMoDjIfLm1nbXQudjFhbHBoYTEuQWNjb3VudEhvb2tFdmVudBIwCgZjb25maWcYBiABKAsyIC5tZ210LnYxYWxwaGExLkFjY291bnRIb29rQ29uZmlnEhoKEmNyZWF0ZWRfYnlfdXNlcl9pZBgHIAEoCRIuCgpjcmVhdGVkX2F0GAggASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcBIaChJ1cGRhdGVkX2J5X3VzZXJfaWQYCSABKAkSLgoKdXBkYXRlZF9hdBgKIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXASDwoHZW5hYmxlZBgLIAEoCCLdAQoOTmV3QWNjb3VudEhvb2sSJwoEbmFtZRgBIAEoCUIZukgWchQyEl5bYS16MC05LV17MywxMDB9JBIcCgtkZXNjcmlwdGlvbhgCIAEoCUIHukgEcgIQARI5CgZldmVudHMYAyADKA4yHy5tZ210LnYxYWxwaGExLkFjY291bnRIb29rRXZlbnRCCLpIBZIBAggBEjgKBmNvbmZpZxgEIAEoCzIgLm1nbXQudjFhbHBoYTEuQWNjb3VudEhvb2tDb25maWdCBrpIA8gBARIPCgdlbmFibGVkGAUgASgIIr4BChFBY2NvdW50SG9va0NvbmZpZxI7Cgd3ZWJob29rGAEgASgLMigubWdtdC52MWFscGhhMS5BY2NvdW50SG9va0NvbmZpZy5XZWJIb29rSAAaWwoHV2ViSG9vaxIVCgN1cmwYASABKAlCCLpIBXIDiAEBEhcKBnNlY3JldBgCIAEoCUIHukgEcgIQARIgChhkaXNhYmxlX3NzbF92ZXJpZmljYXRpb24YAyABKAhCDwoGY29uZmlnEgW6SAIIASI2ChZHZXRBY2NvdW50SG9va3NSZXF1ZXN0EhwKCmFjY291bnRfaWQYASABKAlCCLpIBXIDsAEBIkQKF0dldEFjY291bnRIb29rc1Jlc3BvbnNlEikKBWhvb2tzGAEgAygLMhoubWdtdC52MWFscGhhMS5BY2NvdW50SG9vayItChVHZXRBY2NvdW50SG9va1JlcXVlc3QSFAoCaWQYASABKAlCCLpIBXIDsAEBIkIKFkdldEFjY291bnRIb29rUmVzcG9uc2USKAoEaG9vaxgBIAEoCzIaLm1nbXQudjFhbHBoYTEuQWNjb3VudEhvb2sibQoYQ3JlYXRlQWNjb3VudEhvb2tSZXF1ZXN0EhwKCmFjY291bnRfaWQYASABKAlCCLpIBXIDsAEBEjMKBGhvb2sYAiABKAsyHS5tZ210LnYxYWxwaGExLk5ld0FjY291bnRIb29rQga6SAPIAQEiRQoZQ3JlYXRlQWNjb3VudEhvb2tSZXNwb25zZRIoCgRob29rGAEgASgLMhoubWdtdC52MWFscGhhMS5BY2NvdW50SG9vayL1AQoYVXBkYXRlQWNjb3VudEhvb2tSZXF1ZXN0EhQKAmlkGAEgASgJQgi6SAVyA7ABARInCgRuYW1lGAIgASgJQhm6SBZyFDISXlthLXowLTktXXszLDEwMH0kEhwKC2Rlc2NyaXB0aW9uGAMgASgJQge6SARyAhABEjkKBmV2ZW50cxgEIAMoDjIfLm1nbXQudjFhbHBoYTEuQWNjb3VudEhvb2tFdmVudEIIukgFkgECCAESMAoGY29uZmlnGAUgASgLMiAubWdtdC52MWFscGhhMS5BY2NvdW50SG9va0NvbmZpZxIPCgdlbmFibGVkGAYgASgIIkUKGVVwZGF0ZUFjY291bnRIb29rUmVzcG9uc2USKAoEaG9vaxgBIAEoCzIaLm1nbXQudjFhbHBoYTEuQWNjb3VudEhvb2siMAoYRGVsZXRlQWNjb3VudEhvb2tSZXF1ZXN0EhQKAmlkGAEgASgJQgi6SAVyA7ABASJFChlEZWxldGVBY2NvdW50SG9va1Jlc3BvbnNlEigKBGhvb2sYASABKAsyGi5tZ210LnYxYWxwaGExLkFjY291bnRIb29rImoKIUlzQWNjb3VudEhvb2tOYW1lQXZhaWxhYmxlUmVxdWVzdBIcCgphY2NvdW50X2lkGAEgASgJQgi6SAVyA7ABARInCgRuYW1lGAIgASgJQhm6SBZyFDISXlthLXowLTktXXszLDEwMH0kIjoKIklzQWNjb3VudEhvb2tOYW1lQXZhaWxhYmxlUmVzcG9uc2USFAoMaXNfYXZhaWxhYmxlGAEgASgIIkUKHFNldEFjY291bnRIb29rRW5hYmxlZFJlcXVlc3QSFAoCaWQYASABKAlCCLpIBXIDsAEBEg8KB2VuYWJsZWQYAiABKAgiSQodU2V0QWNjb3VudEhvb2tFbmFibGVkUmVzcG9uc2USKAoEaG9vaxgBIAEoCzIaLm1nbXQudjFhbHBoYTEuQWNjb3VudEhvb2sicwojR2V0QWN0aXZlQWNjb3VudEhvb2tzQnlFdmVudFJlcXVlc3QSHAoKYWNjb3VudF9pZBgBIAEoCUIIukgFcgOwAQESLgoFZXZlbnQYAiABKA4yHy5tZ210LnYxYWxwaGExLkFjY291bnRIb29rRXZlbnQiUQokR2V0QWN0aXZlQWNjb3VudEhvb2tzQnlFdmVudFJlc3BvbnNlEikKBWhvb2tzGAEgAygLMhoubWdtdC52MWFscGhhMS5BY2NvdW50SG9vayqvAQoQQWNjb3VudEhvb2tFdmVudBIiCh5BQ0NPVU5UX0hPT0tfRVZFTlRfVU5TUEVDSUZJRUQQABImCiJBQ0NPVU5UX0hPT0tfRVZFTlRfSk9CX1JVTl9DUkVBVEVEEAESJQohQUNDT1VOVF9IT09LX0VWRU5UX0pPQl9SVU5fRkFJTEVEEAISKAokQUNDT1VOVF9IT09LX0VWRU5UX0pPQl9SVU5fU1VDQ0VFREVEEAMyqAcKEkFjY291bnRIb29rU2VydmljZRJlCg9HZXRBY2NvdW50SG9va3MSJS5tZ210LnYxYWxwaGExLkdldEFjY291bnRIb29rc1JlcXVlc3QaJi5tZ210LnYxYWxwaGExLkdldEFjY291bnRIb29rc1Jlc3BvbnNlIgOQAgESYgoOR2V0QWNjb3VudEhvb2sSJC5tZ210LnYxYWxwaGExLkdldEFjY291bnRIb29rUmVxdWVzdBolLm1nbXQudjFhbHBoYTEuR2V0QWNjb3VudEhvb2tSZXNwb25zZSIDkAIBEmgKEUNyZWF0ZUFjY291bnRIb29rEicubWdtdC52MWFscGhhMS5DcmVhdGVBY2NvdW50SG9va1JlcXVlc3QaKC5tZ210LnYxYWxwaGExLkNyZWF0ZUFjY291bnRIb29rUmVzcG9uc2UiABJoChFVcGRhdGVBY2NvdW50SG9vaxInLm1nbXQudjFhbHBoYTEuVXBkYXRlQWNjb3VudEhvb2tSZXF1ZXN0GigubWdtdC52MWFscGhhMS5VcGRhdGVBY2NvdW50SG9va1Jlc3BvbnNlIgASaAoRRGVsZXRlQWNjb3VudEhvb2sSJy5tZ210LnYxYWxwaGExLkRlbGV0ZUFjY291bnRIb29rUmVxdWVzdBooLm1nbXQudjFhbHBoYTEuRGVsZXRlQWNjb3VudEhvb2tSZXNwb25zZSIAEoMBChpJc0FjY291bnRIb29rTmFtZUF2YWlsYWJsZRIwLm1nbXQudjFhbHBoYTEuSXNBY2NvdW50SG9va05hbWVBdmFpbGFibGVSZXF1ZXN0GjEubWdtdC52MWFscGhhMS5Jc0FjY291bnRIb29rTmFtZUF2YWlsYWJsZVJlc3BvbnNlIgASdAoVU2V0QWNjb3VudEhvb2tFbmFibGVkEisubWdtdC52MWFscGhhMS5TZXRBY2NvdW50SG9va0VuYWJsZWRSZXF1ZXN0GiwubWdtdC52MWFscGhhMS5TZXRBY2NvdW50SG9va0VuYWJsZWRSZXNwb25zZSIAEowBChxHZXRBY3RpdmVBY2NvdW50SG9va3NCeUV2ZW50EjIubWdtdC52MWFscGhhMS5HZXRBY3RpdmVBY2NvdW50SG9va3NCeUV2ZW50UmVxdWVzdBozLm1nbXQudjFhbHBoYTEuR2V0QWN0aXZlQWNjb3VudEhvb2tzQnlFdmVudFJlc3BvbnNlIgOQAgFCzAEKEWNvbS5tZ210LnYxYWxwaGExQhBBY2NvdW50SG9va1Byb3RvUAFaUGdpdGh1Yi5jb20vbnVjbGV1c2Nsb3VkL25lb3N5bmMvYmFja2VuZC9nZW4vZ28vcHJvdG9zL21nbXQvdjFhbHBoYTE7bWdtdHYxYWxwaGExogIDTVhYqgINTWdtdC5WMWFscGhhMcoCDU1nbXRcVjFhbHBoYTHiAhlNZ210XFYxYWxwaGExXEdQQk1ldGFkYXRh6gIOTWdtdDo6VjFhbHBoYTFiBnByb3RvMw", [file_buf_validate_validate, file_google_protobuf_timestamp]);

/**
 * @generated from message mgmt.v1alpha1.AccountHook
 */
export type AccountHook = Message<"mgmt.v1alpha1.AccountHook"> & {
  /**
   * The unique identifier of this hook.
   *
   * @generated from field: string id = 1;
   */
  id: string;

  /**
   * Name of the hook for display/reference.
   *
   * @generated from field: string name = 2;
   */
  name: string;

  /**
   * Description of what this hook does.
   *
   * @generated from field: string description = 3;
   */
  description: string;

  /**
   * The unique identifier of the account this hook belongs to.
   *
   * @generated from field: string account_id = 4;
   */
  accountId: string;

  /**
   * The events that will trigger this hook.
   *
   * @generated from field: repeated mgmt.v1alpha1.AccountHookEvent events = 5;
   */
  events: AccountHookEvent[];

  /**
   * Hook-type specific configuration.
   *
   * @generated from field: mgmt.v1alpha1.AccountHookConfig config = 6;
   */
  config?: AccountHookConfig;

  /**
   * The user that created this hook.
   *
   * @generated from field: string created_by_user_id = 7;
   */
  createdByUserId: string;

  /**
   * The time this hook was created.
   *
   * @generated from field: google.protobuf.Timestamp created_at = 8;
   */
  createdAt?: Timestamp;

  /**
   * The user that last updated this hook.
   *
   * @generated from field: string updated_by_user_id = 9;
   */
  updatedByUserId: string;

  /**
   * The last time this hook was updated.
   *
   * @generated from field: google.protobuf.Timestamp updated_at = 10;
   */
  updatedAt?: Timestamp;

  /**
   * Whether or not the hook is enabled.
   *
   * @generated from field: bool enabled = 11;
   */
  enabled: boolean;
};

/**
 * Describes the message mgmt.v1alpha1.AccountHook.
 * Use `create(AccountHookSchema)` to create a new message.
 */
export const AccountHookSchema: GenMessage<AccountHook> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 0);

/**
 * @generated from message mgmt.v1alpha1.NewAccountHook
 */
export type NewAccountHook = Message<"mgmt.v1alpha1.NewAccountHook"> & {
  /**
   * Name of the hook for display/reference.
   *
   * @generated from field: string name = 1;
   */
  name: string;

  /**
   * Description of what this hook does.
   *
   * @generated from field: string description = 2;
   */
  description: string;

  /**
   * The events that will trigger this hook.
   *
   * @generated from field: repeated mgmt.v1alpha1.AccountHookEvent events = 3;
   */
  events: AccountHookEvent[];

  /**
   * Hook-type specific configuration.
   *
   * @generated from field: mgmt.v1alpha1.AccountHookConfig config = 4;
   */
  config?: AccountHookConfig;

  /**
   * Whether or not the hook is enabled.
   *
   * @generated from field: bool enabled = 5;
   */
  enabled: boolean;
};

/**
 * Describes the message mgmt.v1alpha1.NewAccountHook.
 * Use `create(NewAccountHookSchema)` to create a new message.
 */
export const NewAccountHookSchema: GenMessage<NewAccountHook> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 1);

/**
 * Hook-specific configuration
 *
 * @generated from message mgmt.v1alpha1.AccountHookConfig
 */
export type AccountHookConfig = Message<"mgmt.v1alpha1.AccountHookConfig"> & {
  /**
   * @generated from oneof mgmt.v1alpha1.AccountHookConfig.config
   */
  config: {
    /**
     * Webhook-based hooks
     *
     * Slack-based hooks
     * SlackHook slack = 2;
     * Future: Discord, Teams, etc.
     *
     * @generated from field: mgmt.v1alpha1.AccountHookConfig.WebHook webhook = 1;
     */
    value: AccountHookConfig_WebHook;
    case: "webhook";
  } | { case: undefined; value?: undefined };
};

/**
 * Describes the message mgmt.v1alpha1.AccountHookConfig.
 * Use `create(AccountHookConfigSchema)` to create a new message.
 */
export const AccountHookConfigSchema: GenMessage<AccountHookConfig> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 2);

/**
 * Webhook-specific configuration
 *
 * @generated from message mgmt.v1alpha1.AccountHookConfig.WebHook
 */
export type AccountHookConfig_WebHook = Message<"mgmt.v1alpha1.AccountHookConfig.WebHook"> & {
  /**
   * The webhook URL to send the event to.
   *
   * @generated from field: string url = 1;
   */
  url: string;

  /**
   * The secret to use for the webhook.
   *
   * @generated from field: string secret = 2;
   */
  secret: string;

  /**
   * Whether to disable SSL verification for the webhook.
   *
   * @generated from field: bool disable_ssl_verification = 3;
   */
  disableSslVerification: boolean;
};

/**
 * Describes the message mgmt.v1alpha1.AccountHookConfig.WebHook.
 * Use `create(AccountHookConfig_WebHookSchema)` to create a new message.
 */
export const AccountHookConfig_WebHookSchema: GenMessage<AccountHookConfig_WebHook> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 2, 0);

/**
 * @generated from message mgmt.v1alpha1.GetAccountHooksRequest
 */
export type GetAccountHooksRequest = Message<"mgmt.v1alpha1.GetAccountHooksRequest"> & {
  /**
   * The account ID to retrieve hooks for.
   *
   * @generated from field: string account_id = 1;
   */
  accountId: string;
};

/**
 * Describes the message mgmt.v1alpha1.GetAccountHooksRequest.
 * Use `create(GetAccountHooksRequestSchema)` to create a new message.
 */
export const GetAccountHooksRequestSchema: GenMessage<GetAccountHooksRequest> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 3);

/**
 * @generated from message mgmt.v1alpha1.GetAccountHooksResponse
 */
export type GetAccountHooksResponse = Message<"mgmt.v1alpha1.GetAccountHooksResponse"> & {
  /**
   * The list of account hooks.
   *
   * @generated from field: repeated mgmt.v1alpha1.AccountHook hooks = 1;
   */
  hooks: AccountHook[];
};

/**
 * Describes the message mgmt.v1alpha1.GetAccountHooksResponse.
 * Use `create(GetAccountHooksResponseSchema)` to create a new message.
 */
export const GetAccountHooksResponseSchema: GenMessage<GetAccountHooksResponse> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 4);

/**
 * @generated from message mgmt.v1alpha1.GetAccountHookRequest
 */
export type GetAccountHookRequest = Message<"mgmt.v1alpha1.GetAccountHookRequest"> & {
  /**
   * The ID of the hook to retrieve.
   *
   * @generated from field: string id = 1;
   */
  id: string;
};

/**
 * Describes the message mgmt.v1alpha1.GetAccountHookRequest.
 * Use `create(GetAccountHookRequestSchema)` to create a new message.
 */
export const GetAccountHookRequestSchema: GenMessage<GetAccountHookRequest> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 5);

/**
 * @generated from message mgmt.v1alpha1.GetAccountHookResponse
 */
export type GetAccountHookResponse = Message<"mgmt.v1alpha1.GetAccountHookResponse"> & {
  /**
   * The account hook.
   *
   * @generated from field: mgmt.v1alpha1.AccountHook hook = 1;
   */
  hook?: AccountHook;
};

/**
 * Describes the message mgmt.v1alpha1.GetAccountHookResponse.
 * Use `create(GetAccountHookResponseSchema)` to create a new message.
 */
export const GetAccountHookResponseSchema: GenMessage<GetAccountHookResponse> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 6);

/**
 * @generated from message mgmt.v1alpha1.CreateAccountHookRequest
 */
export type CreateAccountHookRequest = Message<"mgmt.v1alpha1.CreateAccountHookRequest"> & {
  /**
   * The account ID to create the hook for.
   *
   * @generated from field: string account_id = 1;
   */
  accountId: string;

  /**
   * The new account hook configuration.
   *
   * @generated from field: mgmt.v1alpha1.NewAccountHook hook = 2;
   */
  hook?: NewAccountHook;
};

/**
 * Describes the message mgmt.v1alpha1.CreateAccountHookRequest.
 * Use `create(CreateAccountHookRequestSchema)` to create a new message.
 */
export const CreateAccountHookRequestSchema: GenMessage<CreateAccountHookRequest> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 7);

/**
 * @generated from message mgmt.v1alpha1.CreateAccountHookResponse
 */
export type CreateAccountHookResponse = Message<"mgmt.v1alpha1.CreateAccountHookResponse"> & {
  /**
   * The newly created account hook.
   *
   * @generated from field: mgmt.v1alpha1.AccountHook hook = 1;
   */
  hook?: AccountHook;
};

/**
 * Describes the message mgmt.v1alpha1.CreateAccountHookResponse.
 * Use `create(CreateAccountHookResponseSchema)` to create a new message.
 */
export const CreateAccountHookResponseSchema: GenMessage<CreateAccountHookResponse> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 8);

/**
 * @generated from message mgmt.v1alpha1.UpdateAccountHookRequest
 */
export type UpdateAccountHookRequest = Message<"mgmt.v1alpha1.UpdateAccountHookRequest"> & {
  /**
   * The ID of the hook to update.
   *
   * @generated from field: string id = 1;
   */
  id: string;

  /**
   * Name of the hook for display/reference.
   *
   * @generated from field: string name = 2;
   */
  name: string;

  /**
   * Description of what this hook does.
   *
   * @generated from field: string description = 3;
   */
  description: string;

  /**
   * The events that will trigger this hook.
   *
   * @generated from field: repeated mgmt.v1alpha1.AccountHookEvent events = 4;
   */
  events: AccountHookEvent[];

  /**
   * Hook-type specific configuration.
   *
   * @generated from field: mgmt.v1alpha1.AccountHookConfig config = 5;
   */
  config?: AccountHookConfig;

  /**
   * Whether or not the hook is enabled.
   *
   * @generated from field: bool enabled = 6;
   */
  enabled: boolean;
};

/**
 * Describes the message mgmt.v1alpha1.UpdateAccountHookRequest.
 * Use `create(UpdateAccountHookRequestSchema)` to create a new message.
 */
export const UpdateAccountHookRequestSchema: GenMessage<UpdateAccountHookRequest> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 9);

/**
 * @generated from message mgmt.v1alpha1.UpdateAccountHookResponse
 */
export type UpdateAccountHookResponse = Message<"mgmt.v1alpha1.UpdateAccountHookResponse"> & {
  /**
   * The updated account hook.
   *
   * @generated from field: mgmt.v1alpha1.AccountHook hook = 1;
   */
  hook?: AccountHook;
};

/**
 * Describes the message mgmt.v1alpha1.UpdateAccountHookResponse.
 * Use `create(UpdateAccountHookResponseSchema)` to create a new message.
 */
export const UpdateAccountHookResponseSchema: GenMessage<UpdateAccountHookResponse> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 10);

/**
 * @generated from message mgmt.v1alpha1.DeleteAccountHookRequest
 */
export type DeleteAccountHookRequest = Message<"mgmt.v1alpha1.DeleteAccountHookRequest"> & {
  /**
   * The ID of the hook to delete.
   *
   * @generated from field: string id = 1;
   */
  id: string;
};

/**
 * Describes the message mgmt.v1alpha1.DeleteAccountHookRequest.
 * Use `create(DeleteAccountHookRequestSchema)` to create a new message.
 */
export const DeleteAccountHookRequestSchema: GenMessage<DeleteAccountHookRequest> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 11);

/**
 * @generated from message mgmt.v1alpha1.DeleteAccountHookResponse
 */
export type DeleteAccountHookResponse = Message<"mgmt.v1alpha1.DeleteAccountHookResponse"> & {
  /**
   * The deleted account hook.
   *
   * @generated from field: mgmt.v1alpha1.AccountHook hook = 1;
   */
  hook?: AccountHook;
};

/**
 * Describes the message mgmt.v1alpha1.DeleteAccountHookResponse.
 * Use `create(DeleteAccountHookResponseSchema)` to create a new message.
 */
export const DeleteAccountHookResponseSchema: GenMessage<DeleteAccountHookResponse> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 12);

/**
 * @generated from message mgmt.v1alpha1.IsAccountHookNameAvailableRequest
 */
export type IsAccountHookNameAvailableRequest = Message<"mgmt.v1alpha1.IsAccountHookNameAvailableRequest"> & {
  /**
   * The account ID to check the name for.
   *
   * @generated from field: string account_id = 1;
   */
  accountId: string;

  /**
   * The name to check.
   *
   * @generated from field: string name = 2;
   */
  name: string;
};

/**
 * Describes the message mgmt.v1alpha1.IsAccountHookNameAvailableRequest.
 * Use `create(IsAccountHookNameAvailableRequestSchema)` to create a new message.
 */
export const IsAccountHookNameAvailableRequestSchema: GenMessage<IsAccountHookNameAvailableRequest> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 13);

/**
 * @generated from message mgmt.v1alpha1.IsAccountHookNameAvailableResponse
 */
export type IsAccountHookNameAvailableResponse = Message<"mgmt.v1alpha1.IsAccountHookNameAvailableResponse"> & {
  /**
   * Whether the name is available.
   *
   * @generated from field: bool is_available = 1;
   */
  isAvailable: boolean;
};

/**
 * Describes the message mgmt.v1alpha1.IsAccountHookNameAvailableResponse.
 * Use `create(IsAccountHookNameAvailableResponseSchema)` to create a new message.
 */
export const IsAccountHookNameAvailableResponseSchema: GenMessage<IsAccountHookNameAvailableResponse> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 14);

/**
 * @generated from message mgmt.v1alpha1.SetAccountHookEnabledRequest
 */
export type SetAccountHookEnabledRequest = Message<"mgmt.v1alpha1.SetAccountHookEnabledRequest"> & {
  /**
   * The ID of the hook to enable/disable.
   *
   * @generated from field: string id = 1;
   */
  id: string;

  /**
   * Whether to enable or disable the hook.
   *
   * @generated from field: bool enabled = 2;
   */
  enabled: boolean;
};

/**
 * Describes the message mgmt.v1alpha1.SetAccountHookEnabledRequest.
 * Use `create(SetAccountHookEnabledRequestSchema)` to create a new message.
 */
export const SetAccountHookEnabledRequestSchema: GenMessage<SetAccountHookEnabledRequest> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 15);

/**
 * @generated from message mgmt.v1alpha1.SetAccountHookEnabledResponse
 */
export type SetAccountHookEnabledResponse = Message<"mgmt.v1alpha1.SetAccountHookEnabledResponse"> & {
  /**
   * The updated account hook.
   *
   * @generated from field: mgmt.v1alpha1.AccountHook hook = 1;
   */
  hook?: AccountHook;
};

/**
 * Describes the message mgmt.v1alpha1.SetAccountHookEnabledResponse.
 * Use `create(SetAccountHookEnabledResponseSchema)` to create a new message.
 */
export const SetAccountHookEnabledResponseSchema: GenMessage<SetAccountHookEnabledResponse> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 16);

/**
 * @generated from message mgmt.v1alpha1.GetActiveAccountHooksByEventRequest
 */
export type GetActiveAccountHooksByEventRequest = Message<"mgmt.v1alpha1.GetActiveAccountHooksByEventRequest"> & {
  /**
   * The account ID to retrieve hooks for.
   *
   * @generated from field: string account_id = 1;
   */
  accountId: string;

  /**
   * The event to retrieve hooks for.
   * A specific event will return hooks that are listening to that specific event as well as wildcard hooks.
   * If you want to retrieve only wildcard hooks, use ACCOUNT_HOOK_EVENT_UNSPECIFIED.
   *
   * @generated from field: mgmt.v1alpha1.AccountHookEvent event = 2;
   */
  event: AccountHookEvent;
};

/**
 * Describes the message mgmt.v1alpha1.GetActiveAccountHooksByEventRequest.
 * Use `create(GetActiveAccountHooksByEventRequestSchema)` to create a new message.
 */
export const GetActiveAccountHooksByEventRequestSchema: GenMessage<GetActiveAccountHooksByEventRequest> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 17);

/**
 * @generated from message mgmt.v1alpha1.GetActiveAccountHooksByEventResponse
 */
export type GetActiveAccountHooksByEventResponse = Message<"mgmt.v1alpha1.GetActiveAccountHooksByEventResponse"> & {
  /**
   * The list of active account hooks.
   *
   * @generated from field: repeated mgmt.v1alpha1.AccountHook hooks = 1;
   */
  hooks: AccountHook[];
};

/**
 * Describes the message mgmt.v1alpha1.GetActiveAccountHooksByEventResponse.
 * Use `create(GetActiveAccountHooksByEventResponseSchema)` to create a new message.
 */
export const GetActiveAccountHooksByEventResponseSchema: GenMessage<GetActiveAccountHooksByEventResponse> = /*@__PURE__*/
  messageDesc(file_mgmt_v1alpha1_account_hook, 18);

/**
 * Enum of all possible events that can trigger an account hook.
 *
 * @generated from enum mgmt.v1alpha1.AccountHookEvent
 */
export enum AccountHookEvent {
  /**
   * If unspecified, hook will be triggered for all events.
   *
   * @generated from enum value: ACCOUNT_HOOK_EVENT_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * Triggered when a job run is created.
   *
   * @generated from enum value: ACCOUNT_HOOK_EVENT_JOB_RUN_CREATED = 1;
   */
  JOB_RUN_CREATED = 1,

  /**
   * Triggered when a job run fails.
   *
   * @generated from enum value: ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED = 2;
   */
  JOB_RUN_FAILED = 2,

  /**
   * Triggered when a job run succeeds.
   *
   * @generated from enum value: ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED = 3;
   */
  JOB_RUN_SUCCEEDED = 3,
}

/**
 * Describes the enum mgmt.v1alpha1.AccountHookEvent.
 */
export const AccountHookEventSchema: GenEnum<AccountHookEvent> = /*@__PURE__*/
  enumDesc(file_mgmt_v1alpha1_account_hook, 0);

/**
 * @generated from service mgmt.v1alpha1.AccountHookService
 */
export const AccountHookService: GenService<{
  /**
   * Retrieves all account hooks.
   *
   * @generated from rpc mgmt.v1alpha1.AccountHookService.GetAccountHooks
   */
  getAccountHooks: {
    methodKind: "unary";
    input: typeof GetAccountHooksRequestSchema;
    output: typeof GetAccountHooksResponseSchema;
  },
  /**
   * Retrieves a specific account hook.
   *
   * @generated from rpc mgmt.v1alpha1.AccountHookService.GetAccountHook
   */
  getAccountHook: {
    methodKind: "unary";
    input: typeof GetAccountHookRequestSchema;
    output: typeof GetAccountHookResponseSchema;
  },
  /**
   * Creates a new account hook.
   *
   * @generated from rpc mgmt.v1alpha1.AccountHookService.CreateAccountHook
   */
  createAccountHook: {
    methodKind: "unary";
    input: typeof CreateAccountHookRequestSchema;
    output: typeof CreateAccountHookResponseSchema;
  },
  /**
   * Updates an existing account hook.
   *
   * @generated from rpc mgmt.v1alpha1.AccountHookService.UpdateAccountHook
   */
  updateAccountHook: {
    methodKind: "unary";
    input: typeof UpdateAccountHookRequestSchema;
    output: typeof UpdateAccountHookResponseSchema;
  },
  /**
   * Deletes an account hook.
   *
   * @generated from rpc mgmt.v1alpha1.AccountHookService.DeleteAccountHook
   */
  deleteAccountHook: {
    methodKind: "unary";
    input: typeof DeleteAccountHookRequestSchema;
    output: typeof DeleteAccountHookResponseSchema;
  },
  /**
   * Checks if an account hook name is available.
   *
   * @generated from rpc mgmt.v1alpha1.AccountHookService.IsAccountHookNameAvailable
   */
  isAccountHookNameAvailable: {
    methodKind: "unary";
    input: typeof IsAccountHookNameAvailableRequestSchema;
    output: typeof IsAccountHookNameAvailableResponseSchema;
  },
  /**
   * Enables or disables an account hook.
   *
   * @generated from rpc mgmt.v1alpha1.AccountHookService.SetAccountHookEnabled
   */
  setAccountHookEnabled: {
    methodKind: "unary";
    input: typeof SetAccountHookEnabledRequestSchema;
    output: typeof SetAccountHookEnabledResponseSchema;
  },
  /**
   * Retrieves all active account hooks for a specific event.
   *
   * @generated from rpc mgmt.v1alpha1.AccountHookService.GetActiveAccountHooksByEvent
   */
  getActiveAccountHooksByEvent: {
    methodKind: "unary";
    input: typeof GetActiveAccountHooksByEventRequestSchema;
    output: typeof GetActiveAccountHooksByEventResponseSchema;
  },
}> = /*@__PURE__*/
  serviceDesc(file_mgmt_v1alpha1_account_hook, 0);

