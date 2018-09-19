# Quick Steps to Setup SQL Databases for Flux Testing

## Timescale

### Start an instance of Timescale

Username is `postgres`, password is `password`. Connect via `localhost:5432`. Full Golang DSN is `postgres://postgres:password@0.tcp.ngrok.io:13399/flux?sslmode=disable`.

```
$ docker run -d --name timescaledb -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password timescale/timescaledb
80cdb578d31d181777e49cab12f190cad8c5cfd9412d98ec1e40d33783df6045
```

### Create `flux` database

```
$ docker run -it --net=host --rm timescale/timescaledb psql -h localhost -U postgres
Password for user postgres:
psql (9.6.10)
Type "help" for help.

postgres=# CREATE DATABASE flux;
CREATE DATABASE
postgres=# \c flux
You are now connected to database "flux" as user "postgres".
flux=# CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
WARNING:
WELCOME TO
 _____ _                               _     ____________
|_   _(_)                             | |    |  _  \ ___ \
  | |  _ _ __ ___   ___  ___  ___ __ _| | ___| | | | |_/ /
  | | | |  _ ` _ \ / _ \/ __|/ __/ _` | |/ _ \ | | | ___ \
  | | | | | | | | |  __/\__ \ (_| (_| | |  __/ |/ /| |_/ /
  |_| |_|_| |_| |_|\___||___/\___\__,_|_|\___|___/ \____/
               Running version 0.12.0
For more information on TimescaleDB, please visit the following links:

 1. Getting started: https://docs.timescale.com/getting-started
 2. API reference documentation: https://docs.timescale.com/api
 3. How TimescaleDB is designed: https://docs.timescale.com/introduction/architecture

Note: TimescaleDB collects anonymous reports to better understand and assist our users.
For more information and how to disable, please see our docs https://docs.timescaledb.com/using-timescaledb/telemetry.

CREATE EXTENSION
flux=#
```

### Insert some test data

Keeping the previous `psql` session open, paste this into your console:

```
create table legacy (
  _time        TIMESTAMP NOT NULL,
  host        varchar(255),
  temperature float
);

SELECT create_hypertable('legacy', '_time');

insert into legacy (_time, host, temperature)
values (now(), 'dodger', random() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', random() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', random() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', random() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', random() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', random() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', random() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', random() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', random() * 10 + 90);

select *
from legacy;
```

### Query the data with Flux

```
$ go run _examples/fluxcli/main.go -q 'fromSQL(driverName:"postgres",dataSourceName:"postgres://postgres:password@0.tcp.ngrok.io:13399/flux?sslmode=disable",query:"SELECT * FROM legacy") |> range(start:-5000h) |> yield()' 
#datatype,string,long,dateTime:RFC3339,string,double,dateTime:RFC3339,dateTime:RFC3339
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,_time,host,temperature,_start,_stop
,,0,2018-09-19T22:00:20.751878Z,dodger,96.69449340552092,2018-02-23T14:00:43.896212Z,2018-09-19T22:00:43.896212Z
,,0,2018-09-19T22:00:20.771323Z,dodger,92.88152559660375,2018-02-23T14:00:43.896212Z,2018-09-19T22:00:43.896212Z
,,0,2018-09-19T22:00:20.784044Z,dodger,98.96036252845079,2018-02-23T14:00:43.896212Z,2018-09-19T22:00:43.896212Z
,,0,2018-09-19T22:00:20.799285Z,dodger,90.15380289871246,2018-02-23T14:00:43.896212Z,2018-09-19T22:00:43.896212Z
,,0,2018-09-19T22:00:20.813326Z,dodger,92.60355510748923,2018-02-23T14:00:43.896212Z,2018-09-19T22:00:43.896212Z
,,0,2018-09-19T22:00:20.826173Z,dodger,96.32819181773812,2018-02-23T14:00:43.896212Z,2018-09-19T22:00:43.896212Z
,,0,2018-09-19T22:00:20.840232Z,dodger,97.96776902396232,2018-02-23T14:00:43.896212Z,2018-09-19T22:00:43.896212Z
,,0,2018-09-19T22:00:20.855579Z,dodger,99.23656451050192,2018-02-23T14:00:43.896212Z,2018-09-19T22:00:43.896212Z
,,0,2018-09-19T22:00:20.869132Z,dodger,99.07267795410007,2018-02-23T14:00:43.896212Z,2018-09-19T22:00:43.896212Z
```

## MariaDB

### Start an instance of MariaDB

Root password is `password`. User is `mysql`, password is `password`. Connect via `localhost:3306`. Full Golang DSN is `mysql:password@tcp(localhost:3306)/flux?parseTime=true`.

```
$ docker run -d --name mariadb -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password -e MYSQL_USER=mysql -e MYSQL_PASSWORD=password -d mariadb
3d4023061aee9fa1f8660cb2dbefc2c3e0092aa94aa928a9a9069f982e6bdb6d
```

### Create `flux` database

```
$ docker exec -it mariadb mysql -u root -p
Enter password:
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 8
Server version: 10.3.9-MariaDB-1:10.3.9+maria~bionic mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE DATABASE flux;
Query OK, 1 row affected (0.000 sec)

MariaDB [(none)]> GRANT ALL ON flux.* TO 'mysql'@'%';
Query OK, 0 rows affected (0.000 sec)

MariaDB [(none)]> exit;
Bye
```

### Insert some test data

```
$ docker exec -it mariadb mysql -u mysql -p
Enter password:
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 12
Server version: 10.3.9-MariaDB-1:10.3.9+maria~bionic mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]>
```

Now paste this SQL into your console:

```sql
use flux;

create table legacy (
  _time       TIMESTAMP NOT NULL,
  host        TEXT,
  temperature float
);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', RAND() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', RAND() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', RAND() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', RAND() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', RAND() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', RAND() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', RAND() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', RAND() * 10 + 90);

insert into legacy (_time, host, temperature)
values (now(), 'dodger', RAND() * 10 + 90);

select *
from legacy;
```

### Query the data with Flux

```
$ go run _examples/fluxcli/main.go -q 'fromSQL(driverName:"mysql",dataSourceName:"mysql:password@tcp(localhost:3306)/flux?parseTime=true",query:"SELECT * FROM legacy") |> range(start:-5000h) |> yield()' 
#datatype,string,long,dateTime:RFC3339,string,string,dateTime:RFC3339,dateTime:RFC3339
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,_time,host,temperature,_start,_stop
,,0,2018-09-19T22:29:02Z,dodger,91.8406,2018-02-23T14:32:27.62447Z,2018-09-19T22:32:27.62447Z
,,0,2018-09-19T22:29:02Z,dodger,95.8584,2018-02-23T14:32:27.62447Z,2018-09-19T22:32:27.62447Z
,,0,2018-09-19T22:29:02Z,dodger,93.7704,2018-02-23T14:32:27.62447Z,2018-09-19T22:32:27.62447Z
,,0,2018-09-19T22:29:02Z,dodger,91.2769,2018-02-23T14:32:27.62447Z,2018-09-19T22:32:27.62447Z
,,0,2018-09-19T22:29:02Z,dodger,95.0731,2018-02-23T14:32:27.62447Z,2018-09-19T22:32:27.62447Z
,,0,2018-09-19T22:29:02Z,dodger,91.535,2018-02-23T14:32:27.62447Z,2018-09-19T22:32:27.62447Z
,,0,2018-09-19T22:29:02Z,dodger,92.4556,2018-02-23T14:32:27.62447Z,2018-09-19T22:32:27.62447Z
,,0,2018-09-19T22:29:02Z,dodger,97.6731,2018-02-23T14:32:27.62447Z,2018-09-19T22:32:27.62447Z
,,0,2018-09-19T22:29:02Z,dodger,90.9985,2018-02-23T14:32:27.62447Z,2018-09-19T22:32:27.62447Z
```