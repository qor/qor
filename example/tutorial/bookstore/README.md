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

#### PostgreSQL

    sudo su - postgres
    postgres@lain:~$ psql
    psql (9.4.4)
    Type "help" for help.

    postgres=# create database qor_bookstore;
    CREATE DATABASE

    postgres=# \c template1
    You are now connected to database "template1" as user "postgres".
    template1=# CREATE USER qor WITH PASSWORD 'qor';
    template1=# GRANT all ON DATABASE qor_bookstore TO qor;


#### MySQL

    mysql> DROP DATABASE IF EXISTS qor_bookstore;
    mysql> CREATE DATABASE qor_bookstore DEFAULT CHARACTER SET utf8mb4;

    mysql> CREATE USER 'qor'@'%' IDENTIFIED BY 'qor';         -- some versions don't like this use the previous line instead

    OR

    mysql> CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor'; -- some versions don't like this use the next line instead

    mysql> GRANT ALL ON qor_bookstore.* TO 'qor'@'localhost';
    mysql> FLUSH PRIVILEGES;

You should now be able to connect to your database from the console like this:

    $ mysql -uqor -p --database qor_bookstore


## Get Started

We want to create a simple bookstore application. We will start by building a catalog of books and then later add a storefront. We will then add a staging environment so that editors can preview their changes and then publish them to a live system.

Later we will add L10n/I18n support and look at roles for the editorial process.

Continuous TODO: Add the next planned steps for the tutorial here.

### The basic models

We will need the following two models to start with:

* Author
* Book

The `Author` model is very simple:

    type Author struct {
	    gorm.Model
	    Name string
        publish.Status
    }

All qor models "inherit" from `gorm.model`. (see https://github.com/jinzhu/gorm).
Our author model for now only has a `Name`.
Ignore the `publish.Status` for now - we will address that in a later part of the tutorial.

The Bookmodel has a few more fields:

    type Book struct {
    	gorm.Model
	    Title       string
	    Synopsis    string
	    ReleaseDate time.Time
	    Authors     []*Author `gorm:"many2many:book_authors"`
	    Price       float64
	    CoverImage  media_library.FileSystem
        publish.Status
    }

The only interesting part here is the gorm struct tag: `gorm:many2many:book_authors"`; It tells `gorm` to create a join table `book_authors`.

Ignore the `publish.Status` for now - we will address that in a later part of the tutorial.

That's almost it: If you [look at](https://github.com/qor/qor/tree/master/example/tutorial/bookstore/01/models.go) you can see an `init()` function at the end: It sets up a db connection and `db.AutoMigrate(&Author{}, &Book{}, &User{})` tells QOR to automatically create the tables for our models.

You can ignore the user model for now - we will look at that part later.

Let's start the tutorial app once to see what happens when models get auto-migrated.

    go/src/github.com/qor/qor/example/tutorial/bookstore/01 [01 (master)] $ fresh

or if you don't want to use fresh you can build and run the app:

    /go/src/github.com/qor/qor/example/tutorial/bookstore/01 [01 (master)] $ go build -o tutorial main.go
    /go/src/github.com/qor/qor/example/tutorial/bookstore/01 [01 (master)] $ ./tutorial

If you now check your db you would see something like this:

#### PostgreSQL



#### MySQL

    mysql> show tables;
    +-------------------------+
    | Tables_in_qor_bookstore |
    +-------------------------+
    | authors                 |
    | authors_draft           |
    | book_authors            |
    | book_authors_draft      |
    | books                   |
    | books_draft             |
    | translations            |
    | users                   |
    +-------------------------+
    8 rows in set (0.00 sec)

    mysql> describe authors;
    +----------------+------------------+------+-----+---------+----------------+
    | Field          | Type             | Null | Key | Default | Extra          |
    +----------------+------------------+------+-----+---------+----------------+
    | id             | int(10) unsigned | NO   | PRI | NULL    | auto_increment |
    | created_at     | timestamp        | YES  |     | NULL    |                |
    | updated_at     | timestamp        | YES  |     | NULL    |                |
    | deleted_at     | timestamp        | YES  |     | NULL    |                |
    | publish_status | tinyint(1)       | YES  |     | NULL    |                |
    | language_code  | varchar(6)       | NO   | PRI |         |                |
    | name           | varchar(255)     | YES  |     | NULL    |                |
    +----------------+------------------+------+-----+---------+----------------+
    7 rows in set (0.00 sec)

As you can see QOR/gorm added an `id` field as well as timestamp fields to keep track of creation, modification, and deletion times. We can ignore this for now - the main point is that you create your models without a unique identifier - QOR/gorm will do this for you automatically. (TODO: @jinzhu please confirm)

NB: If you add new fields to your model they will get added to the database automatically with `DB.AutoMigrate` - deletions or *changes* of eg. the type will *not* be automigrated. (TODO: @jinzhu please confirm)


### Admin

If the bookstore app is not yet running, start it by running `fresh` in the bookstore directory:

    go/src/github.com/qor/qor/example/tutorial/bookstore/01 [bookstore (master)] $ fresh

Go to http://localhost:9000/admin and you should see the main admin interface:

TODO: add screenshot

The menu at the top gets created by adding your models as resources to the admin in [main.go](https://github.com/qor/qor/blob/docs_and_tutorial/example/tutorial/bookstore/01/main.go):

	Admin := admin.New(&qor.Config{DB: &db})

	Admin.AddResource(
		&User{},
		&admin.Config{
			Menu: []string{"User Management"},
			Name: "Users",
		},
	)

you can see how the rest of the resources was added in [resources.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/resources.go), the `db` object referenced here is set up in [models.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/models/models.go#L66:L72)

Go ahead and go to the authors admin and add an author...

TODO: add screenshots

... and then a book via the admin:

TODO: add screenshots

#### Meta Module - Controlling display and editable fields in the admin

Go to http://localhost:9000/admin/books.
Now comment the following line from [resources.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/resources.go)

	book.IndexAttrs("ID", "Title", "AuthorNames", "FormattedDate", "DisplayPrice")

and reload the books admin page.

TODO: add screenshot

You will see a much more crowded list: We had 5 attributes `"ID", "Title", "AuthorNames", "FormattedDate", and "DisplayPrice"`. `Id`, `Title`, and `ReleaseDate` are defined on our `Book` model, but the other two are not. For the Authors field you only see a list of References to the `Author` objects - something like `[0xc208161bc0]`. We want the list of Authors devided by `,` instead. You can achieve that by defining a `Meta` field:

	book.Meta(&admin.Meta{
		Name:  "AuthorNames",
		Label: "Authors",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			if value == nil {
				return value
			}
			book := value.(*Book)
			if err := db.Model(&book).Related(&book.Authors, "Authors").Error; err != nil {
				panic(err)
			}

			log.Println(book.Authors)
			var authors string
			for i, author := range book.Authors {
				if i >= 1 {
					authors += ", "
				}
				authors += author.Name
			}
			return authors
		},
	})

We define a `Meta` field for the `book` `Resource`. It's internal name is "AuthorNames", which we use in `book.IndexAttrs()` to use it in our admin book listing. The "Label` is what goes into the table header and the "Valuer" is a function that will return the display value we want - in our case the comma separated list of author names.

##### Editable fields

By default all defined model attributes and `Meta` attributes are included in the edit interface. If you need to limit the fields that are editable you can manually set the `EditAttrs`:

	book.EditAttrs("Title", "Authors", "Synopsis", "ReleaseDate", "Price", "CoverImage")

#### Searchable Fields

To get a searchfield on the list display of your resource you simply add a line like this:

    book.SearchAttrs("ID", "Title")

Wich will add a search(field) for resources matching on the defined fields.

#### Meta Field Types

QOR will pick an input type based on your struct types - but sometimes you want to change the default. For example we might want to have a text area with some editing functions instead of just an `<input type="text">`:

	book.Meta(&admin.Meta{
		Name: "Synopsis",
		Type: "rich_editor",
	})

TODO: other types - at least select_one and select_many. add list.


### Frontend

QOR does not provide any builtin templating or routing support - use whatever library is best fit for your needs. In this tutorial we will use [gin](https://github.com/gin-gonic/gin) and the stl `html/template`s.

#### Books Listing

#### Product Page



#### MediaLibrary - Adding product images

qor/media_library

`Base` is the low level object to deal with images offering cropping, resizing, and URL contruction for images.

INERT INTO users (name) VALUES ("admin");


#### I18n and L10n
