DROP TABLE IF EXISTS blocks;
-- TODO: Restructure into multiple tables, to have better encapsulation of block, action log, begin/end block, etc.
CREATE TABLE blocks (
    id INTEGER PRIMARY KEY,
    height INT,
    module TEXT, 
    name TEXT, 
    comment TEXT, 
    passed BOOL, 
    gasWanted INT, 
    gasUsed INT, 
    msg STRING, 
    resData STRING, 
    appHash STRING);
