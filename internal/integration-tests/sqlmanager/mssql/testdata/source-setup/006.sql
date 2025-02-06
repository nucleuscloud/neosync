CREATE PROCEDURE [mssqlinit].[ProcessOrder]
    @OrderDetails [mssqlinit].[OrderDetailType] READONLY
AS
BEGIN
    INSERT INTO OrderDetails (OrderLineId, ProductId, Quantity, UnitPrice)
    SELECT OrderLineId, ProductId, Quantity, UnitPrice
    FROM @OrderDetails;
END;
