DROP TABLE IF EXISTS blocks;
;; TODO: 
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
