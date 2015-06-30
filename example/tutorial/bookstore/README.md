# Tutorial

This tutorial shows you

## Prerequisites

* GoLang 1.x+ (at the time of writing I am using >=1.4.0 versions)
* Install qor:

    go get github.com/qor/qor

NB: I you clone qor and use it from your own account you will have to create a symlink like this:

    go/src/github.com/qor [qor] $ ln -s ../YOUR_GITHUB_USRNAME/qor/ qor

* A database - for example PostgreSQL or MySQL
* Install dependencies: cd into the qor source directory and run

    go get ./...

* Install Gin - QOR does not require gin, but we use :

    go get github.com/gin-gonic/gin

* [Optional: fresh](https://github.com/pilu/fresh) being installed:

    go get github.com/pilu/fresh

fresh is not necessary to use qor, but it will make your life easier when playing with the tutorial: it monitors for file changes and automatically recompiles your code every time something has changed.


### Create a database and a db user for the tutorial

Before we dive into our models we need to create a database:

    mysql> CREATE DATABASE qor_bookstore DEFAULT CHARACTER SET utf8mb4;
    Query OK, 1 row affected (0.16 sec)

    mysql> CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor'; -- some versions don't like this use the next line instead
    mysql> CREATE USER 'qor'@'%' IDENTIFIED BY 'qor';         -- some versions don't like this use the previous line instead
    Query OK, 0 rows affected (0.00 sec)

    mysql> GRANT ALL ON qor_bookstore.* TO 'qor'@'localhost';
    Query OK, 0 rows affected (0.00 sec)

    mysql> FLUSH PRIVILEGES;
    Query OK, 0 rows affected (0.00 sec)

You should now be able to connect to your database from the console like this:

    $ mysql -uqor -p --database qor_bookstore


## Get Started

We want to create a simple bookstore application. We will start by building a catalog of books and then later add a storefront. We will then add a staging environment so that editors can preview their changes and then publish them to a live system.

Later we will add L10n/I18n support and look at roles for the editorial process.

Continuous TODO: Add the next planned steps for the tutorial here.

### Create the basic models

We will need the following two models to start with:

* Author
* Book

The `Author` model is very simple:

    type Author struct {
	    gorm.Model
	    Id   int64
	    Name string
    }

All qor models "inherit" from `gorm.model`. (see https://github.com/jinzhu/gorm).
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

The only interesting part here is the gorm struct tag: `gorm:many2many:book_authors"`; It tells `gorm` to create a join table `book_authors`.

That's almost it: If you [look at](https://github.com/fvbock/qor/tree/master/example/tutorial/models.go) you can see an `init()` function at the end: It sets up a db connection and `DB.AutoMigrate(&Author{}, &Book{})` tells QOR to automatically create the tables for our models.

Let's start the tutorial app once to see what happens when models get auto-migrated.

    go/src/github.com/qor/qor/example/tutorial/bookstore [bookstore (master)] $ fresh


If you now check your db you would see something like this:

    mysql> show tables;
    +-------------------------+
    | Tables_in_qor_bookstore |
    +-------------------------+
    | authors                 |
    | book_authors            |
    | books                   |
    | users                   |
    +-------------------------+
    4 rows in set (0.00 sec)

    mysql> describe authors;
    +------------+--------------+------+-----+---------+----------------+
    | Field      | Type         | Null | Key | Default | Extra          |
    +------------+--------------+------+-----+---------+----------------+
    | id         | int(11)      | NO   | PRI | NULL    | auto_increment |
    | created_at | timestamp    | YES  |     | NULL    |                |
    | updated_at | timestamp    | YES  |     | NULL    |                |
    | deleted_at | timestamp    | YES  |     | NULL    |                |
    | name       | varchar(255) | YES  |     | NULL    |                |
    +------------+--------------+------+-----+---------+----------------+
    5 rows in set (0.00 sec)

As you can see QOR/gorm added an `id` field as well as timestamp fields to keep track of creation, modification, and deletion times. We can ignore this for now - the main point is that you create your models without a unique identifier - QOR/gorm will do this for you automatically. (TODO: @jinzhu please confirm)

NB: If you add new fields to your model they will get added to the database automatically with `DB.AutoMigrate` - deletions or *changes* of eg. the type will *not* be automigrated. (TODO: @jinzhu please confirm)


### Admin

If the bookstore app is not yet running, start it by running `fresh` in the bookstore directory:

    go/src/github.com/qor/qor/example/tutorial/bookstore [bookstore (master)] $ fresh

Go to http://localhost:9000/admin and you should see the main admin interface:

![admin](https://raw.github.com/qor/fvbock/master/images/admin.png)

The Menu at the top gets created by adding your models as resources to the admin:

	Admin := admin.New(&qor.Config{DB: &db})
	Admin.AddResource(
		&Author{},
		&admin.Config{Menu: []string{
			"Author Management"},
			Name: "Author",
		},
	)

you can see how the rest of the resources was added in [main.go](https://github.com/fvbock/qor/blob/master/example/tutorial/bookstore/main.go#L14:L50), the `db` object referenced here is set up in [models.go](https://github.com/fvbock/qor/blob/master/example/tutorial/bookstore/models.go#L62:L68)

Go ahead an add an author and then a book via the admin.

#### Meta Module - Controlling display and editable fields in the admin



### First frontend

QOR does not provide any templating or routing support - use whatever library is best fit for your needs. In this tutorial we will use [gin]()

#### List of Books


#### MediaLibrary - Adding product images

qor/media_library

`Base` is the low level object to deal with images offering cropping, resizing, and URL contruction for images.

#### Shopping

    INERT INTO users (name) VALUES ("admin");

#### First we need users that can register to buy books




#### Add Locales && translations


-- later


### Add customers (model)

### Add orders
