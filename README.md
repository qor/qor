# QOR

English Chat Room: [![Join the chat at https://gitter.im/qor/qor](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/qor/qor?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

中文聊天室： [![加入中国Qor聊天室 https://gitter.im/qor/qor/china](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/qor/qor/china)

[![Build Status](https://semaphoreci.com/api/v1/theplant/qor/branches/action/badge.svg)](https://semaphoreci.com/theplant/qor)

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


## Example Application

[The example application](https://github.com/qor/qor-example) is a work in progress but already covers the following modules:

* [x] Setup & Installation
* [x] Define a first set of resources (qor/admin)
* [x] Introduce Meta - Back Office display control for your resources
* [x] Basic Media library usage
* [x] Using Publish
* [x] L10n & I18n
* [\] Roles (very little)
* [ ] Worker


## Frontend Development

Requires [Node.js](https://nodejs.org/) and [Gulp](http://gulpjs.com/) for building frontend files

```bash
npm install && npm install -g gulp
```

- Watch SCSS/JavaScript changes: `gulp`
- Build Release files: `gulp release`


## Q&A

1. How to integrate with [beego](https://github.com/astaxie/beego)

```
	adm := admin.New(&qor.Config{DB: &db.DB})
	adm.AddResource(&db.User{}, &admin.Config{Menu: []string{"管理"}})

	mux := http.NewServeMux()
	adm.MountTo("/admin", mux)

	beego.Handler("/admin/*", mux)
	beego.Run()
```

2. How to integrate with [Gin](https://github.com/gin-gonic/gin)

```
	mux := http.NewServeMux()
	admin.Admin.MountTo("/admin", mux)

	r := gin.Default()
	r.Any("/admin/*w", gin.WrapH(mux))
	r.Run()
```
