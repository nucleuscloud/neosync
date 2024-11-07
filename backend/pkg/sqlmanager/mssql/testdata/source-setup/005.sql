

CREATE PROCEDURE [sqlmanagermssql2].[ProcessOrder]
    @OrderDetails [sqlmanagermssql2].[OrderDetailType] READONLY
AS
BEGIN
    INSERT INTO OrderDetails (OrderLineId, ProductId, Quantity, UnitPrice)
    SELECT OrderLineId, ProductId, Quantity, UnitPrice
    FROM @OrderDetails;
END;
