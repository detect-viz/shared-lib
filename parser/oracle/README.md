
統一返回格式：
timestamp: Unix 時間戳
dbname: 數據庫名稱
status: 連接狀態 (success/failed)
value: 數值化狀態 (success = 0, failed = 1)


## [Alert] 每小時運行 ipoc_awrrpt.sh
- 檔名：`${DBName}_${SNAP_START}_${SNAP_END}_${YYYYMMDDHHMM}_${YYYYMMDDHHMM}.html`
- 傳送到 upload 資料夾
### 特別處理：
  - 內容全部寫入 `parser_awrrpt_files` 表中的 `content` 欄位
  - 每一個 html 檔有多張 table，有參數可以設定哪些 table 要解析
  - 參數設定在 `config.yml` 的 `awr_setting` 中
  - 例如 `global.EnvConfig.Server.AwrSetting.SqlStatistics` 設定是否解析 SQLStatistics 的 table
  ```yaml
    awr_setting:
    information: true
    sql_statistics: true
    tablespace_io_stats: true
    iostat: true
    undo_segment: true
    memory_statistics: true
    cache_sizes: true
    shared_pool_statistics: true
    instance_efficiency: true
    operating_system: true
    instance_activity: true
    wait_event_histogram: true
    buffer_pool_statistics: true
    segment_statistics: true
    default_tables: true
    database_summary: true
  ```
  - table 要處理成 `map[string]interface{}` ，並且 key 為 metric
  - 針對 SQLStatistics 的 SQLText 寫入 `parser_awrrpt_sqls` ，欄位有`SQLId`, `SQLText` ，因為sql會重複執行，所以 使用 `id` 當作主鍵，`id` 由 `SQLId` 組成，如果 `SQLId` 相同，則更新 `SQLText` ，如果 `SQLId` 不同，則新增 `SQLText` ，如果 `SQLId` 為空，則不處理
  -  `Container` 只有 19c 才有，可以濾掉

```sql
CREATE TABLE `parser_awrrpt_files` (
  `realm_name` varchar(20) NOT NULL,
  `host` varchar(50) NOT NULL,
  `db_name` varchar(50) NOT NULL,
  `date` date NOT NULL,
  `hour` tinyint NOT NULL,
  `content` mediumblob NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`realm_name`,`host`,`db_name`,`date`,`hour`),
  KEY `idx_awrrpt_files_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `parser_awrrpt_sqls` (
  `realm_name` varchar(20) NOT NULL,
  `db_name` varchar(50) NOT NULL,
  `id` varchar(255) NOT NULL,
  `text` longtext NOT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_awrrpt_sqls_realm` (`realm_name`),
  CONSTRAINT `fk_sqls_realms` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```


## [Alert] 每天早上 10 點運行 ipoc_tbs.sh
- 檔名：`${HOSTNAME}_${YYYYMMDDHHMM}_tableSpace.csv`
- 傳送到 upload 資料夾
- 可以指定欄位，並設定閾值，超過閾值、或是跟昨天(上一筆)相差，超過閾值即告警

```log
tablespace_name,autoextensible,files_in_tablespace,total_tablespace_space,total_used_space,total_tablespace_free_space,total_used_pct,total_free_pct,max_size_of_tablespace,max_free_size,total_auto_used_pct,total_auto_free_pct,date_time,sid,hostname
PSAPUNDO        , YES    ,                   1 ,                   7.32 ,              .17 ,                        7.15 ,           2.33 ,          97.67 ,                   9.77 ,          9.59 ,                1.75 ,               98.25 , 202408310000 , ADV                                                            
```


## [Alert] 每 15 分鐘運行 ipoc_conn.sh
- 檔名：`${HOSTNAME}_${YYYYMMDDHHMM}_ipoc_conn.err`
- 有連線錯誤才會紀錄到 .err 檔案，傳送到 upload 資料夾
  
```log
1703829605,testlog,failed
1703829605,scm,failed
1703829608,bugutf8,failed
1703829611,gpmdb,failed
1703829614,npiutf8,failed
1703829618,orcl,failed
1703829621,rmautf8,failed
1703829621,wind,failed
```

## [Report] 每天早上 10 點運行 ipoc_conn.sh
- 檔名：`${HOSTNAME}_${YYYYMMDDHHMM}_ipoc_conn.log`
- 檔案包含所有監控名單，無論成功失敗與否，傳送到 upload 資料夾
  
```log
1703827125,adv,success
1703827125,aqa,success
1703827125,pid,success
1703827126,pqa,success
1703827126,bid,success
1703827126,sol,failed
1703827129,testlog,failed
1703827129,scm,failed
1703827129,edv,success
1703827129,eqa,success
1703827129,asrsdb,failed
1703827129,bip,success
1703827132,bugutf8,failed
1703827132,cntrdb,success
1703827132,epd,success
1703827135,gpmdb,failed
1703827135,iebdb,success
1703827139,npiutf8,failed
1703827142,orcl,failed
1703827142,ppd,success
1703827145,rmautf8,failed
1703827145,twmesdb,success
1703827145,wind,failed
```
