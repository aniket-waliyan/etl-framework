-- Create the sink table to store consolidated data
CREATE TABLE sink_table (
    customer_id INT PRIMARY KEY,
    shard_id INT NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    email VARCHAR(100),
    created_at TIMESTAMP,
    country VARCHAR(50),
    amount DECIMAL(10,2),
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 

-- Create tables for user connection data
CREATE TABLE IF NOT EXISTS user_connection_history (
    dealer_id VARCHAR(50),
    group_id VARCHAR(50),
    dealer_code VARCHAR(50),
    logon_logoff_time BIGINT,
    login_allowed INTEGER,
    success_failure SMALLINT,
    logon_logoff_flag CHAR(1),
    details VARCHAR(255),
    mode_of_connection INTEGER,
    connection_number INTEGER,
    entry_sequence INTEGER,
    oms_sequence_no BIGINT,
    session_id VARCHAR(100),
    source_table VARCHAR(50),
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (dealer_id, logon_logoff_time, entry_sequence)
);

CREATE TABLE IF NOT EXISTS user_connection_log (
    dealer_id VARCHAR(50),
    group_id VARCHAR(50),
    dealer_code VARCHAR(50),
    logon_logoff_time BIGINT,
    login_allowed INTEGER,
    success_failure SMALLINT,
    logon_logoff_flag CHAR(1),
    details VARCHAR(255),
    mode_of_connection INTEGER,
    connection_number INTEGER,
    entry_sequence INTEGER,
    oms_sequence_no BIGINT,
    session_id VARCHAR(100),
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (dealer_id, logon_logoff_time, entry_sequence)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_history_logon_time ON user_connection_history(logon_logoff_time);
CREATE INDEX IF NOT EXISTS idx_history_dealer ON user_connection_history(dealer_id);
CREATE INDEX IF NOT EXISTS idx_history_processed ON user_connection_history(processed_at);

CREATE INDEX IF NOT EXISTS idx_log_logon_time ON user_connection_log(logon_logoff_time);
CREATE INDEX IF NOT EXISTS idx_log_dealer ON user_connection_log(dealer_id);
CREATE INDEX IF NOT EXISTS idx_log_processed ON user_connection_log(processed_at); 