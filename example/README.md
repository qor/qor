# How to setup and run QOR example

1. Setup database.

    1. CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor';
    2. CREATE DATABASE qor_example;
    3. GRANT ALL PRIVILEGES ON qor_example.* TO 'qor'@'localhost';

2. Start project

    `cd example && go get && RUNNER_EXTRA_DIRS="$GOPATH/src/github.com/qor,$GOPATH/src/github.com/jinzhu" fresh`
