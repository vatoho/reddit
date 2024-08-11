SET NAMES utf8;
SET time_zone = '+00:00';
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

CREATE TABLE IF NOT EXISTS `users`(
                         `id` varchar(255) NOT NULL,
                         `username` varchar(255) NOT NULL UNIQUE,
                         `password` varchar(255) NOT NULL,
                         PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE IF NOT EXISTS `sessions`(
    `token` varchar(255) NOT NULL,
    `user_id` varchar(255) NOT NULL,
    PRIMARY KEY (`token`),
    FOREIGN KEY (`user_id`)  REFERENCES `users`(`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8;
