# qor tutorial

## Prerequisites

* GoLang 1.x+ (?)

* Install qor `git clone https://github.com/qor/qor.git`

NB: I you clone qor and use it from your own account you will have to create a symlink like this:

    go/src/github.com/qor [qor] $ ln -s ../YOUR_GITHUB_USRNAME/qor/ qor

* A database - for example PostgreSQL or MySQL

* [Pak](https://github.com/theplant/pak) being installed:

    go get github.com/theplant/pak

* go into the qor source directory and run `pak get` - this will install all dependencies qor itself needs to run

## A bookstore

We want to create a simple bookstore application. We will start by building a catalog of books and then later add a storefront that can take orders from users that can then be processed in a backoffice application by the store owner.

### Create a database and a db user for the tutorial

    mysql> CREATE DATABASE qor_bookstore DEFAULT CHARACTER SET utf8mb4;
    Query OK, 1 row affected (0.16 sec)

    mysql> use qor_bookstore
    Database changed

    mysql> CREATE USER 'qor_tutorial'@'%' IDENTIFIED BY 'qor_tutorial';
    Query OK, 0 rows affected (0.00 sec)

    mysql> GRANT ALL ON qor_bookstore.* TO 'qor_tutorial'@'localhost';
    Query OK, 0 rows affected (0.00 sec)

    mysql> FLUSH PRIVILEGES;
    Query OK, 0 rows affected (0.00 sec)

You should now be able to connect to your database from the console like this:

    $ mysql -uqor_tutorial -p --database qor_bookstore

### Create the basic models

We will need the following two models to start with:

Before we dive into our models we need to create a database that

* Author
* Book

The `Author` model is very simple:

    type Author struct {
	    gorm.Model
	    Id   int64
	    Name string
    }

All qor models "inherit" from `gorm.model`. (TODO: add some gorm info/link).
Our author model for now only has an `Id` and a `Name`.

The Bookmodel has a few more fields:

    type Book struct {
    	gorm.Model
    	Id          int64
    	Title       string
    	Synopsis    string
    	ReleaseDate time.Time
    	Authors     []*Author `gorm:"many2many:book_authors"`
    	Price       float64
    }

The only interesting part here is the gorm struct tag: `gorm:"one2many:authors"` ... TODO

That's almost it: If you [look at](https://github.com/fvbock/qor/tree/master/example/tutorial/models.go) you can see an `init()` function at the end

...

### Admin

insert screenshots

### First frontend



#### List of Books

### Add customers (model)

### Add orders
