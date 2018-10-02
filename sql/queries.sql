-- DROP TABLE griddy.t1;
CREATE TABLE T1 (
  key serial primary key not null, 
  value text unique
);

-- DROP TABLE griddy.t2;
CREATE TABLE T2 (
  key serial primary key not null,
  t1key integer REFERENCES t1(key) ON DELETE CASCADE,
  value text not null,
  createdate timestamp not null default current_timestamp
);

-- Inner Join
SELECT t1.key AS T1Key, t1.value AS T1Value, 
  t2.key AS T2Key, t2.value as T2Value, createdate
FROM griddy.t1
JOIN griddy.t2 ON (t1.key = t2.t1key);

-- Left Join
SELECT t1.key AS T1Key, t1.value AS T1Value, 
  t2.key AS T2Key, t2.value as T2Value, createdate
FROM griddy.t1
LEFT JOIN griddy.t2 ON (t2.t1key = t1.key);

-- Right Join
SELECT t1.key AS T1Key, t1.value AS T1Value, 
  t2.key AS T2Key, t2.value as T2Value, createdate
FROM griddy.t1
RIGHT JOIN griddy.t2 ON (t2.t1key = t1.key);