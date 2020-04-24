require("dotenv").config();
base = require('./base');
fs = require('fs');

PUP_URL_BASE="https://www.linkedin.com/learning";
PUP_URL_LOGIN=PUP_URL_BASE+"/me";
PUP_URL_LOGOUT=PUP_URL_BASE+"/logout";
PUP_URL_COMPLETED=PUP_URL_BASE+"/me/completed";


// in ms
PAGE_WAIT = 1000;
PAGE_WAIT_LOGIN = 2000;
PAGE_WAIT_LOGIN_DONE = 3000;
PAGE_WAIT_COMPLETED = 2000;

const SAMPLE_FILE = "./sample.json";


const sampleData = require(SAMPLE_FILE);

const process_login = async (browser, options) => {
  if (options.useSampleData) {
    return;
  }
  var waitMs = PAGE_WAIT_LOGIN + base.random_int(100);
  //console.log('process_login')
  const page = await base.browser_get(browser, PUP_URL_LOGIN, (waitMs));
  await base.process_options(browser, options);
  await page.type("#auth-id-input", process.env.PUP_USERNAME);
  await base.delay(waitMs);
  await page.click("#auth-id-button"); // Email page "Continue"
  await base.delay(waitMs);
  await base.process_options(browser, options);
  await page.type("#password", process.env.PUP_PASSWORD);
  await base.delay(waitMs);
  await page.click(".btn__primary--large"); // Password page "Continue"
  await base.delay(PAGE_WAIT_LOGIN_DONE);
  await base.process_options(browser, options);
  //console.log("process_login done")
};

const process_logout = async (browser, options) => {
  if (options.useSampleData) {
    return;
  }
  //console.log('process_logout')
  const page = await base.browser_get(browser, PUP_URL_LOGOUT, PAGE_WAIT);
  //console.log("process_logout done")
};

async function auto_scroll(page){
  await page.evaluate(async () => {
      await new Promise((resolve, reject) => {
          var totalHeight = 0;
          var distance = 400;
          var timer = setInterval(() => {
              var scrollHeight = document.body.scrollHeight;
              window.scrollBy(0, distance);
              totalHeight += distance;

              if(totalHeight >= scrollHeight){
                  clearInterval(timer);
                  resolve();
              }
          }, 500);
      });
  });
}

const process_completed = async (browser, options, data) => {
  //console.log("process_completed");
  var newdata;

  if (options.useSampleData) {
    newdata = sampleData;
  } else {
    const page = await base.browser_get(browser, PUP_URL_COMPLETED, PAGE_WAIT_COMPLETED);
    if (options.scrollToBottom) {
      await auto_scroll(page);
    }
    await base.delay(PAGE_WAIT_COMPLETED);
    await base.process_options(browser, options);

    newdata = await page.evaluate(() => {
      let result = {};

      // parse: 'Learning History (108)'
      let count = document.querySelector('.me__content-tab--completed').innerText;
      result['count'] = count.replace(')','').split('(')[1];

      // parse: table of completed courses
      // TODO:
      // - copy thumbnail
      // below requires navigation
      //  - course details
      //  - author LinkedIn link
      //  - course toc
      //    - sections
      //      - title
      //      - subsections
      //        - title
      //        - description
      //        - durration
      //  - course exercise files?
      //  - **could** also grab transcript???

      // course lings
      result['links'] = [...document.querySelectorAll('.lls-card-detail-card-body__headline a.card-entity-link')].map(elem => elem.href);
      result['titles'] = [...document.querySelectorAll('.lls-card-detail-card-body__headline a.card-entity-link')].map(elem => elem.innerText);
      result['authors'] = [...document.querySelectorAll('.lls-card-detail-card-body__primary-metadata .lls-card-authors span')].map(elem => elem.innerText);
      result['released'] = [...document.querySelectorAll('.lls-card-detail-card-body__primary-metadata span.lls-card-released-on')].map(elem => elem.innerText);
      result['duration'] = [...document.querySelectorAll('span.lls-card-duration-label')].map(elem => elem.innerText);
      result['completed'] = [...document.querySelectorAll('.lls-card-detail-card-body__footer span.lls-card-completion-state--completed')].map(elem => elem.innerText);

      return result;
    });
    if (options.saveSampleData) {
      fs.writeFileSync(SAMPLE_FILE, JSON.stringify(newdata, null, 2));
    }
  }

  // assemble nested data from lists, assume collated
  var length;
  let expectedCount = parseInt(newdata['count']);
  //let expectedCount = newdata['links'].length;
  length = newdata['links'].length;
  if (length != expectedCount) console.log("WARNING: links.length %d != %d", length, expectedCount);
  length = newdata['titles'].length;
  if (length != expectedCount) console.log("WARNING: titles.length %d != %d", length, expectedCount);
  length = newdata['authors'].length;
  if (length != expectedCount) console.log("WARNING: authors.length %d != %d", length, expectedCount);
  length = newdata['released'].length;
  if (length != expectedCount) console.log("WARNING: released.length %d != %d", length, expectedCount);
  length = newdata['duration'].length;
  if (length != expectedCount) console.log("WARNING: links.duration %d != %d", length, expectedCount);
  length = newdata['completed'].length;
  if (length != expectedCount) console.log("WARNING: links.completed %d != %d", length, expectedCount);

  data['completed-courses'] = []
  for (i=0; i<expectedCount; i++) {
    entry = {}
    if (newdata['titles'][i]) {
      entry['title'] = newdata['titles'][i];
      entry['link'] = newdata['links'][i];
      entry['author'] = newdata['authors'][i];
      entry['released-date'] = newdata['released'][i];
      entry['duration'] = newdata['duration'][i];
      entry['completed-date'] = newdata['completed'][i];
      data['completed-courses'].push(entry);  
    }
  }
  //console.log("process_completed done");
};

exports.process_login = process_login;
exports.process_logout = process_logout;
exports.process_completed = process_completed;
