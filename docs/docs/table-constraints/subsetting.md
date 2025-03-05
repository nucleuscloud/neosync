---
title: Subsetting
description: Learn how Neosync subsets data for a better local development experience
id: subsetting
hide_title: false
slug: /table-constraints/subsetting
---

# Subsetting with Neosync

Neosync offers a powerful subsetting feature that allows users to efficiently manage and extract relevant data from complex database schemas. This feature is particularly useful for maintaining data integrity and optimizing data processing tasks.

## Key Features of Subsetting

### 1. Foreign Key-Based Subsetting

Neosync leverages foreign key constraints to automatically subset related tables. When a subset condition (such as a WHERE clause) is specified on a table, Neosync can propagate this condition to related tables through foreign key relationships. This ensures that only the relevant data is included in your subset, maintaining referential integrity across your data. Neosync processes the subsetting directionally from top down to ensure data integrity.

- **Example**: If you have a `users` table with a condition `id < 100` and an `orders` table with a foreign key to `users`, Neosync will ensure that only orders associated with the subset of users are included.

### 2. Enabling Subsetting by Foreign Key

This feature can be toggled on or off when creating your job. When enabled, Neosync uses foreign key relationships to create appropriate joins between tables, ensuring that your subset maintains referential integrity throughout the database. If this feature is not enabled, the WHERE clause is only applied to the table it is set on, and related tables will not be subsetted based on foreign key constraints.

- **Multiple Subset Queries**: You can add multiple subset queries to different tables, and Neosync will handle the relationships between them appropriately.

### 3. Handling Complex Relationships

Neosync is adept at managing self-referencing tables and circular dependencies. It can handle complex database schemas, provided there is at least one nullable column within the circular dependency cycle to serve as a viable entry point.

- **Circular Dependencies**: Neosync can identify and manage circular dependencies, ensuring that data is processed in the correct order to maintain integrity.

### 4. Query Building and Execution

Neosync builds SQL SELECT queries for each table based on the specified subset conditions. It supports various database drivers, including PostgreSQL, MySQL, and SQL Server, and can generate queries that respect foreign key constraints. Users can view the generated SQL SELECT queries on the job run page.

## Conclusion

Neosync's subsetting feature simplifies the process of creating data subsets from complex databases. By leveraging foreign key constraints and handling complex relationships, Neosync ensures that all related data is cohesively maintained while preserving the relationships between your tables.

## Examples Diagram

![circref](/img/subset-diagram.png)

## Examples: Subsetting by Foreign Key Enabled

### Subsetting with a WHERE Clause on the `addresses` Table

Consider the diagram above where we have three tables: `customers`, `addresses`, and `orders`. The `addresses` table has a foreign key relationship with the `customers` table, and the `orders` table has a foreign key relationship with the `addresses` table.

If a WHERE clause is applied to the `addresses` table, Neosync will propagate this condition to the related tables as follows:

**Addresses Table**: The WHERE clause is directly applied to the `addresses` table. For example, if the condition is `city = 'New York'`, only the addresses in New York will be included in the subset.

```sql
SELECT
	*
FROM
	addresses
WHERE
	city = 'New York';
```

**Customers Table**: Neosync will identify the foreign key relationship between the `addresses` and `customers` tables. It will then include only the customers who have addresses that meet the WHERE clause condition. This ensures that the subset of customers is relevant to the subset of addresses.

```sql
SELECT
	*
FROM
	customers
	JOIN addresses ON customers.address_id = addresses.id
WHERE
	addresses.city = 'New York';
```

**Orders Table**: Neosync will identify the foreign key relationships between the `orders`, `addresses`, and `customers` tables. It will join the `orders` table with both the `addresses` and `customers` tables to ensure that only the orders associated with the addresses meeting the WHERE clause condition and their corresponding customers are included. This ensures that the subset of orders is relevant to the subset of addresses and customers.

```sql
SELECT
	*
FROM
	orders
	JOIN customers ON orders.customer_id = customers.id
	JOIN addresses ON orders.address_id = addresses.id
WHERE
	addresses.city = 'New York';
```

**Payments Table**: Neosync will identify the foreign key relationships between the `payments`, `orders`, `addresses`, and `customers` tables. It will join the `payments` table with both the `orders`, `addresses`, and `customers` tables to ensure that only the payments associated with the orders meeting the WHERE clause condition and their corresponding addresses and customers are included. This ensures that the subset of payments is relevant to the subset of addresses and customers.

```sql
SELECT
	*
FROM
	payments
	JOIN orders ON payments.order_id = orders.id
	JOIN customers ON orders.customer_id = customers.id
	JOIN addresses ON orders.address_id = addresses.id
WHERE
	addresses.city = 'New York';
```

### Subsetting with a WHERE Clause on the `customers` Table

If a WHERE clause is applied to the `customers` table, Neosync will propagate this condition to the related tables as follows:

**Addresses Table**: No subset is applied to the `addresses` table because Neosync takes a top-down approach.

**Customers Table**: Neosync will apply the WHERE clause directly to the `customers` table. For example, if the condition is `name = 'Jane'`, only the customers named Jane will be included in the subset.

```sql
SELECT
	*
FROM
	customers
WHERE
	customers.name = 'Jane';
```

**Orders Table**: Neosync will identify the foreign key relationships between the `orders` and `customers` tables. It will join the `orders` table with the `customers` table to ensure that only the orders associated with the customers meeting the WHERE clause condition are included. This ensures that the subset of orders is relevant to the subset of customers.

```sql
SELECT
	*
FROM
	orders
	JOIN customers ON orders.customer_id = customers.id
WHERE
	customers.name = 'Jane';
```

**Payments Table**: Neosync will identify the foreign key relationships between the `payments`, `orders`, and `customers` tables. It will join the `payments` table with both the `orders` and `customers` tables to ensure that only the payments associated with the orders meeting the WHERE clause condition and their corresponding customers are included. This ensures that the subset of payments is relevant to the subset of customers.

```sql
SELECT
	*
FROM
	payments
	JOIN orders ON payments.order_id = orders.id
	JOIN customers ON orders.customer_id = customers.id
WHERE
	customers.name = 'Jane';
```

### Subsetting with a WHERE Clause on the `orders` Table

If a WHERE clause is applied to the `orders` table, Neosync will propagate this condition to the related tables as follows:

**Addresses Table**: No subset is applied to the `addresses` table because Neosync takes a top-down approach.

**Customers Table**: No subset is applied to the `customers` table because Neosync takes a top-down approach.

**Payments Table**: No subset is applied to the `payments` table because Neosync takes a top-down approach.

**Orders Table**: Neosync will apply the WHERE clause directly to the `orders` table. For example, if the condition is `total_amount > 100`, only orders with a total amount greater than 100 will be included in the subset.

### Subsetting with WHERE Clauses on Both `addresses` and `customers` Tables

Suppose you want to subset data such that you only include addresses in New York and customers named Jane. Neosync will handle these conditions as follows:

**Addresses Table**: The WHERE clause is directly applied to the `addresses` table to include only addresses in New York.

```sql
SELECT
    *
FROM
    addresses
WHERE
    city = 'New York';
```

**Customers Table**: The WHERE clause is directly applied to the `customers` table to include only customers named Jane.

```sql
SELECT
	*
FROM
	customers
	JOIN addresses ON customers.address_id = addresses.id
WHERE
	addresses.city = 'New York'
  AND customers.name = 'Jane';
```

**Orders Table**: Neosync will join the `orders` table with both the `addresses` and `customers` tables. It will ensure that only orders associated with the addresses in New York and customers named Jane are included.

```sql
SELECT
    *
FROM
    orders
JOIN customers ON orders.customer_id = customers.id
JOIN addresses ON orders.address_id = addresses.id
WHERE
    addresses.city = 'New York'
    AND customers.name = 'Jane';
```

**Payments Table**: Neosync will join the `payments` table with the `orders`, `addresses`, and `customers` tables. It will ensure that only payments associated with the orders meeting both the WHERE clause conditions on addresses and customers are included.

```sql
SELECT
    *
FROM
    payments
JOIN orders ON payments.order_id = orders.id
JOIN customers ON orders.customer_id = customers.id
JOIN addresses ON orders.address_id = addresses.id
WHERE
    addresses.city = 'New York'
    AND customers.name = 'Jane';
```

In this example, Neosync effectively combines the subset conditions on both the `addresses` and `customers` tables to ensure that all related data in the `orders` and `payments` tables is relevant to the specified conditions.

## Examples: Subsetting by Foreign Key Disabled

### Subsetting with WHERE Clauses on Both `addresses` and `customers` Tables

Neosync will handle these conditions as follows with subsetting by foreign key disabled:

**Addresses Table**: The WHERE clause is directly applied to the `addresses` table to include only addresses in New York.

```sql
SELECT * FROM addresses WHERE city = 'New York';
```

**Customers Table**: The WHERE clause is directly applied to the `customers` table to include only customers named Jane.

```sql
SELECT * FROM customers WHERE customers.name = 'Jane';
```

**Orders Table**: Since foreign key subsetting is disabled, no joins or filters will be applied to the orders table. The orders table will be selected in its entirety without any filtering conditions.

```sql
SELECT * FROM orders;
```

**Payments Table**: Since foreign key subsetting is disabled, no joins or filters will be applied to the payments table. The payments table will be selected in its entirety without any filtering conditions.

```sql
SELECT * FROM payments;
```

In this example, since foreign key subsetting is disabled, Neosync applies the WHERE clauses independently to each table. The `addresses` and `customers` tables get their respective conditions, while the `orders` and `payments` tables remain unfiltered since no direct WHERE clauses were specified for them. This means all orders and payments will be included, regardless of their relationships to the filtered addresses and customers. Unless you are certain this won't cause foreign key violations in your target database, it's recommended to enable the "Skip Foreign Key Violations" option for jobs where foreign key subsetting is disabled.
