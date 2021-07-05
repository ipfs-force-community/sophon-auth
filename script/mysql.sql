CREATE DATABASE `venus_auth` /*!40100 DEFAULT CHARACTER SET utf8 */;
USE `venus_auth`;

# 如果出现错误提示:ERROR 1071 (42000): Specified key was too long; max key length is 767 bytes”
# 请添运行以下代码:
# set global innodb_large_prefix = 1;
# set global innodb_file_format = BARRACUDA;

CREATE TABLE `token`
(
    `name`       varchar(50)  NOT NULL,
    `token`      varchar(512) NOT NULL,
    `createTime` datetime     NOT NULL,
    `perm`       varchar(50)  NOT NULL,
    `extra`      varchar(255) DEFAULT NULL,
    UNIQUE KEY `token_token_IDX` (`token`) USING HASH
)
#   如果出现错误提示:ERROR 1071 (42000): Specified key was too long; max key length is 767 bytes”
#   运行时添加:
#   ROW_FORMAT = DYNAMIC
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8;


CREATE TABLE `users`
(
    `id`         varchar(255) NOT NULL,
    `name`       varchar(255) NOT NULL,
    `miner`      varchar(255) NOT NULL,
    `state`      tinyint(4)   NOT NULL DEFAULT '0',
    `comment`    varchar(255) NOT NULL,
    `createTime` datetime     NOT NULL,
    `updateTime` datetime     NOT NULL,
    `stype`      tinyint(4)   NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `users_name_IDX` (`name`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;
