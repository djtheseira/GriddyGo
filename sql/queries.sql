-- DROP TABLE griddy.t1;
CREATE TABLE T1 (
  key serial primary key not null, 
  value text unique
);

-- DROP TABLE griddy.t2;
CREATE TABLE T2 (
  key serial primary key not null,
  t1key integer REFERENCES t1(key) ON DELETE RESTRICT,
  value text not null,
  createdate timestamp not null default current_timestamp
);


-- Inner Join
SELECT *
FROM t2
JOIN t1 ON (t1.key = t2.t1key)

-- Left Join
SELECT *
FROM t2
LEFT JOIN t1 ON (t2.t1key = t1.key)

-- Right Join
SELECT *
FROM t2
RIGHT JOIN t1 ON (t2.t1key = t1.key)