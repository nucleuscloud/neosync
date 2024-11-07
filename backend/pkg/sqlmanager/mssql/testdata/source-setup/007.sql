CREATE TRIGGER trg_UpdateOrderTotal
ON sqlmanagermssql2.OrderItems
AFTER INSERT, UPDATE, DELETE
AS
BEGIN
    SET NOCOUNT ON;
    
    UPDATE o
    SET TotalAmount = (
        SELECT SUM(Subtotal)
        FROM OrderItems
        WHERE OrderID = o.OrderID
    )
    FROM Orders o
    WHERE EXISTS (
        SELECT 1 
        FROM inserted i 
        WHERE i.OrderID = o.OrderID
    )
    OR EXISTS (
        SELECT 1 
        FROM deleted d 
        WHERE d.OrderID = o.OrderID
    );
END;
