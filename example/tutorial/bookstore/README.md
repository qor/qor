# QOR example application

This is an example application to show and explain features of QOR.

You need basic understanding of Go to understand the documentation of this app.

Run the code from the QOR repository in

    cd $GOPATH/src/github.com/qor/qor/example/tutorial/01

Once you have gone through this doc you could copy the directory structure of the app to use it as a template or start your own application from scratch.


## Get Started

We want to create a simple bookstore application. We will start by building a catalog of books and then later add a storefront. We will add a staging environment (database rather) so that editors can make changes to contents and then later publish them to a live database.

We will add Localization(L10n) for the books and authors and Internationalization(I18n) support for the complete backoffice we built with `qor/admin`.


## Prerequisites

* GoLang 1.x+ (at the time of writing I am using >=1.4.0 versions)
* Install QOR:

    go get github.com/qor/qor

* A database - for example PostgreSQL or MySQL
* Install dependencies: cd into the QOR source directory and run

    cd $GOPATH/src/github.com/qor/qor
    go get -u ./...

* Install [Gin](https://github.com/gin-gonic/gin) - QOR does not require gin, but we use it in the example application for routing and templating:

    go get github.com/gin-gonic/gin

* [Optional: fresh](https://github.com/pilu/fresh) being installed:

    go get github.com/pilu/fresh

`fresh` is not necessary to use QOR, but it will make your life easier when playing with the tutorial: It monitors for file changes and automatically recompiles your code every time something has changed.

If you don't want to go with fresh you will have to terminate, rebuild, and rerun your code every time instead.



### Create a database and a db user for the example application

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
    $ mysql -uroot -p
    mysql> DROP DATABASE IF EXISTS qor_bookstore;
    mysql> CREATE DATABASE qor_bookstore DEFAULT CHARACTER SET utf8mb4;
    mysql> CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor';
    mysql> GRANT ALL ON qor_bookstore.* TO 'qor'@'localhost';
    mysql> FLUSH PRIVILEGES;

TODO: it's a bug that this is needed - but for now this needs to be called manually:
    mysql> use qor_bookstore;
    mysql> CREATE TABLE `translations` (`key` varchar(255),`locale` varchar(255),`value` varchar(255) , PRIMARY KEY (`key`(100),`locale`(100)));


### The basic models

We will need the following two models to start with:

* Author
* Book

The `Author` model is very simple:

    type Author struct {
    	gorm.Model
    	publish.Status
    	l10n.Locale

    	Name string
    }

All QOR models "inherit" from `gorm.model`. (see https://github.com/jinzhu/gorm).
Our author model for now only has a `Name`.
Ignore `publish.Status` and `l10n.Locale` for now - we will address these in later parts of the tutorial.

The Book model has a few more fields:

    type Book struct {
    	gorm.Model
    	publish.Status
    	l10n.Locale

    	Title       string
    	Synopsis    string
    	ReleaseDate time.Time
    	Authors     []*Author `gorm:"many2many:book_authors"`
    	Price       float64
    	CoverImage  media_library.FileSystem
    }


The only interesting part here is the gorm struct tag: `gorm:many2many:book_authors"`; It tells `gorm` to create a join table `book_authors`.

Ignore `publish.Status`, `l10n.Locale` and `media_library.FileSystem` for now - we will address these in later parts of the tutorial.

That's almost it: If you [look at models.go](https://github.com/qor/qor/tree/master/example/tutorial/bookstore/01/app/models/models.go) you can see an `init()` function at the end: It sets up a db connection and `db.AutoMigrate(&Author{}, &Book{}, &User{})` tells QOR to automatically create the tables for our models.

You can ignore the User model for now - we will look at that part later.

Let's start the tutorial app once to see what happens when models get auto-migrated.

    go/src/github.com/qor/qor/example/tutorial/bookstore/01 [01 (master)] $ fresh

or if you don't want to use fresh you can build and run the app:

    /go/src/github.com/qor/qor/example/tutorial/bookstore/01 [01 (master)] $ go build -o tutorial main.go
    /go/src/github.com/qor/qor/example/tutorial/bookstore/01 [01 (master)] $ ./tutorial

If you now check your db you would see something like this:

#### PostgreSQL

    qor_bookstore=# \d
    List of relations
    Schema |         Name         |   Type   | Owner
    --------+----------------------+----------+-------
    public | authors              | table    | qor
    public | authors_draft        | table    | qor
    public | authors_draft_id_seq | sequence | qor
    public | authors_id_seq       | sequence | qor
    public | book_authors         | table    | qor
    public | books                | table    | qor
    public | books_draft          | table    | qor
    public | books_draft_id_seq   | sequence | qor
    public | books_id_seq         | sequence | qor
    public | translations         | table    | qor
    public | users                | table    | qor
    public | users_id_seq         | sequence | qor
    (12 rows)

    qor_bookstore=# \d authors
    Table "public.authors"
    Column     |           Type           |                      Modifiers
    ----------------+--------------------------+------------------------------------------------------
    id             | integer                  | not null default nextval('authors_id_seq'::regclass)
    created_at     | timestamp with time zone |
    updated_at     | timestamp with time zone |
    deleted_at     | timestamp with time zone |
    publish_status | boolean                  |
    language_code  | character varying(6)     | not null
    name           | character varying(255)   |
    Indexes:
    "authors_pkey" PRIMARY KEY, btree (id, language_code)


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

#### Insert some users

In the next step we want to log into the admin so we need some users. Run this on your database:

    INSERT INTO users (name,role) VALUES ('admin','admin');
    INSERT INTO users (name,role) VALUES ('user1','user');

You must have run the application *once* before this step, otherwise the users table would not yet exist.

### Directory structure

Before we look at the actual admin here is a brief breakdown of the directory structure of the example app:

    .
    ├── app
    │   ├── handlers
    │   │   └── handlers.go
    │   ├── models
    │   │   └── models.go
    │   └── resources
    │       └── resources.go
    ├── main.go
    ├── public -> ../public
    │   ├── assets
    │   │   └── css
    │   │       └── bookstore.css
    │   ├── system
    └── templates -> ../templates
        ├── book.tmpl
        └── list.tmpl

* The controllers are in `app/handlers`
* models and db initialization happen in `app/models`
* Resources are an itegral part of QOR/admin. Whenever resources or `Meta...`  is mentioned in this doc you will find the code it's referring to in `app/resources`
* main.go starts the webserver and additionally contains the routes right now. In a bigger project you would put them probably somewhere like `app/config/routes.go`
* Static files are served from `public`. `public/system` is where the `qor/medialibrary` puts files related to your resources - eg. ad uploaded image

NB: The symlinks (public, templates) are here so that we can reuse them in later parts of a tutorial.

### Admin

If the bookstore app is not yet running, start it by running `fresh` in the bookstore directory:

    go/src/github.com/qor/qor/example/tutorial/bookstore/01 [bookstore (master)] $ fresh

Go to [http://localhost:9000/admin](http://localhost:9000/admin) and log in as `admin`. You should see the main admin interface:

![qor_admin](https://raw.githubusercontent.com/qor/qor/docs_and_tutorial/example/tutorial/bookstore/screenshots/qor_admin1.png)

The menu at the top gets created by adding your models as resources to the admin in [main.go](https://github.com/qor/qor/blob/docs_and_tutorial/example/tutorial/bookstore/01/app/resources/resources.go):

	Admin := admin.New(&qor.Config{DB: &db})

	Admin.AddResource(
		&User{},
		&admin.Config{
			Menu: []string{"User Management"},
			Name: "Users",
		},
	)

You can see how the rest of the resources was added in [resources.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/resources/resources.go), the `db` object referenced here is set up in [models.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/models/models.go)

Go ahead and go to the authors admin and add an author...

![qor_admin_add_author](https://raw.githubusercontent.com/qor/qor/docs_and_tutorial/example/tutorial/bookstore/screenshots/qor_admin_add_author.png)

... and then a book via the admin:

![qor_admin_add_book](https://raw.githubusercontent.com/qor/qor/docs_and_tutorial/example/tutorial/bookstore/screenshots/qor_admin_add_book.png)

![qor_admin_books1](https://raw.githubusercontent.com/qor/qor/docs_and_tutorial/example/tutorial/bookstore/screenshots/qor_admin_books1.png)


#### Meta Module - Controlling display and editable fields in the admin

Go to [http://localhost:9000/admin/books](http://localhost:9000/admin/books).
Now comment the following line from [resources.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/resources/resources.go)

	book.IndexAttrs("ID", "Title", "AuthorNames", "FormattedDate", "DisplayPrice")

and reload the books admin page.

![qor_admin_books2](https://raw.githubusercontent.com/qor/qor/docs_and_tutorial/example/tutorial/bookstore/screenshots/qor_admin_books2.png)

You will see a much more crowded list: We had 5 attributes `"ID", "Title", "AuthorNames", "FormattedDate", and "DisplayPrice"`. `Id`, `Title`, and `ReleaseDate` are defined on our `Book` model, but the others are not. For the Authors field you only see a list of References to the `Author` objects - something like `[0xc208161bc0]`. We want the list of author names devided by `,` instead. You can achieve that by defining a `Meta` field:

	book.Meta(&admin.Meta{
		Name:  "AuthorNames",
		Label: "Authors",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			// log.Println("LOCALE:", context.MustGet("locale"))
			// log.Println("ctxt", context)
			if value == nil {
				return value
			}
			book := value.(*Book)
			if err := context.GetDB().Model(&book).Related(&book.Authors, "Authors").Error; err != nil {
				panic(err)
			}

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

We define a `Meta` field for the `book` `Resource`. It's internal name is "AuthorNames", which we use in `book.IndexAttrs()` to use it in our admin book listing. The `Label` is what goes into the table header and the `Valuer` is a function that will return the display value we want - in our case the comma separated list of author names.


##### Editable and Create fields

By default all defined model attributes and `Meta` attributes are included in the edit interface. If you need to limit the fields that are editable you can manually set the `EditAttrs` (for editing existing objects) and `NewAttrs` (for object creation):

	book.NewAttrs("Title", "Authors", "Synopsis", "ReleaseDate", "Price", "CoverImage")
	book.EditAttrs("Title", "Authors", "Synopsis", "ReleaseDate", "Price", "CoverImage")


#### Searchable Fields

To get a searchfield on the list display of your resource you simply add a line like this:

    book.SearchAttrs("ID", "Title")

Which will add a search(field) for resources matching on the defined fields.


#### Meta Field Types

QOR will pick an input type based on your struct types - but sometimes you want to change the default. For example we might want to have a text area with some editing functions instead of just an `<input type="text">`:

	book.Meta(&admin.Meta{
		Name: "Synopsis",
		Type: "rich_editor",
	})

TODO: other types - at least select_one and select_many. add list.



### Publish - Edit first then push to production DB

    import "github.com/qor/qor/publish"

The Publish module allows you edit contents of your site without having them go online right away. Every model that you want to be able to `Publish` needs to inherit `publish.Status`:

    type Author struct {
    	gorm.Model
    	publish.Status
    	l10n.Locale

    	Name string
    }

Then initialize `Publish` and set up AutoMigrate (see [init() in models.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/models/models.go#L106)):

	Pub = publish.New(&Db)
	Pub.AutoMigrate(&Author{}, &Book{})

	StagingDB = Pub.DraftDB()         // Draft resources are saved here
	ProductionDB = Pub.ProductionDB() // Published resources are saved here

And change the `DB` config of `admin`:

	// Admin := admin.New(&qor.Config{DB: &db})         // this is without publish
	Admin = admin.New(&qor.Config{DB: Pub.DraftDB()})   // with publish
	Admin.AddResource(Pub)

You have now a *"Publish"* menu: Changes you make on *publishable* objects are not going online right away. Add an author and/or a book and check out the [Publish section](http://localhost:9000/admin/publish):

![qor_publish](https://raw.githubusercontent.com/qor/qor/docs_and_tutorial/example/tutorial/bookstore/screenshots/qor_publish.png)

You can select the changes you want to publish (check out the "View Diff" link to see changes) and then either publish them to the live DB or discard them.

You can check that before bublishing the first time your `authors` table should be empty, while `authors_draft` contains the contents you see in QOR admin. After publishing these contents get copied to the live `authors` table.



#### MediaLibrary - Adding product images

    import "github.com/qor/qor/media_library"

We will only briefly touch on the `qor/media_library`. It provides support for uploading, storage, and resizing of images. Define an attribute with the `media_library.FileSystem` type:

    type Book struct {
    	gorm.Model
    	publish.Status
    	l10n.Locale

        [...]
    	CoverImage  media_library.FileSystem
    }

and you're almost done. You need a route to serve the files from:

	router.StaticFS("/system/", http.Dir("public/system")) # this is in main.go

Support for publish (draft version, publish to live) is built in. This is what the directory structure for the `Book` `CoverImage`s looks like:

    /public [public (docs_and_tutorial)] $ tree
    .
    └── system
        ├── books
        │   └── 1
        │       └── CoverImage
        │           ├── P1210896.20150604163815084702067.jpg
        │           └── P1210896.20150604163815084702067.original.jpg
        └── books_draft
            └── 1
                └── CoverImage
                    ├── P1210896.20150604163815084702067.jpg
                    └── P1210896.20150604163815084702067.original.jpg

In your templates you can use the image like this:

    <img src="{{.book.CoverImage}}" />

Edit the book you previously created and click on the image you uploaded there. The crop interface will pop up:

![qor_media](https://raw.githubusercontent.com/qor/qor/docs_and_tutorial/example/tutorial/bookstore/screenshots/qor_media.png)



### L10n - Localizing your resources

To localize your resources, for example having an english and a japanese "version" of an author or a book you need to use the `l10n` module.

    import "github.com/qor/qor/l10n"

Any model you want to have localization support on needs to inherit from l10n.Locale:

    type Author struct {
    	gorm.Model
    	publish.Status
    	l10n.Locale

    	Name string
    }

Set your default locale:

    func init() {
        l10n.Global = "en"
	    l10n.RegisterCallbacks(&Db)
    }


TODO @jinzhu: what exactly does l10n.RegisterCallbacks

You're almost done



### I18n - Translating strings

    import "github.com/qor/qor/i18n"

To add I18N support for the `qor/admin`

    var (
    	Admin *admin.Admin
    	I18n  *i18n.I18n       // this needs to be exported
    )

    func init() {
    	// setting up QOR admin
    	Admin = admin.New(&qor.Config{DB: Pub.DraftDB()})
    	Admin.AddResource(Pub)
    	Admin.SetAuth(&Auth{})

    	I18n := i18n.New(database.New(StagingDB))
    	Admin.AddResource(I18n)

TODO: screenshots

TODO: Add the funcMap with the `T` template functions and import frontend strings into QOR too. (@bom-d-van)



### Frontend

QOR does not provide any builtin templating or routing support - use whatever library is best fit for your needs. In this tutorial we will use [gin](https://github.com/gin-gonic/gin) and the stl `html/template`s.

http://localhost:9000/books

TODO: switching the language/locale on the frontend
