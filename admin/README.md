# Qor Admin

## Introduction

## Installation

## Setup

### Database

  Only accept https://github.com/jinzhu/gorm. "mysql", "sqlite", "postgresql" are supported by gorm

  Initialize db by

      DB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPwd, dbName))

### Roles

  // placeholder for role README

### l10n

  // placeholder for l10n README

### Components

  // placeholder for publish README

### Initialize Admin

      config := qor.Config{DB: &DB}
      Admin := admin.New(&config)

## Config

### Authentication

  Auth must implement 3 functions

  1. Login(*Context)
  2. Logout(*Context)
  3. GetCurrentUser(*Context) qor.CurrentUser

  Then use

    Admin.SetAuth(&Auth{})

### Menu control

  Menu accept 4 parameters.

  1. Name, the text displayed in menu link
  2. params, parameters attached to menu link
  3. Link, url the menu link linked to
  4. Items, type is `[]*Menu`, used for build sub menus

  Register menu like

    Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})

### Add resource

### Meta config

#### text field

#### select one

#### select many

#### rich editor

#### media upload

### Filter
### Search
### Route
### Server
