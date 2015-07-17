# QOR example application

This is an example application to show and explain features of QOR.

You need basic understanding of Go to understand the documentation of this app.

Run the code from the QOR repository in

    cd $GOPATH/src/github.com/qor/qor/example/tutorial/01

Once you have gone through this doc you could copy the directory structure of the app to use it as a template or start your own application from scratch.


## Get Started

We want to create a simple bookstore application. We will start by building a catalog of books and then later add a storefront. We will add a staging environment (database rather) so that editors can make changes to contents and then later publish them to a live database.

We will add Localization (L10n) for the books and authors and Internationalization (I18n) support for the complete back-office we built with `qor/admin`.


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



### Create a database and a DB user for the example application

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


The only interesting part here is the Gorm struct tag: `gorm:many2many:book_authors"`; It tells `gorm` to create a join table `book_authors`.

Ignore `publish.Status`, `l10n.Locale` and `media_library.FileSystem` for now - we will address these in later parts of the tutorial.

That's almost it: If you [look at models.go](https://github.com/qor/qor/tree/master/example/tutorial/bookstore/01/app/models/models.go) you can see an `init()` function at the end: It sets up a db connection and `db.AutoMigrate(&Author{}, &Book{}, &User{})` tells QOR to automatically create the tables for our models.

You can ignore the User model for now - we will look at that part later.

Let's start the tutorial app once to see what happens when models get auto-migrated.

    go/src/github.com/qor/qor/example/tutorial/bookstore/01 [01 (master)] $ fresh

If you don't want to use fresh you can build and run the app:

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

As you can see QOR/Gorm added an `id` field as well as timestamp fields to keep track of creation, modification, and deletion times. We can ignore this for now - the main point is that you create your models without a unique identifier - QOR/Gorm will do this for you automatically. (TODO: @jinzhu please confirm)

NB: If you add new fields to your model they will get added to the database automatically with `DB.AutoMigrate` - deletions or *changes* of eg. the type will *not* be automatically migrated. (TODO: @jinzhu please confirm)

#### Insert some users

In the next step we want to log into the admin so we need some users. Run this on your database:

    INSERT INTO users (name,role) VALUES ('admin','admin');
    INSERT INTO users (name,role) VALUES ('user1','user');

You must have run the application *once* before this step, otherwise the users table would not yet exist.

### Directory structure

Before we look at the actual admin here is a brief breakdown of the directory structure of the example app:

    .
    ├── app
    │   ├── controllers
    │   │   └── controllers.go
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

* The controllers are in `app/controllers`
* models and db initialization happen in `app/models`
* Resources are an integral part of QOR/admin. Whenever resources or `Meta...`  is mentioned in this doc you will find the code it's referring to in `app/resources`
* main.go starts the webserver and additionally contains the routes right now. In a bigger project you would put them probably somewhere like `app/config/routes.go`
* Static files are served from `public`. `public/system` is where the `qor/medialibrary` puts files related to your resources - eg. an uploaded image

NB: The symlinks (public, templates) are here so that we can reuse them in later parts of a tutorial.

### Admin

If the bookstore app is not yet running, start it by running `fresh` in the bookstore directory:

    go/src/github.com/qor/qor/example/tutorial/bookstore/01 [bookstore (master)] $ fresh

Go to [http://localhost:9000/admin](http://localhost:9000/admin) and log in as `admin`. You should see the main admin interface:

![qor_admin](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_admin1.png)

The menu at the top gets created by adding your models as resources to the admin in [main.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/resources/resources.go):

	Admin := admin.New(&qor.Config{DB: &db})

	Admin.AddResource(
		&User{},
		&admin.Config{
			Menu: []string{"User Management"},
			Name: "Users",
		},
	)

You can see how the rest of the resources were added in [resources.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/resources/resources.go), the `db` object referenced here is set up in [models.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/models/models.go)

Go ahead and go to the authors admin and add an author...

![qor_admin_add_author](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_admin_add_author.png)

... and then a book via the admin:

![qor_admin_add_book](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_admin_add_book.png)

![qor_admin_books1](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_admin_books1.png)


#### Meta Module - Controlling display and editable fields in the admin

Go to [http://localhost:9000/admin/books](http://localhost:9000/admin/books).
Look for the following line from [resources.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/resources/resources.go)

	book.IndexAttrs("ID", "Title", "Authors", "FormattedDate", "Price")

and change `FormattedDate` to `ReleaseDate`

![qor_admin_books2](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_admin_books2.png)

We are getting the value of the `Book.ReleaseDate`. We want to show the date only so we define a `Meta` field:

	book.Meta(&admin.Meta{
		Name:  "FormattedDate",
		Label: "Release Date",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			book := value.(*Book)
			return book.ReleaseDate.Format("Jan 2, 2006")
		},
	})

We define a `Meta` field for the `book` `Resource`. It's internal name is "FormattedDate", which we use in `book.IndexAttrs()` to use it in our admin book listing. The `Label` is what goes into the table header and the `Valuer` is a function that will return the display value we want - in our case the formatted date that does not include the time.


##### Fields for Creation and Edit

By default all defined model attributes and `Meta` attributes are included in the edit interface. If you need to limit the fields that are editable you can call `EditAttrs` (for editing existing objects) and `NewAttrs` (for object creation):

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

TODO: other types - at least select_one and select_many.



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

You now have a *Publish* menu: Changes you make on *publishable* objects do not appear in the front end until they are *published*. Add an author and/or a book and check out the [Publish section](http://localhost:9000/admin/publish):

![qor_publish](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_publish.png)

You can select the changes you want to publish (check out the "View Diff" link to see changes) and then either publish them to the live DB or discard them.

You can check that before publishing the first time your `authors` table should be empty, while `authors_draft` contains the contents you see in the QOR admin. After publishing these contents get copied to the live `authors` table.



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

The `public/system` directories must already exist - they do not get created by QOR.
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

The directories for your resources (like books and books_draft) are created by the media_library. In your templates you can use the image like this:

    <img src="{{.book.CoverImage}}" />

Edit the book you previously created and click on the image you uploaded there. The crop interface will appear:

![qor_media](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_media.png)



### L10n - Localizing your resources

To localize your resources, for example to have an English and a Japanese "version" of an author or a book you need to use the `l10n` module.

    import "github.com/qor/qor/l10n"

Any model you want to have localization support on needs to inherit from l10n.Locale:

    type Author struct {
    	gorm.Model
    	publish.Status
    	l10n.Locale

    	Name string
    }

Set your default locale (In the example app these are called at the end of the `init()` function in [app/models/models.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/models/models.go)):

    func init() {
        l10n.Global = "en-US"
	    l10n.RegisterCallbacks(&Db)
    }

`l10n.Global = "en-US"` is the default used by the `l10n` package, but better to be explicit here.
`l10n.RegisterCallbacks` registers callbacks with `gorm` to keep track of changes in the different locales.

The last step is to define who can view or edit which locales. In `models.go` we have two methods defined on our `User` type:

    func (User) ViewableLocales() []string {
	return []string{l10n.Global, "ja-JP"}
    }

And to make the different locales viewable and

    func (user User) EditableLocales() []string {
    	if user.Role == "admin" {
    		return []string{l10n.Global, "ja-JP"}
    	} else {
    		return []string{}
    	}
    }

to make the different locales editable by the `admin` role.


### I18n - Translating strings

    import "github.com/qor/qor/i18n"

To add I18n support for `qor/admin` you need to register an `i18n.I18n` resource:

    var (
    	Admin *admin.Admin
    	I18n  *i18n.I18n       // this needs to be exported
    )

    func init() {
    	// setting up QOR admin
    	Admin = admin.New(&qor.Config{DB: Pub.DraftDB()})
        [...]

    	I18n := i18n.New(database.New(StagingDB))
    	Admin.AddResource(I18n)
        [...]

You can find the code in [app/resources/resources.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/resources/resources.go)

This will give you the `I18n` menu entry in admin.

Go ahead and look for the translation key `qor_admin.I18n` and translate it:

![qor_translate1](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_translate1.png)

Set the English translation to `Translations`, use the target language switcher in the table header and change the target language to `ja-JP` (Japanese) and translate it to `翻訳`. Now reload admin and you will see your translation in the menu on the left.

![qor_translate2](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_translate2.png)

If you go to eg. Authors and set the locale to `ja-JP` you will see your translation appear in the admin menu on the left:

![qor_translate3](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_translate3.png)

NB: Currently the example app is set up in a way that translation keys are added to the database the first time they are read from templates. In order to see all the keys you have to access each page where they are used once.

TODO: Add an example on how to import keys from eg. a YAML file.



### Front-end

QOR does not provide any built-in templating or routing support - you can use whatever library best fits your needs. In this example application tutorial we use [gin](https://github.com/gin-gonic/gin):

In [main.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/main.go)

	// frontend routes
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

we initialize a `gin.Router` and tell it where it can find our templates.

	// serve static files
	router.StaticFS("/system/", http.Dir("public/system"))
	router.StaticFS("/assets/", http.Dir("public/assets"))

add routes for static files

	// books
	bookRoutes := router.Group("/books")
	{
		// listing
		bookRoutes.GET("", controllers.ListBooksHandler)
		// single book - product page
		bookRoutes.GET("/:id", controllers.ViewBookHandler)
	}

and add two endpoints - one to list all our books and one book details page.

The controllers are defined in [app/controllers/controllers.go](https://github.com/qor/qor/blob/master/example/tutorial/bookstore/01/app/controllers/controllers.go):

    func ListBooksHandler(ctx *gin.Context) {
        # get the books data
        [...]

    	ctx.HTML(
    		http.StatusOK,
    		"list.tmpl",
    		gin.H{
    			"books": books,
    			"t": func(key string, args ...interface{}) template.HTML {
    				return template.HTML(resources.I18n.T(retrieveLocale(ctx), key, args...))
    			},
    		},
    	)
    }

    func ViewBookHandler(ctx *gin.Context) {
        [...]
    }

We get the data for all or one book and then pass it to the template in `ctx.HTML()`. One thing to note here is that we pass not only the data ( `"books": books`) but also a function `t` which is the translation function. It's used in the templates like this:

    <h1>{{call .t "frontend.books.List of Books"}}</h1>

`frontend.books.List of Books` will become the key that will appear in your translations resource. (after you accessed it once. See [I18n](https://github.com/qor/qor/tree/master/example/tutorial/bookstore#i18n---translating-strings) - the NB at the end of the section if you don't know why).

Go ahead and point your browser to:

[http://localhost:9000/books](http://localhost:9000/books)

If you have books in your system but see an empty page you have most likely not yet published your data. Go to [Publish section](http://127.0.0.1:9000/admin/publish), select all items and hit publish:

![qor_publish2](https://raw.githubusercontent.com/qor/qor/master/example/tutorial/bookstore/screenshots/qor_publish2.png)

[http://localhost:9000/books](http://localhost:9000/books) should now show your books.

TODO: switching the language/locale on the frontend


### Next steps

This example app will be extended to eventually showcase most QOR features and a tutorial that goes through building this app step by step is planned too.

Go ahead and copy the example application and start using your own resources. Have fun with QOR!


### Questions, suggestions, etc.

[@QORSKD](https://twitter.com/qorsdk)
