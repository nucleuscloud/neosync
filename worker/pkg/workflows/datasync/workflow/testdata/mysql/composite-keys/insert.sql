use composite;

INSERT INTO `orders` (order_id, product_id, order_date) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'd94f8bdc-6a8b-4d3b-8b1d-601c6c5d5d10', '2023-01-01'),
('1e79c7a0-57ab-4f92-8b7d-f2e8a5b730f4', '6a2e8e9e-7181-44e3-899f-6c8a58f2e007', '2023-01-02'),
('c8f3cba3-bbf3-4e09-928a-d4b589b7c4d2', '77e4c9d8-6353-407f-b2a1-fffd0c10b7b5', '2023-01-03'),
('f1348ad8-8f34-4f94-bd9d-30c7f451edb2', '62c9953c-3cb1-4b4d-a8e5-474a10e1f2db', '2023-01-04'),
('3b1c51c1-4c14-4d23-a5c6-c5f83437b5c7', '8b1f3a15-c4f7-47b3-9091-df6b92e90b8f', '2023-01-05'),
('a8b1a8f9-7574-448e-87c6-8b78b8a1ae34', 'cc4a0e1f-8b6d-46f5-b7d5-b48e5a82f237', '2023-01-06'),
('e2d6f173-6480-4d64-b8b6-4f3c96a9b5d6', 'd5e7b8d4-c9e8-44fa-8b7d-7f2a2e8e9c8d', '2023-01-07'),
('66a9f5b6-4e3c-4f5e-b8a4-5d6c7c8f5b7d', 'f5b6e3c4-7a8b-4d8e-8b1c-7c5f8e9e7d3f', '2023-01-08'),
('8b1f4a7c-3b1d-47e8-8b5d-4c8e7d8e7a1f', '4d5e8f9b-6c8d-4e3f-8b1c-5f6c8e9e7d3a', '2023-01-09'),
('9e7a8d6c-5d4e-4b8f-8b3f-6c8e7d8a4b5e', '8b7c5f6d-7e8a-4c3f-8b1e-7d6f8e9a7d4c', '2023-01-10');


INSERT INTO order_details (detail_id, order_id, product_id, quantity) VALUES
('2c6e7f8b-4e3d-4b5d-8b1e-6d7c8e9a5f4b', '550e8400-e29b-41d4-a716-446655440000', 'd94f8bdc-6a8b-4d3b-8b1d-601c6c5d5d10', 10),
('5f6c7d8e-4b3d-4c6e-8b1d-7d8e9a5c6b7a', '1e79c7a0-57ab-4f92-8b7d-f2e8a5b730f4', '6a2e8e9e-7181-44e3-899f-6c8a58f2e007', 20),
('7d6e8a9b-5c4d-4e3f-8b1e-9e8a5d7c6b4f', 'c8f3cba3-bbf3-4e09-928a-d4b589b7c4d2', '77e4c9d8-6353-407f-b2a1-fffd0c10b7b5', 30),
('8b7f5d6e-6a8b-4e3f-8b1c-7d5c6e8a9b4a', 'f1348ad8-8f34-4f94-bd9d-30c7f451edb2', '62c9953c-3cb1-4b4d-a8e5-474a10e1f2db', 40),
('9e7c6b5d-5f4e-4e3a-8b1d-6c8e9a7d5b4f', '3b1c51c1-4c14-4d23-a5c6-c5f83437b5c7', '8b1f3a15-c4f7-47b3-9091-df6b92e90b8f', 50),
('a8e9b5c4-7d6e-4e3f-8b1e-9a7d5c6e8b4f', 'a8b1a8f9-7574-448e-87c6-8b78b8a1ae34', 'cc4a0e1f-8b6d-46f5-b7d5-b48e5a82f237', 60),
('d7f8e9a6-4e3d-4b5f-8b1e-6c8e9a5f7b4a', 'e2d6f173-6480-4d64-b8b6-4f3c96a9b5d6', 'd5e7b8d4-c9e8-44fa-8b7d-7f2a2e8e9c8d', 70),
('7d8e9a5c-4b5d-4e3f-8b1d-9a6c7e5f8b4a', '66a9f5b6-4e3c-4f5e-b8a4-5d6c7c8f5b7d', 'f5b6e3c4-7a8b-4d8e-8b1c-7c5f8e9e7d3f', 80),
('8a7c6e5d-5f4e-4e3a-8b1e-6d8e9a7d5b4f', '8b1f4a7c-3b1d-47e8-8b5d-4c8e7d8e7a1f', '4d5e8f9b-6c8d-4e3f-8b1c-5f6c8e9e7d3a', 90),
('9e8a7c6d-5c4d-4e3f-8b1e-7d5c6e8a9b4a', '9e7a8d6c-5d4e-4b8f-8b3f-6c8e7d8a4b5e', '8b7c5f6d-7e8a-4c3f-8b1e-7d6f8e9a7d4c', 100);



INSERT INTO order_shipping (shipping_id, order_id, product_id, shipping_date) VALUES
('a6c8e9a7-5b4d-4e3f-8b1e-9d7c5f6e8a4a', '550e8400-e29b-41d4-a716-446655440000', 'd94f8bdc-6a8b-4d3b-8b1d-601c6c5d5d10', '2023-01-11'),
('b7e8a9c5-4d5f-4e3a-8b1e-6c9a5d7c8b4d', '1e79c7a0-57ab-4f92-8b7d-f2e8a5b730f4', '6a2e8e9e-7181-44e3-899f-6c8a58f2e007', '2023-01-12'),
('c8f9e7d5-5a4e-4e3f-8b1d-7c6a9d8e5b4a', 'c8f3cba3-bbf3-4e09-928a-d4b589b7c4d2', '77e4c9d8-6353-407f-b2a1-fffd0c10b7b5', '2023-01-13'),
('d9a7e8b6-6b4d-4e3f-8b1e-9d5c6e7a8b4a', 'f1348ad8-8f34-4f94-bd9d-30c7f451edb2', '62c9953c-3cb1-4b4d-a8e5-474a10e1f2db', '2023-01-14'),
('e7a8b9c6-4d5f-4e3a-8b1e-6d9a5c7e8b4d', '3b1c51c1-4c14-4d23-a5c6-c5f83437b5c7', '8b1f3a15-c4f7-47b3-9091-df6b92e90b8f', '2023-01-15'),
('f6c8a9b7-5b4d-4e3f-8b1e-9d7c5f6e8a4d', 'a8b1a8f9-7574-448e-87c6-8b78b8a1ae34', 'cc4a0e1f-8b6d-46f5-b7d5-b48e5a82f237', '2023-01-16'),
('g7b8a9c6-4d5f-4e3a-8b1e-6d9a5c7e8b4a', 'e2d6f173-6480-4d64-b8b6-4f3c96a9b5d6', 'd5e7b8d4-c9e8-44fa-8b7d-7f2a2e8e9c8d', '2023-01-17'),
('h8a9c7b6-5b4d-4e3f-8b1e-9d6a7c8e5b4d', '66a9f5b6-4e3c-4f5e-b8a4-5d6c7c8f5b7d', 'f5b6e3c4-7a8b-4d8e-8b1c-7c5f8e9e7d3f', '2023-01-18'),
('i7b8c9a6-4d5f-4e3a-8b1e-6d9a5c7e8b4d', '8b1f4a7c-3b1d-47e8-8b5d-4c8e7d8e7a1f', '4d5e8f9b-6c8d-4e3f-8b1c-5f6c8e9e7d3a', '2023-01-19'),
('j8c9a7b6-5b4d-4e3f-8b1e-9d6a7c8e5b4a', '9e7a8d6c-5d4e-4b8f-8b3f-6c8e7d8a4b5e', '8b7c5f6d-7e8a-4c3f-8b1e-7d6f8e9a7d4c', '2023-01-20');


INSERT INTO shipping_status (status_id, order_id, product_id, status, status_date) VALUES
('a1b2c3d4-e5f6-4789-abc1-23d456e78901', '550e8400-e29b-41d4-a716-446655440000', 'd94f8bdc-6a8b-4d3b-8b1d-601c6c5d5d10', 'Shipped', '2023-02-01'),
('b2c3d4e5-f6a7-4b89-bc1d-23e456f78901', '1e79c7a0-57ab-4f92-8b7d-f2e8a5b730f4', '6a2e8e9e-7181-44e3-899f-6c8a58f2e007', 'In Transit', '2023-02-02'),
('c3d4e5f6-a7b8-4c89-c1d2-34f56789012b', 'c8f3cba3-bbf3-4e09-928a-d4b589b7c4d2', '77e4c9d8-6353-407f-b2a1-fffd0c10b7b5', 'Delivered', '2023-02-03'),
('d4e5f6a7-b8c9-4d89-d1c2-45f67890123c', 'f1348ad8-8f34-4f94-bd9d-30c7f451edb2', '62c9953c-3cb1-4b4d-a8e5-474a10e1f2db', 'Shipped', '2023-02-04'),
('e5f6a7b8-c9d1-4e89-e1d2-56f78901234d', '3b1c51c1-4c14-4d23-a5c6-c5f83437b5c7', '8b1f3a15-c4f7-47b3-9091-df6b92e90b8f', 'In Transit', '2023-02-05'),
('f6a7b8c9-d1e2-4f89-f1d2-67f89012345e', 'a8b1a8f9-7574-448e-87c6-8b78b8a1ae34', 'cc4a0e1f-8b6d-46f5-b7d5-b48e5a82f237', 'Delivered', '2023-02-06'),
('a7b8c9d1-e2f3-4a89-g1d2-78f90123456f', 'e2d6f173-6480-4d64-b8b6-4f3c96a9b5d6', 'd5e7b8d4-c9e8-44fa-8b7d-7f2a2e8e9c8d', 'Shipped', '2023-02-07'),
('b8c9d1e2-f3a4-4b89-h1d2-89f01234567g', '66a9f5b6-4e3c-4f5e-b8a4-5d6c7c8f5b7d', 'f5b6e3c4-7a8b-4d8e-8b1c-7c5f8e9e7d3f', 'In Transit', '2023-02-08'),
('c9d1e2f3-a4b5-4c89-i1d2-90f12345678h', '8b1f4a7c-3b1d-47e8-8b5d-4c8e7d8e7a1f', '4d5e8f9b-6c8d-4e3f-8b1c-5f6c8e9e7d3a', 'Delivered', '2023-02-09'),
('d1e2f3a4-b5c6-4d89-j1d2-01f23456789i', '9e7a8d6c-5d4e-4b8f-8b3f-6c8e7d8a4b5e', '8b7c5f6d-7e8a-4c3f-8b1e-7d6f8e9a7d4c', 'Shipped', '2023-02-10');
