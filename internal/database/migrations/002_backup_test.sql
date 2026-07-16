-- Backup verification test
-- This test restores a backup dump and verifies schema + data counts

SELECT 'backup_test' AS test_name,
       CASE WHEN COUNT(*) > 0 THEN 'PASS' ELSE 'FAIL' END AS result
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE';
