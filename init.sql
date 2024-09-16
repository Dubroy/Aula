-- init.sql

-- 创建 items 表
CREATE TABLE IF NOT EXISTS `items` (
  `itemid` int NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `item_name` varchar(255) NOT NULL,
  `category` varchar(255) NOT NULL,
  `item_type` varchar(255) NOT NULL,
  PRIMARY KEY (`itemid`)
) ENGINE=InnoDB;

-- 创建 users 表
CREATE TABLE IF NOT EXISTS `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB;
