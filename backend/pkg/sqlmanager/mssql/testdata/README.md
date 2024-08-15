# MSSQL Test Data

You'll note a somewhat homegrown migration file type setup here.
Mssql does not support multiple statements via the Go Driver, so each statement must be done in isolation.

As a result, any time a new statement must be added, create a new file with the number that you wish for it to run and add that statement there.
This will result in the system executing it in that order.
