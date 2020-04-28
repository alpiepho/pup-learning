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
PAGE_WAIT_COMPLETED = 3000;
PAGE_WAIT_DETAILS = 1000;

const SAMPLE_FILE = "./artifacts/sample.json";


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
          }, 1000);
      });
  });
}

const process_course_details = async (browser, options, href) => {
  console.log("process_course_details");
  var newdata = {};
  newdata['linkedin'] = "";
  newdata['details'] = "";

  const page = await base.browser_get(browser, href, PAGE_WAIT_DETAILS);
  //console.log("process_course_details - after page")

  newdata = await page.evaluate(() => {
    let result = {};
    // parse: courses
    // TODO:
    //  - course toc
    //    - sections
    //      - title
    //      - subsections
    //        - title
    //        - description
    //        - durration
    //  - course exercise files?
    //  - **could** also grab transcript???
    // WARNING: with limit on number of detail pages, will need to start with clear sample.json and rebuild saved details if any more details are parsed

    result['linkedin'] = "";
    result['details'] = "";
    a = document.querySelectorAll('a.course-author-entity__meta-action');
    if (a.length) {
      result['linkedin'] = a[0].href;
    }
    a = document.querySelectorAll('.classroom-layout-panel-layout__main p');
    if (a.length) {
      result['details'] = a[0].innerText;
    }
    return result;
  });

  //console.log("process_course_details done");
  return [newdata['linkedin'], newdata['details']];
};

const process_completed = async (browser, options, data) => {
  //console.log("process_completed");
  var newdata;

  if (options.useSampleData) {
    newdata = sampleData;
  } else {
    const page = await base.browser_get(browser, PUP_URL_COMPLETED, PAGE_WAIT_COMPLETED);

    newdata = await page.evaluate(() => {
      let result = {};

      // parse: 'Learning History (108)'
      let count = document.querySelector('.me__content-tab--completed').innerText;
      result['count'] = count.replace(')','').split('(')[1];
      return result;
    });

    // check for optimization, of count is same, then we are done.
    if (!options.forceFullGather && sampleData['count'] == newdata['count']) {
      console.log("same expected course count, nothing to do.")
      data['completed-courses'] = []
      return;
    }
    
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

      // course links
      result['links'] = [...document.querySelectorAll('.lls-card-detail-card-body__headline a.card-entity-link')].map(elem => elem.href);
      result['titles'] = [...document.querySelectorAll('.lls-card-detail-card-body__headline a.card-entity-link')].map(elem => elem.innerText);
      result['authors'] = [...document.querySelectorAll('.lls-card-detail-card-body__primary-metadata .lls-card-authors span')].map(elem => elem.innerText);
      result['released'] = [...document.querySelectorAll('.lls-card-detail-card-body__primary-metadata span.lls-card-released-on')].map(elem => elem.innerText);
      result['duration'] = [...document.querySelectorAll('span.lls-card-duration-label')].map(elem => elem.innerText);
      result['completed'] = [...document.querySelectorAll('.lls-card-detail-card-body__footer span.lls-card-completion-state--completed')].map(elem => elem.innerText);
      result['imgs'] = [...document.querySelectorAll('.lls-card-entity-thumbnails__image img')].map(elem => elem.src);

      return result;
    });

  // assemble nested data from lists, assume collated
  var length;
  let expectedCount = parseInt(newdata['count']);
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
  length = newdata['imgs'].length;
  if (length != expectedCount) console.log("WARNING: links.imgs %d != %d", length, expectedCount);

  newdata['linkedin'] = [];
  newdata['details'] = [];
  if (options.preloadDetails) {
    newdata['linkedin'] = sampleData['linkedin'];
    newdata['details'] = sampleData['details'];
  }

  if (options.gatherDetails) {
    // HACK: found limit of 20-40 detail pages, will need to run this multiple times
    console.log("HACK: get next 10 detail pages");
    let newDetailCount = 0;
    for (i=0; i<newdata['links'].length && newDetailCount < 10; i++) {
      if (!newdata['details'][i]) {
          [temp1, temp2] = await process_course_details(browser, options, newdata['links'][i]);
          newdata['linkedin'].push(temp1);
          newdata['details'].push(temp2);
          newDetailCount += 1;
      }
    }
  }

  if (options.saveSampleData) {
    fs.writeFileSync(SAMPLE_FILE, JSON.stringify(newdata, null, 2));
  }
}

  data['completed-courses'] = []
  for (i=0; i<newdata['links'].length; i++) {
    entry = {}
    if (newdata['titles'][i]) {
      entry['title'] = newdata['titles'][i];
      entry['link'] = newdata['links'][i];
      entry['author'] = newdata['authors'][i];
      entry['released-date'] = newdata['released'][i];
      entry['duration'] = newdata['duration'][i];
      entry['completed-date'] = newdata['completed'][i];
      entry['img'] = newdata['imgs'][i];
      entry['linkedin']= newdata['linkedin'][i];
      entry['details']= newdata['details'][i];
      data['completed-courses'].push(entry);  
    }
  }
  //console.log("process_completed done");
};



exports.process_login = process_login;
exports.process_logout = process_logout;
exports.process_course_details = process_course_details;
exports.process_completed = process_completed;
