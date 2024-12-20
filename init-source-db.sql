-- Create sample customer table and insert data for each shard
CREATE DATABASE my_source_db;
GO
USE my_source_db;
GO

CREATE TABLE source_table (
    customer_id INT PRIMARY KEY,
    shard_id INT NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    email VARCHAR(100),
    created_at DATETIME DEFAULT GETDATE(),
    country VARCHAR(50),
    amount DECIMAL(10,2)
);
GO

-- Shard 1 Data (customers from USA)
INSERT INTO source_table (customer_id, shard_id, first_name, last_name, email, country, amount)
VALUES 
    (1, 1, 'John', 'Doe', 'john.doe@email.com', 'USA', 1500.50),
    (2, 1, 'Jane', 'Smith', 'jane.smith@email.com', 'USA', 2200.75),
    (3, 1, 'Mike', 'Johnson', 'mike.j@email.com', 'USA', 750.25);

-- Shard 2 Data (customers from Canada)
INSERT INTO source_table (customer_id, shard_id, first_name, last_name, email, country, amount)
VALUES 
    (4, 2, 'Sarah', 'Wilson', 'sarah.w@email.com', 'Canada', 1800.00),
    (5, 2, 'Robert', 'Brown', 'robert.b@email.com', 'Canada', 3200.50),
    (6, 2, 'Emma', 'Davis', 'emma.d@email.com', 'Canada', 950.75);

-- Shard 3 Data (customers from UK)
INSERT INTO source_table (customer_id, shard_id, first_name, last_name, email, country, amount)
VALUES 
    (7, 3, 'James', 'Taylor', 'james.t@email.com', 'UK', 2100.25),
    (8, 3, 'Oliver', 'White', 'oliver.w@email.com', 'UK', 1600.00),
    (9, 3, 'Sophie', 'Clark', 'sophie.c@email.com', 'UK', 2800.50);

-- Shard 4 Data (customers from Australia)
INSERT INTO source_table (customer_id, shard_id, first_name, last_name, email, country, amount)
VALUES 
    (10, 4, 'William', 'Lee', 'william.l@email.com', 'Australia', 1750.75),
    (11, 4, 'Lucy', 'Anderson', 'lucy.a@email.com', 'Australia', 2900.25),
    (12, 4, 'Thomas', 'Martin', 'thomas.m@email.com', 'Australia', 1300.50); 