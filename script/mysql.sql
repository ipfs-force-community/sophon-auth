CREATE DATABASE `auth`;
USE `auth`;
-- auth.token definition

CREATE TABLE `token` (
                         `token` varchar(512) NOT NULL,
                         `createTime` datetime NOT NULL,
                         UNIQUE KEY `NewTable_token_IDX` (`token`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
