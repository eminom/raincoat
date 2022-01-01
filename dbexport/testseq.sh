#!/bin/bash

OUT=$(mktemp)

echo """

select * from dtu_op;
select count(*) from dtu_op;
select idx from dtu_op;

""" > ${OUT}

sqlite3 foo.db < ${OUT}