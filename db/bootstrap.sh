createdb poddpadd

create role poddpadd login password 'poddpadd';
grant all privileges on database poddpadd to poddpadd;
grant all privileges on all tables in schema public to poddpadd;

# in /etc/postgresql/9.1/main/pg_hba.conf:

# local    all             all                                     md5
