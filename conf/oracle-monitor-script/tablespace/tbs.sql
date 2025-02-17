spool &1

SET lines 300 pages
SET pagesize 0
SET linesize 1000
SET trimout on
SET trimspool on
SET newp none
SET echo off
SET feed off
SET underline off
SET feedback off
TTITLE off

COLUMN tablespace_name              format a15
COLUMN autoextensible               format a6
COLUMN files_in_tablespace          format 99
COLUMN total_tablespace_space       format 9999.99
COLUMN total_used_space             format 9999.99
COLUMN total_tablespace_free_space  format 9999.99
COLUMN total_used_pct               format 99.99
COLUMN total_free_pct               format 99.99
COLUMN max_size_of_tablespace       format 9999.99
COLUMN max_free_size                format 9999.99
COLUMN total_auto_used_pct          format 99.99
COLUMN total_auto_free_pct          format 999.99


WITH tbs_auto AS
     (SELECT DISTINCT tablespace_name, autoextensible
                 FROM dba_data_files
                WHERE autoextensible = 'YES'),
     files AS
     (SELECT   tablespace_name, COUNT (*) tbs_files,
               SUM (BYTES) total_tbs_bytes
          FROM dba_data_files
      GROUP BY tablespace_name),
     fragments AS
     (SELECT   tablespace_name, COUNT (*) tbs_fragments,
               SUM (BYTES) total_tbs_free_bytes,
               MAX (BYTES) max_free_chunk_bytes
          FROM dba_free_space
      GROUP BY tablespace_name),
     AUTOEXTEND AS
     (SELECT   tablespace_name, SUM (size_to_grow) total_growth_tbs
          FROM (SELECT   tablespace_name, SUM (maxbytes) size_to_grow
                    FROM dba_data_files
                   WHERE autoextensible = 'YES'
                GROUP BY tablespace_name
                UNION
                SELECT   tablespace_name, SUM (BYTES) size_to_grow
                    FROM dba_data_files
                   WHERE autoextensible = 'NO'
                GROUP BY tablespace_name)
      GROUP BY tablespace_name)
SELECT
       a.tablespace_name,',',
       CASE tbs_auto.autoextensible
          WHEN 'YES'
             THEN 'YES'
          ELSE 'NO'
       END AS autoextensible,',',
       files.tbs_files files_in_tablespace,',',
       files.total_tbs_bytes/1024/1024/1024 total_tablespace_space,',',
       (files.total_tbs_bytes - fragments.total_tbs_free_bytes)/1024/1024/1024 total_used_space,',',
      fragments.total_tbs_free_bytes/1024/1024/1024 total_tablespace_free_space,',',
     (  (  (files.total_tbs_bytes - fragments.total_tbs_free_bytes) / files.total_tbs_bytes) * 100) total_used_pct,',',
       ((fragments.total_tbs_free_bytes / files.total_tbs_bytes) * 100) total_free_pct,',',
       AUTOEXTEND.total_growth_tbs/1024/1024/1024 max_size_of_tablespace,',',
       (AUTOEXTEND.total_growth_tbs- (files.total_tbs_bytes - fragments.total_tbs_free_bytes ))/1024/1024/1024 max_free_size,',',
       ((files.total_tbs_bytes - fragments.total_tbs_free_bytes)/AUTOEXTEND.total_growth_tbs)*100 total_auto_used_pct,',',
       ((AUTOEXTEND.total_growth_tbs- (files.total_tbs_bytes - fragments.total_tbs_free_bytes ))/AUTOEXTEND.total_growth_tbs)*100 total_auto_free_pct,',',
        to_char(sysdate,'yyyymmddHH24MI'),',',
       (select sys_context('USERENV','DB_NAME') as Instance from dual),',',
       (select host_name from v$instance)
  FROM dba_tablespaces a, files, fragments, AUTOEXTEND, tbs_auto
   WHERE a.tablespace_name = files.tablespace_name
   AND a.tablespace_name = fragments.tablespace_name
   AND a.tablespace_name = AUTOEXTEND.tablespace_name
   AND a.tablespace_name = tbs_auto.tablespace_name(+);

spool off
exit
