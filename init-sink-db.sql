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