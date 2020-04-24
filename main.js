fs = require('fs');

base = require('./base');
site = require('./site');

const HTML_FILE = "./index.html";
const MD_FILE = "./learning.md";

const html1 = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width", initial-scale=1.0"/>
    <meta name="Description" content="LinkedIn Learning Courses Completed">
    <meta name="theme-color" content="#d36060"/>
    <title>
    LinkedIn Learning Courses Completed
    </title>
    <link rel="stylesheet" href="./css/style.css" />
    <link rel="manifest" href="./manifest.json" />
    <link rel="icon" 
      type="image/png" 
      href="./favicon.ico" />
  </head>
  <body class="body">
    <main>
    <article class="page">
      <h1>LinkedIn Learning Courses Completed</h1>
      <ul>
`;

const html2 = `
      </ul>
    </article>
  </body>
</html>
`;

const md1 = `
---
title: LinkedIn Completed Courses
date: "2020-04-24"
description: "Summary of my LinkedIn Learning Completed Courses"
---

(Warning: many images) This a summary of all the Linked-In courses I have completed. 
This is just the direct LinkedIn Learning courses.  There are a number of "Lynda.com"
courses that were taken before subscribing to the LinkedIn premium plan.


`;

const md2 = `
`;

const main = async () => {
  // INTERNAL OPTION
  options = { 
    browserType: "firefox", // "chrome, firefox"
    headless: false,
    useSampleData: false, 
    saveSampleData: true,
    screenshot: true, 
    screenshotDir: "/tmp/pup_learning_screenshots",
    scrollToBottom: true
  }
  const browser = await base.browser_init(options);
  if (!options.useSampleData) {
    options.version = await browser.version();
  }
  console.log("options:");
  console.log(options);

  // login, get list of completed courses, logout
  data = {}
  await site.process_login(browser, options);
  await site.process_completed(browser, options, data);
  await site.process_logout(browser, options);
  await base.browser_close(browser);

  //DEBUG
  // console.log("data:");
  // console.log(JSON.stringify(data, null, space=2));

  // generate artifacts from data - html
  let htmlStr = html1;
  data['completed-courses'].forEach(entry => {
    htmlStr += "            <li>\n";
    htmlStr += "              <ul>\n";
    htmlStr += "                <li>\n";
    htmlStr += "                  <a target=\"_blank\" href=\"" + entry['link'] + "\">\n";
    htmlStr += "                    " + entry['title'] + "\n";
    htmlStr += "                  </a>\n";
    htmlStr += "                </li>\n";
    htmlStr += "                <li>" + entry['author'] + "</li>\n";
    htmlStr += "                <li>" + entry['released-date'] + "</li>\n";
    htmlStr += "                <li>" + entry['duration'] + "</li>\n";
    htmlStr += "                <li>" + entry['completed-date'] + "</li>\n";
    htmlStr += "              </ul>\n";
    htmlStr += "            </li>\n";
  });
  htmlStr += html2;
  fs.writeFileSync(HTML_FILE, htmlStr);
   
  // TODO: generate markdown (.mdx) for blog
  let mdStr = md1;
  data['completed-courses'].forEach(entry => {
    mdStr += "[" + entry['title'] + "](" + entry['link'] + ")\n";
    mdStr += "- " + entry['author'] + "\n";
    mdStr += "- " + entry['released-date'] + "\n";
    mdStr += "- " + entry['duration'] + "\n";
    mdStr += "- " + entry['completed-date'] + "\n";
    mdStr += "\n";
  });
  mdStr += md2;
  fs.writeFileSync(MD_FILE, mdStr);



  // TODO: generate html for deploy on GH Pages

  console.log("done.");
};

main();

  
  