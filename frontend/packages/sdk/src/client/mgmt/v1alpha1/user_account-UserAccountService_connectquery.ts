// @generated by protoc-gen-connect-query v1.4.2 with parameter "target=ts,import_extension=.js"
// @generated from file mgmt/v1alpha1/user_account.proto (package mgmt.v1alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";
import { AcceptTeamAccountInviteRequest, AcceptTeamAccountInviteResponse, ConvertPersonalToTeamAccountRequest, ConvertPersonalToTeamAccountResponse, CreateTeamAccountRequest, CreateTeamAccountResponse, GetAccountBillingCheckoutSessionRequest, GetAccountBillingCheckoutSessionResponse, GetAccountBillingPortalSessionRequest, GetAccountBillingPortalSessionResponse, GetAccountOnboardingConfigRequest, GetAccountOnboardingConfigResponse, GetAccountStatusRequest, GetAccountStatusResponse, GetAccountTemporalConfigRequest, GetAccountTemporalConfigResponse, GetBillingAccountsRequest, GetBillingAccountsResponse, GetSystemInformationRequest, GetSystemInformationResponse, GetTeamAccountInvitesRequest, GetTeamAccountInvitesResponse, GetTeamAccountMembersRequest, GetTeamAccountMembersResponse, GetUserAccountsRequest, GetUserAccountsResponse, GetUserRequest, GetUserResponse, InviteUserToTeamAccountRequest, InviteUserToTeamAccountResponse, IsAccountStatusValidRequest, IsAccountStatusValidResponse, IsUserInAccountRequest, IsUserInAccountResponse, RemoveTeamAccountInviteRequest, RemoveTeamAccountInviteResponse, RemoveTeamAccountMemberRequest, RemoveTeamAccountMemberResponse, SetAccountOnboardingConfigRequest, SetAccountOnboardingConfigResponse, SetAccountTemporalConfigRequest, SetAccountTemporalConfigResponse, SetBillingMeterEventRequest, SetBillingMeterEventResponse, SetPersonalAccountRequest, SetPersonalAccountResponse, SetUserRequest, SetUserResponse, SetUserRoleRequest, SetUserRoleResponse } from "./user_account_pb.js";

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetUser
 */
export const getUser = {
  localName: "getUser",
  name: "GetUser",
  kind: MethodKind.Unary,
  I: GetUserRequest,
  O: GetUserResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.SetUser
 */
export const setUser = {
  localName: "setUser",
  name: "SetUser",
  kind: MethodKind.Unary,
  I: SetUserRequest,
  O: SetUserResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetUserAccounts
 */
export const getUserAccounts = {
  localName: "getUserAccounts",
  name: "GetUserAccounts",
  kind: MethodKind.Unary,
  I: GetUserAccountsRequest,
  O: GetUserAccountsResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.SetPersonalAccount
 */
export const setPersonalAccount = {
  localName: "setPersonalAccount",
  name: "SetPersonalAccount",
  kind: MethodKind.Unary,
  I: SetPersonalAccountRequest,
  O: SetPersonalAccountResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * Convert a personal account to a team account retaining all of the jobs and connections. This will also create a new empty personal account.
 *
 * @generated from rpc mgmt.v1alpha1.UserAccountService.ConvertPersonalToTeamAccount
 */
export const convertPersonalToTeamAccount = {
  localName: "convertPersonalToTeamAccount",
  name: "ConvertPersonalToTeamAccount",
  kind: MethodKind.Unary,
  I: ConvertPersonalToTeamAccountRequest,
  O: ConvertPersonalToTeamAccountResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * Creates a new team account
 *
 * @generated from rpc mgmt.v1alpha1.UserAccountService.CreateTeamAccount
 */
export const createTeamAccount = {
  localName: "createTeamAccount",
  name: "CreateTeamAccount",
  kind: MethodKind.Unary,
  I: CreateTeamAccountRequest,
  O: CreateTeamAccountResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.IsUserInAccount
 */
export const isUserInAccount = {
  localName: "isUserInAccount",
  name: "IsUserInAccount",
  kind: MethodKind.Unary,
  I: IsUserInAccountRequest,
  O: IsUserInAccountResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetAccountTemporalConfig
 */
export const getAccountTemporalConfig = {
  localName: "getAccountTemporalConfig",
  name: "GetAccountTemporalConfig",
  kind: MethodKind.Unary,
  I: GetAccountTemporalConfigRequest,
  O: GetAccountTemporalConfigResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.SetAccountTemporalConfig
 */
export const setAccountTemporalConfig = {
  localName: "setAccountTemporalConfig",
  name: "SetAccountTemporalConfig",
  kind: MethodKind.Unary,
  I: SetAccountTemporalConfigRequest,
  O: SetAccountTemporalConfigResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetTeamAccountMembers
 */
export const getTeamAccountMembers = {
  localName: "getTeamAccountMembers",
  name: "GetTeamAccountMembers",
  kind: MethodKind.Unary,
  I: GetTeamAccountMembersRequest,
  O: GetTeamAccountMembersResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.RemoveTeamAccountMember
 */
export const removeTeamAccountMember = {
  localName: "removeTeamAccountMember",
  name: "RemoveTeamAccountMember",
  kind: MethodKind.Unary,
  I: RemoveTeamAccountMemberRequest,
  O: RemoveTeamAccountMemberResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.InviteUserToTeamAccount
 */
export const inviteUserToTeamAccount = {
  localName: "inviteUserToTeamAccount",
  name: "InviteUserToTeamAccount",
  kind: MethodKind.Unary,
  I: InviteUserToTeamAccountRequest,
  O: InviteUserToTeamAccountResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetTeamAccountInvites
 */
export const getTeamAccountInvites = {
  localName: "getTeamAccountInvites",
  name: "GetTeamAccountInvites",
  kind: MethodKind.Unary,
  I: GetTeamAccountInvitesRequest,
  O: GetTeamAccountInvitesResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.RemoveTeamAccountInvite
 */
export const removeTeamAccountInvite = {
  localName: "removeTeamAccountInvite",
  name: "RemoveTeamAccountInvite",
  kind: MethodKind.Unary,
  I: RemoveTeamAccountInviteRequest,
  O: RemoveTeamAccountInviteResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.AcceptTeamAccountInvite
 */
export const acceptTeamAccountInvite = {
  localName: "acceptTeamAccountInvite",
  name: "AcceptTeamAccountInvite",
  kind: MethodKind.Unary,
  I: AcceptTeamAccountInviteRequest,
  O: AcceptTeamAccountInviteResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetSystemInformation
 */
export const getSystemInformation = {
  localName: "getSystemInformation",
  name: "GetSystemInformation",
  kind: MethodKind.Unary,
  I: GetSystemInformationRequest,
  O: GetSystemInformationResponse,
      idempotency: MethodIdempotency.NoSideEffects,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetAccountOnboardingConfig
 */
export const getAccountOnboardingConfig = {
  localName: "getAccountOnboardingConfig",
  name: "GetAccountOnboardingConfig",
  kind: MethodKind.Unary,
  I: GetAccountOnboardingConfigRequest,
  O: GetAccountOnboardingConfigResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * @generated from rpc mgmt.v1alpha1.UserAccountService.SetAccountOnboardingConfig
 */
export const setAccountOnboardingConfig = {
  localName: "setAccountOnboardingConfig",
  name: "SetAccountOnboardingConfig",
  kind: MethodKind.Unary,
  I: SetAccountOnboardingConfigRequest,
  O: SetAccountOnboardingConfigResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * Returns different metrics on the account status for the active billing period
 *
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetAccountStatus
 */
export const getAccountStatus = {
  localName: "getAccountStatus",
  name: "GetAccountStatus",
  kind: MethodKind.Unary,
  I: GetAccountStatusRequest,
  O: GetAccountStatusResponse,
      idempotency: MethodIdempotency.NoSideEffects,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * Distils the account status down to whether not it is in a valid state.
 *
 * @generated from rpc mgmt.v1alpha1.UserAccountService.IsAccountStatusValid
 */
export const isAccountStatusValid = {
  localName: "isAccountStatusValid",
  name: "IsAccountStatusValid",
  kind: MethodKind.Unary,
  I: IsAccountStatusValidRequest,
  O: IsAccountStatusValidResponse,
      idempotency: MethodIdempotency.NoSideEffects,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * Returns a new checkout session for the account to subscribe
 *
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetAccountBillingCheckoutSession
 */
export const getAccountBillingCheckoutSession = {
  localName: "getAccountBillingCheckoutSession",
  name: "GetAccountBillingCheckoutSession",
  kind: MethodKind.Unary,
  I: GetAccountBillingCheckoutSessionRequest,
  O: GetAccountBillingCheckoutSessionResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * Returns a new billing portal session if the account has a billing customer id
 *
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetAccountBillingPortalSession
 */
export const getAccountBillingPortalSession = {
  localName: "getAccountBillingPortalSession",
  name: "GetAccountBillingPortalSession",
  kind: MethodKind.Unary,
  I: GetAccountBillingPortalSessionRequest,
  O: GetAccountBillingPortalSessionResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * Returns user accounts that have a billing id.
 *
 * @generated from rpc mgmt.v1alpha1.UserAccountService.GetBillingAccounts
 */
export const getBillingAccounts = {
  localName: "getBillingAccounts",
  name: "GetBillingAccounts",
  kind: MethodKind.Unary,
  I: GetBillingAccountsRequest,
  O: GetBillingAccountsResponse,
      idempotency: MethodIdempotency.NoSideEffects,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * Sends a new metered event to the billing system
 *
 * @generated from rpc mgmt.v1alpha1.UserAccountService.SetBillingMeterEvent
 */
export const setBillingMeterEvent = {
  localName: "setBillingMeterEvent",
  name: "SetBillingMeterEvent",
  kind: MethodKind.Unary,
  I: SetBillingMeterEventRequest,
  O: SetBillingMeterEventResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;

/**
 * Sets the users role
 *
 * @generated from rpc mgmt.v1alpha1.UserAccountService.SetUserRole
 */
export const setUserRole = {
  localName: "setUserRole",
  name: "SetUserRole",
  kind: MethodKind.Unary,
  I: SetUserRoleRequest,
  O: SetUserRoleResponse,
  service: {
    typeName: "mgmt.v1alpha1.UserAccountService"
  }
} as const;
