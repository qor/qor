DROP DATABASE IF EXISTS qor_bookstore;
CREATE DATABASE qor_bookstore DEFAULT CHARACTER SET utf8mb4;
CREATE USER 'qor'@'%' IDENTIFIED BY 'qor';         -- some versions don't like this use the previous line instead
-- CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor'; -- some versions don't like this use the next line instead
GRANT ALL ON qor_bookstore.* TO 'qor'@'localhost';
FLUSH PRIVILEGES;
