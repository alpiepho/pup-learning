fs = require('fs');

base = require('./base');
site = require('./site');

const HTML_FILE = "./public/index.html";
const MDX_FILE = "./artifacts/learning-summary.mdx";
const MD_FILE = "./artifacts/learning-summary.md";
const SCREENSHOT_DIR = "./screenshots";

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
    <link rel="stylesheet" href="./style.css" />
    <link rel="manifest" href="./manifest.json" />
    <link rel="icon"
      type="image/png" 
      href="./favicon.ico" />
  </head>
  <body class="body">
    <main>
    <article class="page">
      <h1  id=\"top\">LinkedIn Learning Courses Completed</h1>

      <div class="introduction">
      <p>
      This a summary of all the Linked-In courses I have completed. 
      This is just the direct LinkedIn Learning courses.  There are a number of "Lynda.com"
      courses that were taken before subscribing to the LinkedIn premium plan.
      </p>
      <p>
      This list is generated from a tool called "pup-learning" that can be found
      <a
        href="https://github.com/alpiepho/pup-learning"
        target="_blank"
        rel="noreferrer"
      >here</a>.  This tool needs to be run manually to parse the LinkedIn Learning
      site to gather the list of courses I have taken.
      </p>
      <p>
        If you look over the list of courses, there is variety.  I fully admit that
        my attention for some courses was less that other.  My form of bing watching :) 
      </p>
      </div>
`;

const html2 = `
    <div id=\"bottom\"></div>
    </article>
  </body>
</html>
`;

const md1 = `---
title: LinkedIn Completed Courses
date: "2020-06-03"
description: "Summary of my LinkedIn Learning Completed Courses"
---

(Warning: many images) This a summary of all the Linked-In courses I have completed. 
This is just the direct LinkedIn Learning courses.  There are a number of "Lynda.com"
courses that were taken before subscribing to the LinkedIn premium plan.

A full summary with more details can be found [here](https://alpiepho.github.io/pup-learning/).

#### top

`;

const md2 = `

#### bottom
`;

function build_hours_minutes(data) {
  // Derive timestamps and duration, sort
  let totalSec = 0;
  data['completed-courses'].forEach(entry => {
    // assume "An Bm" or "Bm"
    let parts = entry['duration'].split(' ');
    for (i=0; i<parts.length; i++) {
      if (parts[i].includes('h')) {
        val = parseInt(parts[i].replace('h', ''));
        totalSec += val*60*60; 
      }
      if (parts[i].includes('m')) {
        val = parseInt(parts[i].replace('m', ''));
        totalSec += val*60; 
      }
      if (parts[i].includes('s')) {
        val = parseInt(parts[i].replace('s', ''));
        totalSec += val; 
      }
    }
    entry['released-ts'] = Date.parse(entry['released-date']);
    entry['completed-ts'] = Date.parse(entry['completed-date']);
  });

  let totalMin = Math.floor(totalSec / 60);
  totalH = Math.floor(totalMin / 60); 
  totalM = totalMin - (totalH*60);
  return [totalH, totalM];
}


function build_html(data, totalH, totalM) {
  // generate artifacts from data - html
  let htmlStr = html1;

  today = new Date()
  htmlStr += "<sup><sub>(updated " + today + ")</sub></sup>\n\n"

  htmlStr += "      <br/><p>Totals - Course: " + data['completed-courses'].length + ", Time: " + totalH + "h " + totalM + "m</p><br/>\n\n";
  htmlStr += "      <ul class=\"mainul\">\n";
  data['completed-courses'].forEach(entry => {
    htmlStr += "            <li>\n";
    htmlStr += "              <ul>\n";
    htmlStr += "                <li>\n";

    htmlStr += "                  <div class=\"leftside\">\n";
    if (entry['img_file'])
      htmlStr += "                    <p><img src=\"" + entry['img_file'] + "\" loading=\"lazy\"</img></p>\n";
    else
      htmlStr += "                    <p><img src=\"" + entry['img'] + "\" loading=\"lazy\"</img></p>\n";

    htmlStr += "                  </div>\n";

    htmlStr += "                  <div class=\"rightside\">\n";
    htmlStr += "                    <ul>\n";
    htmlStr += "                      <li>\n";
    htmlStr += "                        <a target=\"_blank\" href=\"" + entry['link'] + "\">\n";
    htmlStr += "                        " + entry['title'] + "\n";
    htmlStr += "                        </a>  ";
    htmlStr += "                      </li>\n";
    htmlStr += "                      <li>\n";
    htmlStr += "                        <span>(" + entry['released-date'].replace('Updated ','') + " ... " + entry['duration'] + ")</span>\n";
    htmlStr += "                      </li>\n";
    htmlStr += "                      <li>\n";
    if (entry['linkedin']) {
      htmlStr += "                        <li><a target=\"_blank\" href=\"" + entry['linkedin'] + "\">" + entry['author'] + "</a></li>\n";
    } else {
      htmlStr += "                        <li>Author: " + entry['author'] + "</li>\n";
    }
    htmlStr += "                      </li>\n";
    htmlStr += "                      <li>\n";
    htmlStr += "                        <li class=\"completed\"><i>Completed: " + entry['completed-date'] + "</i></li>\n";
    htmlStr += "                      </li>\n";
    htmlStr += "                    </ul>\n";
    htmlStr += "                  </div>\n";
    htmlStr += "                </li>\n";
    htmlStr += "                <li class=\"details\">" + entry['details'] + "</li>\n";
    htmlStr += "                <li class=\"topbottom\"><a href=\"#top\">top</a> / <a href=\"#bottom\">bottom</a></li>\n";



    // if (entry['img_file'])
    //   htmlStr += "                  <p><img src=\"" + entry['img_file'] + "\" loading=\"lazy\"</img></p>\n";
    // else
    //   htmlStr += "                  <p><img src=\"" + entry['img'] + "\" loading=\"lazy\"</img></p>\n";

    // htmlStr += "                  <a target=\"_blank\" href=\"" + entry['link'] + "\">\n";
    // htmlStr += "                    " + entry['title'] + "\n";
    // htmlStr += "                  </a>  ";
    // htmlStr += "                <span>(" + entry['released-date'].replace('Updated ','') + " ... " + entry['duration'] + ")</span>\n";
    // htmlStr += "                </li>\n";
    // if (entry['linkedin']) {
    //   htmlStr += "                <li><a target=\"_blank\" href=\"" + entry['linkedin'] + "\">" + entry['author'] + "</a></li>\n";
    // } else {
    //   htmlStr += "                <li>Author: " + entry['author'] + "</li>\n";
    // }
    // htmlStr += "                <li class=\"completed\"><i>Completed: " + entry['completed-date'] + "</i></li>\n";
    // htmlStr += "                <li class=\"details\">" + entry['details'] + "</li>\n";
    // htmlStr += "                <li class=\"topbottom\"><a href=\"#top\">top</a> / <a href=\"#bottom\">bottom</a></li>\n";
    htmlStr += "              </ul>\n";
    htmlStr += "            </li>\n";
  });
  htmlStr += "      </ul>";
  htmlStr += html2;
  fs.writeFileSync(HTML_FILE, htmlStr);
}

// function build_md(data, totalH, totalM) {
//   // generate markdown (.mdx) for blog
//   let mdStr = md1;
//   mdStr += "Total Completed Courses: " + data['completed-courses'].length + ", Time: " + totalH + "h " + totalM + "m\n";
//   mdStr += "<br/>\n";
//   mdStr += "<br/>\n";
//   mdStr += "<br/>\n";
//   mdStr += "\n";
//   data['completed-courses'].forEach(entry => {
//     mdStr += "\n";
//     if (entry['img']) {
//       mdStr += "![thumbnail](" + entry['img'] + ")\n";
//     }
//     mdStr += "\n";
//     mdStr += "[" + entry['title'] + "](" + entry['link'] + ")\n";
//     mdStr += "- Author: " + entry['author'] + "\n";
//     //mdStr += "- Released: " + entry['released-date'] + "\n";
//     mdStr += "- Duration: " + entry['duration'] + "\n";
//     mdStr += "- Completed: " + entry['completed-date'] + "\n";
//     mdStr += "- Details: " + entry['details'] + "\n";
//     mdStr += "- [top](#top) / [bottom](#bottom)\n";

//     mdStr += "<br/>\n";
//     mdStr += "<br/>\n";
//     mdStr += "<br/>\n";
//       mdStr += "\n";
//   });
//   mdStr += md2;
//   fs.writeFileSync(MDX_FILE, mdStr);
//   fs.writeFileSync(MD_FILE, mdStr);
// }

const main = async () => {
  // INTERNAL OPTIONS
  options = { 
    browserType:     "chrome",  // "chrome, firefox" // WARNING: hit limit on number of detail pages with firefox
    headless:         false, //(process.env.PUP_HEADLESS == 'true'),     // run without windows
    manualLogin:      true,
    forceFullGather:  true,     // skip test for number of course
    scrollToBottom:   true,     // scroll page to bottom (WARNING: non-visible thumbnails are not loaded until page is scrolled)
    gatherCount:      9,    // max courses
    gatherThumbs:     true,     // copy thumbnails
    gatherDetails:    false,     // parse the details
    useSampleData:    false,     // skip browser and use sample data file
    saveSampleData:   false,     // save to sample data file
    screenshot:      false,     // take snapshots
    screenshotDir:   SCREENSHOT_DIR
  }

  console.log("env:");
  console.log(process.env.PUP_USERNAME);
  console.log(process.env.PUP_PASSWORD);
  console.log(process.env.PUP_HEADLESS);
  console.log("options:");
  console.log(options);

  const browser = await base.browser_init(options);
  if (!options.useSampleData) {
    options.version = await browser.version();
  }

  // login, get list of completed courses, logout
  data = {}
  
  await site.process_login(browser, options);
 
  await site.process_completed(browser, options, data);
  await site.process_logout(browser, options);
  await base.browser_close(browser);

  //DEBUG
  // console.log("data:");
  // console.log(JSON.stringify(data, null, space=2));

  if (data['completed-courses'].length > 0) {
    [totalH, totalM] = build_hours_minutes(data);
    data['completed-courses'].sort((a, b) => (a['completed-ts'] < b['completed-ts']) ? 1 : -1) // decending
    if (options.gatherCount < data['completed-courses'].length) {
      data['completed-courses'].splice(options.gatherCount-1, data['completed-courses'].length - options.gatherCount)
    }
    build_html(data, totalH, totalM);
    // build_md(data, totalH, totalM);
  }

  console.log("done.");
};

main();

  
  
