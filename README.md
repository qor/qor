# QOR

[![Join the chat at https://gitter.im/qor/qor](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/qor/qor?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

[![Build Status](https://semaphoreci.com/api/v1/projects/3a3db8d6-c6ac-46b8-9b34-453aabdced22/430434/badge.svg)](https://semaphoreci.com/theplant/qor)

## What is QOR?

QOR is a set of libraries written in Go that abstracts common features needed for business applications, CMSs, and E-commerce systems.

This is actually the third version of QOR: 1 and 2 were written in Ruby and used internally at [The Plant](https://theplant.jp).
We decided to rewrite QOR in Go and open source it - which has happened as of June 2015.

QOR is still beta software - we will probably break an API or two before we release a stable 1.0 (scheduled for September 2015).

While nearing API freeze our other main focus is building up API documentation for each module and a tutorial that will eventually cover features from most of the modules.

### What QOR is not

QOR is not a "boxed turnkey solution". You need proper coding skills to use it. It's designed to make the lives of developers easier when building complex EC systems, not providing you one out of the box.

## The modules

* Admin - The heart of any QOR system, where you manage all your resources

* Publish - Providing a staging environment for all content changes to be reviewed before being published to the live system

* Transition - A configurable State Machine: define states, events (eg. pay order), and validation constraints for state transitions

* Media Library - Asset Management with support for several cloud storage backends and publishing via a CDN

* Worker (Batch processing) - A process scheduler

* Exchange - Data exchange with other business applications using CSV or Excel data

* Internationalization (i18n) - Managing and (inline) editing of translations

* Localization (l10n) - Manage DB-backed models on per-locale basis, with support for defining/editing localizable attributes, and locale-based querying

* Roles - Access Control


## API Documentation

We are planning to update the godoc documentation for all modules after the API for the 1.0 release is frozen. Still outstanding are:

* [ ] Admin
* [ ] Publish
* [ ] Transition
* [ ] Media Library
* [ ] Worker
* [ ] Exchange
* [ ] Internationalization (i18n)
* [ ] Localization (l10n)
* [ ] Roles


## Tutorial

[The tutorial](https://github.com/qor/qor/tree/docs_and_tutorial/example/tutorial/bookstore) is currently a work in progress.

* [x] Setup & Installation
* [x] Define a first set of resources
* [x] Introduce Meta - Back Office display control for your resources
* [x] Using Publish
* [ ] L10n & I18n
* [ ] Roles
* [ ] Worker


## Front End Build

### Main

Main asset directories of Admin and other modules:

```
/
├── admin/views/assets/
│   ├── fonts
│   ├── images
│   ├── javascripts
│   └── stylesheets
│
├── i18n/views/themes/i18n/assets/
│   ├── javascripts
│   └── stylesheets
│
├── l10n/views/themes/l10n/assets/
│   ├── javascripts
│   └── stylesheets
│
└── publish/views/themes/publish/assets/
    ├── javascripts
    └── stylesheets
```


### Build

> Requires [Node.js](https://nodejs.org/) (with [NPM](https://www.npmjs.com/) built-in) development environment.


#### Install [Gulp](http://gulpjs.com/)

```bash
npm install -g gulp
```

#### Install dependencies

```bash
npm install
```

#### Run Admin tasks

- Watch: `gulp`
- Build JS: `gulp js`
- Build CSS: `gulp css`
- Compile SCSS: `gulp sass`
- Release: `gulp release`


#### Run module tasks

Take I18n module for example:

- Watch: `gulp --i18n`
- Build JS: `gulp js --i18n`
- Build CSS: `gulp css --i18n`
- Compile SCSS: `gulp sass --i18n`
- Release: `gulp release --i18n`
