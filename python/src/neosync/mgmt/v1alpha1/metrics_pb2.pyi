from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RangedMetricName(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    RANGED_METRIC_NAME_UNSPECIFIED: _ClassVar[RangedMetricName]
    RANGED_METRIC_NAME_INPUT_RECEIVED: _ClassVar[RangedMetricName]
RANGED_METRIC_NAME_UNSPECIFIED: RangedMetricName
RANGED_METRIC_NAME_INPUT_RECEIVED: RangedMetricName

class Date(_message.Message):
    __slots__ = ("year", "month", "day")
    YEAR_FIELD_NUMBER: _ClassVar[int]
    MONTH_FIELD_NUMBER: _ClassVar[int]
    DAY_FIELD_NUMBER: _ClassVar[int]
    year: int
    month: int
    day: int
    def __init__(self, year: _Optional[int] = ..., month: _Optional[int] = ..., day: _Optional[int] = ...) -> None: ...

class GetDailyMetricCountRequest(_message.Message):
    __slots__ = ("start", "end", "metric", "account_id", "job_id", "run_id")
    START_FIELD_NUMBER: _ClassVar[int]
    END_FIELD_NUMBER: _ClassVar[int]
    METRIC_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    start: Date
    end: Date
    metric: RangedMetricName
    account_id: str
    job_id: str
    run_id: str
    def __init__(self, start: _Optional[_Union[Date, _Mapping]] = ..., end: _Optional[_Union[Date, _Mapping]] = ..., metric: _Optional[_Union[RangedMetricName, str]] = ..., account_id: _Optional[str] = ..., job_id: _Optional[str] = ..., run_id: _Optional[str] = ...) -> None: ...

class GetDailyMetricCountResponse(_message.Message):
    __slots__ = ("results",)
    RESULTS_FIELD_NUMBER: _ClassVar[int]
    results: _containers.RepeatedCompositeFieldContainer[DayResult]
    def __init__(self, results: _Optional[_Iterable[_Union[DayResult, _Mapping]]] = ...) -> None: ...

class DayResult(_message.Message):
    __slots__ = ("date", "count")
    DATE_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    date: Date
    count: int
    def __init__(self, date: _Optional[_Union[Date, _Mapping]] = ..., count: _Optional[int] = ...) -> None: ...

class GetMetricCountRequest(_message.Message):
    __slots__ = ("metric", "account_id", "job_id", "run_id", "start_day", "end_day")
    METRIC_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    START_DAY_FIELD_NUMBER: _ClassVar[int]
    END_DAY_FIELD_NUMBER: _ClassVar[int]
    metric: RangedMetricName
    account_id: str
    job_id: str
    run_id: str
    start_day: Date
    end_day: Date
    def __init__(self, metric: _Optional[_Union[RangedMetricName, str]] = ..., account_id: _Optional[str] = ..., job_id: _Optional[str] = ..., run_id: _Optional[str] = ..., start_day: _Optional[_Union[Date, _Mapping]] = ..., end_day: _Optional[_Union[Date, _Mapping]] = ...) -> None: ...

class GetMetricCountResponse(_message.Message):
    __slots__ = ("count",)
    COUNT_FIELD_NUMBER: _ClassVar[int]
    count: int
    def __init__(self, count: _Optional[int] = ...) -> None: ...
