![website](https://github.com/alpiepho/pup-learning/workflows/website/badge.svg)


**UPDATE** Try the golang version described at the end.
# pup-learning

Deployed on GitHub pages [here](https://alpiepho.github.io/pup-learning/).


A tool to gather Learning classes completed along with details.

When run as a Node.js tool, it will parse all the Completed Courses of the configured
user, It will save that data as ./artifact/sample/json, and it will generate

- public/index.html
- artifacts/learning-summary.mdx

The associated GH Action will deploy the index.html file.  

The learning-summary.mdx can be manually copied to the my-blog2 directory and committed 
to that repo as post in the blog.


## Mac Install

Install npm from the website or homebrew.  You will also need yarn.

NOTE: Normally, you can use either npm or yarn to install modules.  The debug of
Chrome with DNS Manager (mising a valid certificate), discovered that we need older
versions of some packages.  This can be done with either npm or yarn.  We currently only
have this setup with the yarn.lock file.

## Linux Install

For a Unbuntu system, use:

<pre>
sudo apt-get install -y npm yarn
</pre>

For other distributions, please search on Google.

## Install then run

First, there is a one-time setup of enviroment variables like:

```
export PUP_USERNAME=<LinkedIn User Name>
export PUP_PASSWORD=<LinkedIn User Password>
yarn install
```

From a command line:

```
yarn start
```

This should take a few minutes to navigate thru all the tabs. 

You can also runs this 'headless' or without a browser window.  Look for 'headless' in site.js.  


## Internal Settings

There are number of things can be changed quickly in the source code.  Eventually they
will be added as program options.

Look for "INTERNAL OPTION".

```
    browserType:     "firefox", // "chrome, firefox"
    headless:        false,     // run without windows
    scrollToBottom:   true,     // scroll page to bottom (WARNING: non-visible thumbnails are not loaded until page is scrolled)
    gatherDetails:    true,     // parse the details
```


### Chromium vs Firefox for Puppeteer

Puppeteer can run automated tests with both Chrome (technically the Chromium build) and
Firefox.

### Headless

Headless is the ability to run a web page without showing on the screen.  This mode can
be used for automated testing.  This works with both Chrome and Firefox.


### Local Test of index.html

- cd public
- python -m SimpleHTTPServer
- open http://localhost:8000/

## Know Issues

- options must be set in code
- GH Actions almost works, but LinkedIn flags login from an unknown IP address (see saved for .yml file)


## TODO List:

- pull course example files

## Future changes:

Some ideas for future changes:

- Create summary page from data


## Golang version

As a side project to learn Golang, this was rewritten in Go and chromedp.  Go is a little
more difficult than javascript or python, but it seems to run a little faster.  Also,
chromedp will allow automating the login to Linkedin Learning and can run headless.

Install golang [here](https://golang.org/doc/install)

[chromedp](https://github.com/chromedp/chromedp) library.

`go get -u github.com/chromedp/chromedp`

`go run main.go`


