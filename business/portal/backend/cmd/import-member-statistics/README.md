# 会员数据中心 Excel 导入

此命令只处理 Excel 的「总表」，将年度元数据写入 `member_statistics_dataset`，将每一条原始样本写入 `member_statistics_row`。现有 `member_content` 不受影响。后端正常启动时也会通过 `AutoMigrate` 创建这两张表。

在 `business/portal/backend` 目录运行。先预览：

```powershell
go run ./cmd/import-member-statistics -file 'C:\WebstormProjects\环保装备制造业重点企业数据统计(2023-2024).xlsx' -year 2024 -compare-year 2023
```

核对预览结果后，显式加 `-apply` 才连接 `config.yaml` 指向的 MySQL 并写入：

如果 `config.yaml` 中的数据库密码仍是占位符，先在**当前 PowerShell 会话**设置 `PORTAL_DB_PASSWORD`。导入命令只需要数据库密码，不需要门户 JWT 密钥；不要把密码写入命令行参数或提交到仓库。

```powershell
$env:PORTAL_DB_PASSWORD = Read-Host 'MySQL 密码' -MaskInput
```

```powershell
go run ./cmd/import-member-statistics -file 'C:\WebstormProjects\环保装备制造业重点企业数据统计(2023-2024).xlsx' -year 2024 -compare-year 2023 -apply
```

如目标会员栏目名称不是「数据中心」，可添加 `-column '实际栏目名'`。首次导入要求目标栏目已存在且已启用。脚本不会创建会员栏目或发布数据；数据集状态为草稿。

导入前会校验表头、必填分类、金额和同比。重复企业名称会列出原始 Excel 行号，但不会擅自删除其中一条：两条都保留在明细表，预览同时区分「记录数」和「去重企业数」。研发投入空白保留为数据库 `NULL`。收入以 Excel 的万元单位保存，展示时再换算亿元。

同一会员栏目、同一统计年度再次执行 `-apply` 会在单个事务中替换明细并重置为草稿；任何写入失败都会回滚。导入完成后，需在门户后台的会员专区「年度统计」中发布数据集，会员端「数据中心」才会展示。

后台管理页面已提供查看与发布操作：重启 portal 后端并启动/构建后台前端，在「内容管理 → 会员专区」点击「年度统计」。这里可查看数据集汇总、重复企业与研发缺失提示，筛选分页查看原始明细，并发布或下线数据集。会员门户看板尚未接入，因此切换发布状态目前不会自动生成会员端图表。
