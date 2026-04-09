DROP TRIGGER IF EXISTS update_rag_settings_timestamp ON rag_settings;
DROP FUNCTION IF EXISTS update_rag_settings_timestamp;

DROP TABLE IF EXISTS rag_settings_history;
DROP TABLE IF EXISTS rag_settings;
