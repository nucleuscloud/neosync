from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class UserAccountType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    USER_ACCOUNT_TYPE_UNSPECIFIED: _ClassVar[UserAccountType]
    USER_ACCOUNT_TYPE_PERSONAL: _ClassVar[UserAccountType]
    USER_ACCOUNT_TYPE_TEAM: _ClassVar[UserAccountType]
    USER_ACCOUNT_TYPE_ENTERPRISE: _ClassVar[UserAccountType]

class BillingStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    BILLING_STATUS_UNSPECIFIED: _ClassVar[BillingStatus]
    BILLING_STATUS_ACTIVE: _ClassVar[BillingStatus]
    BILLING_STATUS_EXPIRED: _ClassVar[BillingStatus]
    BILLING_STATUS_TRIAL_ACTIVE: _ClassVar[BillingStatus]
    BILLING_STATUS_TRIAL_EXPIRED: _ClassVar[BillingStatus]

class AccountStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ACCOUNT_STATUS_REASON_UNSPECIFIED: _ClassVar[AccountStatus]
    ACCOUNT_STATUS_EXCEEDS_ALLOWED_LIMIT: _ClassVar[AccountStatus]
    ACCOUNT_STATUS_REQUESTED_EXCEEDS_LIMIT: _ClassVar[AccountStatus]
    ACCOUNT_STATUS_ACCOUNT_IN_EXPIRED_STATE: _ClassVar[AccountStatus]
    ACCOUNT_STATUS_ACCOUNT_TRIAL_ACTIVE: _ClassVar[AccountStatus]
    ACCOUNT_STATUS_ACCOUNT_TRIAL_EXPIRED: _ClassVar[AccountStatus]

class AccountRole(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ACCOUNT_ROLE_UNSPECIFIED: _ClassVar[AccountRole]
    ACCOUNT_ROLE_ADMIN: _ClassVar[AccountRole]
    ACCOUNT_ROLE_JOB_DEVELOPER: _ClassVar[AccountRole]
    ACCOUNT_ROLE_JOB_VIEWER: _ClassVar[AccountRole]
    ACCOUNT_ROLE_JOB_EXECUTOR: _ClassVar[AccountRole]
USER_ACCOUNT_TYPE_UNSPECIFIED: UserAccountType
USER_ACCOUNT_TYPE_PERSONAL: UserAccountType
USER_ACCOUNT_TYPE_TEAM: UserAccountType
USER_ACCOUNT_TYPE_ENTERPRISE: UserAccountType
BILLING_STATUS_UNSPECIFIED: BillingStatus
BILLING_STATUS_ACTIVE: BillingStatus
BILLING_STATUS_EXPIRED: BillingStatus
BILLING_STATUS_TRIAL_ACTIVE: BillingStatus
BILLING_STATUS_TRIAL_EXPIRED: BillingStatus
ACCOUNT_STATUS_REASON_UNSPECIFIED: AccountStatus
ACCOUNT_STATUS_EXCEEDS_ALLOWED_LIMIT: AccountStatus
ACCOUNT_STATUS_REQUESTED_EXCEEDS_LIMIT: AccountStatus
ACCOUNT_STATUS_ACCOUNT_IN_EXPIRED_STATE: AccountStatus
ACCOUNT_STATUS_ACCOUNT_TRIAL_ACTIVE: AccountStatus
ACCOUNT_STATUS_ACCOUNT_TRIAL_EXPIRED: AccountStatus
ACCOUNT_ROLE_UNSPECIFIED: AccountRole
ACCOUNT_ROLE_ADMIN: AccountRole
ACCOUNT_ROLE_JOB_DEVELOPER: AccountRole
ACCOUNT_ROLE_JOB_VIEWER: AccountRole
ACCOUNT_ROLE_JOB_EXECUTOR: AccountRole

class GetUserRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetUserResponse(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class SetUserRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class SetUserResponse(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class GetUserAccountsRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetUserAccountsResponse(_message.Message):
    __slots__ = ("accounts",)
    ACCOUNTS_FIELD_NUMBER: _ClassVar[int]
    accounts: _containers.RepeatedCompositeFieldContainer[UserAccount]
    def __init__(self, accounts: _Optional[_Iterable[_Union[UserAccount, _Mapping]]] = ...) -> None: ...

class UserAccount(_message.Message):
    __slots__ = ("id", "name", "type", "has_stripe_customer_id")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    HAS_STRIPE_CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    type: UserAccountType
    has_stripe_customer_id: bool
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., type: _Optional[_Union[UserAccountType, str]] = ..., has_stripe_customer_id: bool = ...) -> None: ...

class ConvertPersonalToTeamAccountRequest(_message.Message):
    __slots__ = ("name", "account_id")
    NAME_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    name: str
    account_id: str
    def __init__(self, name: _Optional[str] = ..., account_id: _Optional[str] = ...) -> None: ...

class ConvertPersonalToTeamAccountResponse(_message.Message):
    __slots__ = ("account_id", "checkout_session_url", "new_personal_account_id")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    CHECKOUT_SESSION_URL_FIELD_NUMBER: _ClassVar[int]
    NEW_PERSONAL_ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    checkout_session_url: str
    new_personal_account_id: str
    def __init__(self, account_id: _Optional[str] = ..., checkout_session_url: _Optional[str] = ..., new_personal_account_id: _Optional[str] = ...) -> None: ...

class SetPersonalAccountRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class SetPersonalAccountResponse(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class IsUserInAccountRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class IsUserInAccountResponse(_message.Message):
    __slots__ = ("ok",)
    OK_FIELD_NUMBER: _ClassVar[int]
    ok: bool
    def __init__(self, ok: bool = ...) -> None: ...

class GetAccountTemporalConfigRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetAccountTemporalConfigResponse(_message.Message):
    __slots__ = ("config",)
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    config: AccountTemporalConfig
    def __init__(self, config: _Optional[_Union[AccountTemporalConfig, _Mapping]] = ...) -> None: ...

class SetAccountTemporalConfigRequest(_message.Message):
    __slots__ = ("account_id", "config")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    config: AccountTemporalConfig
    def __init__(self, account_id: _Optional[str] = ..., config: _Optional[_Union[AccountTemporalConfig, _Mapping]] = ...) -> None: ...

class SetAccountTemporalConfigResponse(_message.Message):
    __slots__ = ("config",)
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    config: AccountTemporalConfig
    def __init__(self, config: _Optional[_Union[AccountTemporalConfig, _Mapping]] = ...) -> None: ...

class AccountTemporalConfig(_message.Message):
    __slots__ = ("url", "namespace", "sync_job_queue_name")
    URL_FIELD_NUMBER: _ClassVar[int]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    SYNC_JOB_QUEUE_NAME_FIELD_NUMBER: _ClassVar[int]
    url: str
    namespace: str
    sync_job_queue_name: str
    def __init__(self, url: _Optional[str] = ..., namespace: _Optional[str] = ..., sync_job_queue_name: _Optional[str] = ...) -> None: ...

class CreateTeamAccountRequest(_message.Message):
    __slots__ = ("name",)
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class CreateTeamAccountResponse(_message.Message):
    __slots__ = ("account_id", "checkout_session_url")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    CHECKOUT_SESSION_URL_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    checkout_session_url: str
    def __init__(self, account_id: _Optional[str] = ..., checkout_session_url: _Optional[str] = ...) -> None: ...

class AccountUser(_message.Message):
    __slots__ = ("id", "name", "image", "email")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    image: str
    email: str
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., image: _Optional[str] = ..., email: _Optional[str] = ...) -> None: ...

class GetTeamAccountMembersRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetTeamAccountMembersResponse(_message.Message):
    __slots__ = ("users",)
    USERS_FIELD_NUMBER: _ClassVar[int]
    users: _containers.RepeatedCompositeFieldContainer[AccountUser]
    def __init__(self, users: _Optional[_Iterable[_Union[AccountUser, _Mapping]]] = ...) -> None: ...

class RemoveTeamAccountMemberRequest(_message.Message):
    __slots__ = ("user_id", "account_id")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    account_id: str
    def __init__(self, user_id: _Optional[str] = ..., account_id: _Optional[str] = ...) -> None: ...

class RemoveTeamAccountMemberResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class InviteUserToTeamAccountRequest(_message.Message):
    __slots__ = ("account_id", "email")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    email: str
    def __init__(self, account_id: _Optional[str] = ..., email: _Optional[str] = ...) -> None: ...

class AccountInvite(_message.Message):
    __slots__ = ("id", "account_id", "sender_user_id", "email", "token", "accepted", "created_at", "updated_at", "expires_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    SENDER_USER_ID_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    ACCEPTED_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    id: str
    account_id: str
    sender_user_id: str
    email: str
    token: str
    accepted: bool
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    expires_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[str] = ..., account_id: _Optional[str] = ..., sender_user_id: _Optional[str] = ..., email: _Optional[str] = ..., token: _Optional[str] = ..., accepted: bool = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class InviteUserToTeamAccountResponse(_message.Message):
    __slots__ = ("invite",)
    INVITE_FIELD_NUMBER: _ClassVar[int]
    invite: AccountInvite
    def __init__(self, invite: _Optional[_Union[AccountInvite, _Mapping]] = ...) -> None: ...

class GetTeamAccountInvitesRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetTeamAccountInvitesResponse(_message.Message):
    __slots__ = ("invites",)
    INVITES_FIELD_NUMBER: _ClassVar[int]
    invites: _containers.RepeatedCompositeFieldContainer[AccountInvite]
    def __init__(self, invites: _Optional[_Iterable[_Union[AccountInvite, _Mapping]]] = ...) -> None: ...

class RemoveTeamAccountInviteRequest(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class RemoveTeamAccountInviteResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class AcceptTeamAccountInviteRequest(_message.Message):
    __slots__ = ("token",)
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    token: str
    def __init__(self, token: _Optional[str] = ...) -> None: ...

class AcceptTeamAccountInviteResponse(_message.Message):
    __slots__ = ("account",)
    ACCOUNT_FIELD_NUMBER: _ClassVar[int]
    account: UserAccount
    def __init__(self, account: _Optional[_Union[UserAccount, _Mapping]] = ...) -> None: ...

class GetSystemInformationRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetSystemInformationResponse(_message.Message):
    __slots__ = ("version", "commit", "compiler", "platform", "build_date")
    VERSION_FIELD_NUMBER: _ClassVar[int]
    COMMIT_FIELD_NUMBER: _ClassVar[int]
    COMPILER_FIELD_NUMBER: _ClassVar[int]
    PLATFORM_FIELD_NUMBER: _ClassVar[int]
    BUILD_DATE_FIELD_NUMBER: _ClassVar[int]
    version: str
    commit: str
    compiler: str
    platform: str
    build_date: _timestamp_pb2.Timestamp
    def __init__(self, version: _Optional[str] = ..., commit: _Optional[str] = ..., compiler: _Optional[str] = ..., platform: _Optional[str] = ..., build_date: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class GetAccountOnboardingConfigRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetAccountOnboardingConfigResponse(_message.Message):
    __slots__ = ("config",)
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    config: AccountOnboardingConfig
    def __init__(self, config: _Optional[_Union[AccountOnboardingConfig, _Mapping]] = ...) -> None: ...

class SetAccountOnboardingConfigRequest(_message.Message):
    __slots__ = ("account_id", "config")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    config: AccountOnboardingConfig
    def __init__(self, account_id: _Optional[str] = ..., config: _Optional[_Union[AccountOnboardingConfig, _Mapping]] = ...) -> None: ...

class SetAccountOnboardingConfigResponse(_message.Message):
    __slots__ = ("config",)
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    config: AccountOnboardingConfig
    def __init__(self, config: _Optional[_Union[AccountOnboardingConfig, _Mapping]] = ...) -> None: ...

class AccountOnboardingConfig(_message.Message):
    __slots__ = ("has_created_source_connection", "has_created_destination_connection", "has_created_job", "has_invited_members", "has_completed_onboarding")
    HAS_CREATED_SOURCE_CONNECTION_FIELD_NUMBER: _ClassVar[int]
    HAS_CREATED_DESTINATION_CONNECTION_FIELD_NUMBER: _ClassVar[int]
    HAS_CREATED_JOB_FIELD_NUMBER: _ClassVar[int]
    HAS_INVITED_MEMBERS_FIELD_NUMBER: _ClassVar[int]
    HAS_COMPLETED_ONBOARDING_FIELD_NUMBER: _ClassVar[int]
    has_created_source_connection: bool
    has_created_destination_connection: bool
    has_created_job: bool
    has_invited_members: bool
    has_completed_onboarding: bool
    def __init__(self, has_created_source_connection: bool = ..., has_created_destination_connection: bool = ..., has_created_job: bool = ..., has_invited_members: bool = ..., has_completed_onboarding: bool = ...) -> None: ...

class GetAccountStatusRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetAccountStatusResponse(_message.Message):
    __slots__ = ("used_record_count", "allowed_record_count", "subscription_status")
    USED_RECORD_COUNT_FIELD_NUMBER: _ClassVar[int]
    ALLOWED_RECORD_COUNT_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_STATUS_FIELD_NUMBER: _ClassVar[int]
    used_record_count: int
    allowed_record_count: int
    subscription_status: BillingStatus
    def __init__(self, used_record_count: _Optional[int] = ..., allowed_record_count: _Optional[int] = ..., subscription_status: _Optional[_Union[BillingStatus, str]] = ...) -> None: ...

class IsAccountStatusValidRequest(_message.Message):
    __slots__ = ("account_id", "requested_record_count")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    REQUESTED_RECORD_COUNT_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    requested_record_count: int
    def __init__(self, account_id: _Optional[str] = ..., requested_record_count: _Optional[int] = ...) -> None: ...

class IsAccountStatusValidResponse(_message.Message):
    __slots__ = ("is_valid", "reason", "should_poll", "used_record_count", "allowed_record_count", "account_status", "trial_expires_at")
    IS_VALID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    SHOULD_POLL_FIELD_NUMBER: _ClassVar[int]
    USED_RECORD_COUNT_FIELD_NUMBER: _ClassVar[int]
    ALLOWED_RECORD_COUNT_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_STATUS_FIELD_NUMBER: _ClassVar[int]
    TRIAL_EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    is_valid: bool
    reason: str
    should_poll: bool
    used_record_count: int
    allowed_record_count: int
    account_status: AccountStatus
    trial_expires_at: _timestamp_pb2.Timestamp
    def __init__(self, is_valid: bool = ..., reason: _Optional[str] = ..., should_poll: bool = ..., used_record_count: _Optional[int] = ..., allowed_record_count: _Optional[int] = ..., account_status: _Optional[_Union[AccountStatus, str]] = ..., trial_expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class GetAccountBillingCheckoutSessionRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetAccountBillingCheckoutSessionResponse(_message.Message):
    __slots__ = ("checkout_session_url",)
    CHECKOUT_SESSION_URL_FIELD_NUMBER: _ClassVar[int]
    checkout_session_url: str
    def __init__(self, checkout_session_url: _Optional[str] = ...) -> None: ...

class GetAccountBillingPortalSessionRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetAccountBillingPortalSessionResponse(_message.Message):
    __slots__ = ("portal_session_url",)
    PORTAL_SESSION_URL_FIELD_NUMBER: _ClassVar[int]
    portal_session_url: str
    def __init__(self, portal_session_url: _Optional[str] = ...) -> None: ...

class GetBillingAccountsRequest(_message.Message):
    __slots__ = ("account_ids",)
    ACCOUNT_IDS_FIELD_NUMBER: _ClassVar[int]
    account_ids: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, account_ids: _Optional[_Iterable[str]] = ...) -> None: ...

class GetBillingAccountsResponse(_message.Message):
    __slots__ = ("accounts",)
    ACCOUNTS_FIELD_NUMBER: _ClassVar[int]
    accounts: _containers.RepeatedCompositeFieldContainer[UserAccount]
    def __init__(self, accounts: _Optional[_Iterable[_Union[UserAccount, _Mapping]]] = ...) -> None: ...

class SetBillingMeterEventRequest(_message.Message):
    __slots__ = ("account_id", "event_name", "value", "event_id", "timestamp")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    event_name: str
    value: str
    event_id: str
    timestamp: int
    def __init__(self, account_id: _Optional[str] = ..., event_name: _Optional[str] = ..., value: _Optional[str] = ..., event_id: _Optional[str] = ..., timestamp: _Optional[int] = ...) -> None: ...

class SetBillingMeterEventResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class SetUserRoleRequest(_message.Message):
    __slots__ = ("account_id", "user_id", "role")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    user_id: str
    role: AccountRole
    def __init__(self, account_id: _Optional[str] = ..., user_id: _Optional[str] = ..., role: _Optional[_Union[AccountRole, str]] = ...) -> None: ...

class SetUserRoleResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
