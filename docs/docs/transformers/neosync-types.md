---
title: Neosync Types
description: Learn about Neosync's internal type system for data transformation
id: neosync-types
hide_title: false
slug: /transformers/neosync-types
---

# Neosync Types

Neosync implements an internal type system to handle data transformations across different databases and formats. When using JavaScript transformers, these types are automatically converted to JavaScript native types and each field in the type is accessible as a property on the object.

### NeosyncArray

Represents array/list data types across databases. Can contain elements of any other Neosync type.

```go
type NeosyncArray struct {
    Elements []NeosyncAdapter
}
```

### NeosyncDateTime 

Handles date and time values with timezone support.

```go
type NeosyncDateTime struct {
    Year      int
    Month     int    // 1-12
    Day       int    // 1-31
    Hour      int    // 0-23
    Minute    int    // 0-59
    Second    int    // 0-59
    Nano      int    // 0-999999999
    TimeZone  string // Optional timezone
    IsBC      bool   // For dates Before Common Era
}
```

### Interval

Represents time intervals and durations.

```go
type Interval struct {
    Microseconds int64
    Days         int32
    Months       int32
}
```

### Binary

Handles binary/blob data types.

```go
type Binary struct {
    Bytes []byte
}
```

### Bits

Represents bit string types.

```go
type Bits struct {
    Bytes []byte
    Len   int32
}
```
