CREATE DATABASE `venus_auth` /*!40100 DEFAULT CHARACTER SET utf8mb4 */;
USE `venus_auth`;
CREATE TABLE `token`
(
    `name`       varchar(50)  NOT NULL,
    `token`      varchar(512) NOT NULL,
    `createTime` datetime     NOT NULL,
    UNIQUE KEY `token_token_IDX` (`token`) USING HASH
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
