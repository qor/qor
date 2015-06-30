# QOR

[![Build Status](https://semaphoreci.com/api/v1/projects/3a3db8d6-c6ac-46b8-9b34-453aabdced22/430434/badge.svg)](https://semaphoreci.com/theplant/qor)

## What is QOR?

QOR is a set of libraries written in Go that abstracts common features needed for business applications, CMSs, and E-commerce systems.

This is actually the third version of QOR: 1 and 2 were written in ruby and used internally at [The Plant](https://theplant.jp).
We decided to rewrite QOR in Go and open source it - which has happened in June 2015. 

QOR is still beta software - we will probably break an API or two before we will release a stable 1.0 - scheduled for September 2015.

While nearing API freeze our other main focus is building up API documentation for each module and a tutorial that eventually cover features from most of the modules.

### What QOR is not

## The modules

* Admin - The heart of any QOR system, where you manage all you resources

* Publish - Providing a staging environment for all content changes to be reviewed before being published to the live system

* Transition - A configurable State Machine: define states, events (eg. pay order), and validation constraints for state transitions

* Media Library - Asset Management with support for several cloud storage backends and publishing via a CDN

* Worker (Batch processing) - A process scheduler

* Exchange - Data exchange with other business applications using CSV or Excel data

* Internationalization (i18n) - Managing and (inline) editing translations

* Localization (l10n) - Managing locales in multilingual environments

* Roles - Access Control

* Layout

## API Documentation

[![GoDoc](https://godoc.org/github.com/qor/qor?status.svg)](https://godoc.org/github.com/qor/qor)




## Tutorial

[The tutorial](https://github.com/qor/qor/tree/master/example/tutorial/bookstore) is currently a work in progress.

- [x] Setup & Installation
- [x] Define a first set of resources
- [x] Introduce Meta - Back Office display control for your resources
- [ ] Using Publish
- [ ] L10n & I18n
- [ ] Roles
