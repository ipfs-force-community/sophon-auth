CREATE DATABASE `venus_auth` /*!40100 DEFAULT CHARACTER SET utf8 */;
USE `venus_auth`;

CREATE TABLE `token` (
     `name` varchar(50) NOT NULL,
     `token` varchar(512) NOT NULL,
     `createTime` datetime NOT NULL,
     `perm` varchar(50) NOT NULL,
     `extra` varchar(255) DEFAULT NULL,
     UNIQUE KEY `token_token_IDX` (`token`) USING HASH
) ENGINE=InnoDB
  DEFAULT CHARSET = utf8
  COLLATE = utf8_general_ci;
