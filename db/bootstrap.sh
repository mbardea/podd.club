createdb podd

create role podd login password 'podd';
grant all privileges on database podd to podd;
grant all privileges on all tables in schema public to podd;

# in /etc/postgresql/9.1/main/pg_hba.conf:

# local    all             all                                     md5
