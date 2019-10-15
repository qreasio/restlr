# Restlr 
Open Source WP Rest API Compatible Golang CMS API.

### Author
Isak Rickyanto

## Status
Alpha (WIP) - Still in early development

Below are working API Paths:
- Posts
- Pages

## Overview
Restlr is experimental Golang based API based CMS that is fully compatible with WP Rest API and can connect directly to existing Wordpress database.

## Why does this project exist?
The purpose of this project is to provide faster, more secure and memory efficient alternative of official Wordpress Rest API and 
will provide CLI that query wordpress posts and pages to generate json file that is suitable for Jamstack that pull content from Wordpress database, 
then generate static site with SSG like Hugo.

## Restlr vs Wordpress Rest API Performance Comparison
Restlr has much better performance difference (8X-9X faster) compare to WP Rest API especially if query posts with parameter _embed like: http://host/wp-json/wp/v2/posts?_embed

- Restlr: 264-500ms
- Wordpress REST API: 1.7secs - 4secs

Will update more about benchmark here

### Libraries:

- Micro services framework using GoKit 
- Router using Chi
- Logging using Logrus
- PHP Session decoder using github.com/yvasiyarov/php_session_decoder
- Env Var .env file using github.com/joho/godotenv

### Required Environment Variables

Database: 
- DATABASE_URL=mysql://root:pass@localhost/dbname?parseTime=true&sql_mode=ansi

API:
- SERVER_PORT=8080
- API_HOST=https://api.example.com
- SITE_URL=https://www.example.com
- UPLOAD_PATH=uploads
- TABLE_PREFIX=wp_
- API_PATH=/wp-json/wp
- VERSION=v2

### How to Run
1. Copy sample.env as .env
2. Run:

    > go get
           
    > go run main.go
3. Access the API to http://localhost:8080/wp-json/wp/v2/posts if using default port 8080