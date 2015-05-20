## How to run QOR integration test

### Install required drivers

You need execute-able chromedriver in your PATH.

### Prepare mysql database

    CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor';
    CREATE DATABASE qor_integration_test;
    GRANT ALL PRIVILEGES ON qor_integration_test.* TO 'qor'@'localhost';

Then run `go test`.

### References

QOR integration test is written with [agouti](http://agouti.org/) and [gomega](http://onsi.github.io/gomega/)
