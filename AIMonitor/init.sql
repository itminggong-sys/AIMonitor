-- AI Monitor 数据库初始化脚本

-- 创建UUID扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建时间扩展
CREATE EXTENSION IF NOT EXISTS "btree_gist";

-- 设置时区
SET timezone = 'UTC';

-- 创建数据库用户权限
GRANT ALL PRIVILEGES ON DATABASE ai_monitor TO ai_monitor;
GRANT ALL PRIVILEGES ON SCHEMA public TO ai_monitor;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ai_monitor;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ai_monitor;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO ai_monitor;

-- 设置默认权限
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ai_monitor;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ai_monitor;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO ai_monitor;

-- 创建一些有用的函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 输出初始化完成信息
SELECT 'AI Monitor database initialized successfully' AS status;