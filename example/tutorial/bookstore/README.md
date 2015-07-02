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

    go/src/github.com/qor/qor/example/tutorial/bookstore/01 [bookstore (master)] $ fresh

Go to http://localhost:9000/admin and you should see the main admin interface:

TODO: add screenshot

The Menu at the top gets created by adding your models as resources to the admin in [main.go](https://github.com/qor/qor/blob/docs_and_tutorial/example/tutorial/bookstore/01/main.go):

	Admin := admin.New(&qor.Config{DB: &db})

	Admin.AddResource(
		&User{},
		&admin.Config{
			Menu: []string{"User Management"},
			Name: "Users",
		},
	)

you can see how the rest of the resources was added in [main.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/main.go#L32:L46), the `db` object referenced here is set up in [models.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/models.go#L66:L72)

Go ahead and go to the authors admin and add an author...

TODO: add screenshots

... and then a book via the admin:

TODO: add screenshots

#### Meta Module - Controlling display and editable fields in the admin

Go to http://localhost:9000/admin/books.
Now comment the following line from [main.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/main.go)

    book.IndexAttrs("Title", "AuthorNames", "ReleaseDate", "DisplayPrice")

and reload the books admin page.

TODO: add screenshot

You will see a much more crowded list: We had 4 attributes `"Title", "AuthorNames", "ReleaseDate", "DisplayPrice"`. `Title` and `ReleaseDate` are both defined on our `Book` model, but the other two are not. For the Authors field you only see a list of References to the `Author` objects - something like `[0xc208161bc0]`. We want the list of Authors devided by `,` instead. You can achieve that by defining a `Meta` field:

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



### Frontend

QOR does not provide any builtin templating or routing support - use whatever library is best fit for your needs. In this tutorial we will use [gin](https://github.com/gin-gonic/gin).

#### List of Books



#### MediaLibrary - Adding product images

qor/media_library

`Base` is the low level object to deal with images offering cropping, resizing, and URL contruction for images.

INERT INTO users (name) VALUES ("admin");


#### Add Locales && translations


-- later


### Add customers (model)
### Add orders