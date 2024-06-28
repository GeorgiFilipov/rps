-- init.sql

-- Switch to the 'elysium' database
\c elysium;

-- Create table 'player'
CREATE TABLE IF NOT EXISTS player (
                                      id SERIAL PRIMARY KEY,
                                      username VARCHAR(255) NOT NULL UNIQUE,
                                      password VARCHAR(255) NOT NULL,
                                      salt VARCHAR(255) NOT NULL,
                                      balance INTEGER NOT NULL
);

-- Alter table 'player' owner to 'postgres'
ALTER TABLE player OWNER TO postgres;

-- Create table 'challenge'
CREATE TABLE IF NOT EXISTS challenge (
                                         challenge_id SERIAL PRIMARY KEY,
                                         challenger VARCHAR(255) NOT NULL,
                                         opponent VARCHAR(255) NOT NULL,
                                         choice INTEGER NOT NULL,
                                         bet INTEGER NOT NULL,
                                         state VARCHAR(50) NOT NULL,
                                         time_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                         time_settled TIMESTAMP,
                                         winner VARCHAR
);

-- Alter table 'challenge' owner to 'postgres'
ALTER TABLE challenge OWNER TO postgres;

-- Create table 'transaction'
CREATE TABLE IF NOT EXISTS transaction (
                                           id SERIAL PRIMARY KEY,
                                           timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                           amount INTEGER NOT NULL,
                                           reason TEXT NOT NULL,
                                           username VARCHAR(15) NOT NULL
);

-- Alter table 'transaction' owner to 'postgres'
ALTER TABLE transaction OWNER TO postgres;
