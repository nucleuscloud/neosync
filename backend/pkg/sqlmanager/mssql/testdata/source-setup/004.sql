CREATE PROCEDURE sp_PlaceOrder
    @CustomerID INT,
    @ProductID INT,
    @Quantity INT
AS
BEGIN
    SET NOCOUNT ON;
    
    DECLARE @CurrentStock INT;
    DECLARE @UnitPrice DECIMAL(10,2);
    
    SELECT @CurrentStock = StockQuantity, @UnitPrice = Price
    FROM Products
    WHERE ProductID = @ProductID;
    
    IF @CurrentStock < @Quantity
        THROW 50001, 'Insufficient stock', 1;
    
    BEGIN TRANSACTION;
    
    BEGIN TRY
        DECLARE @OrderID INT;
        INSERT INTO Orders (CustomerID, TotalAmount)
        VALUES (@CustomerID, @Quantity * @UnitPrice);
        
        SET @OrderID = SCOPE_IDENTITY();
        
        INSERT INTO OrderItems (OrderID, ProductID, Quantity, UnitPrice)
        VALUES (@OrderID, @ProductID, @Quantity, @UnitPrice);
        
        UPDATE Products
        SET StockQuantity = StockQuantity - @Quantity
        WHERE ProductID = @ProductID;
        
        COMMIT;
    END TRY
    BEGIN CATCH
        ROLLBACK;
        THROW;
    END CATCH;
END;
