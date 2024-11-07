CREATE VIEW vw_OrderSummary
AS
SELECT 
    o.OrderID,
    o.OrderNumber,
    c.FullName AS CustomerName,
    o.OrderDate,
    o.TotalAmount,
    o.Status,
    COUNT(oi.OrderItemID) AS TotalItems
FROM sqlmanagermssql2.Orders o
JOIN sqlmanagermssql2.Customers c ON o.CustomerID = c.CustomerID
JOIN sqlmanagermssql2.OrderItems oi ON o.OrderID = oi.OrderID
GROUP BY 
    o.OrderID, 
    o.OrderNumber,
    c.FullName,
    o.OrderDate,
    o.TotalAmount,
    o.Status;
