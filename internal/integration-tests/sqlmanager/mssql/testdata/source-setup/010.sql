CREATE TRIGGER [mssqlinit].tr_TestTable_Update
ON [mssqlinit].TestTable
WITH ENCRYPTION
AFTER UPDATE
AS
BEGIN
    PRINT 'Update Trigger Fired';
    -- Add update-related logic here.
END;
