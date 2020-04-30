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
  await page.type('#auth-id-input', process.env.PUP_USERNAME);
  await base.delay(waitMs);
  await page.click('#auth-id-button'); // Email page "Continue"
  await base.delay(waitMs);
  await base.process_options(browser, options);
  await page.type('#password', process.env.PUP_PASSWORD);
  await base.delay(waitMs);
  await page.click('.btn__primary--large'); // Password page "Continue"
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

async function auto_scroll(page) {
  await page.evaluate(async () => {
    await new Promise((resolve, reject) => {
      var totalHeight = 0;
      var distance = 400;
      var timer = setInterval(() => {
        var scrollHeight = document.body.scrollHeight;
        window.scrollBy(0, distance);
        totalHeight += distance;

        if (totalHeight >= scrollHeight) {
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
    const page = await base.browser_get(
      browser,
      PUP_URL_COMPLETED,
      PAGE_WAIT_COMPLETED
    );

    newdata = await page.evaluate(() => {
      let result = {};

      // parse: 'Learning History (108)'
      let count = document.querySelector('.me__content-tab--completed').innerText;
      result['count'] = count.replace(')','').split('(')[1];
      return result;
    });

    // check for optimization, of count is same, then we are done.
    if (!options.forceFullGather && sampleData['count'] == newdata['count']) {
      console.log("same expected course count, nothing to do.");
      data['completed-courses'] = [];
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

      result['completed-courses'] = []
      let card_conts = document.querySelectorAll('.lls-card-detail-card');
      for (i=0; i<card_conts.length; i++) {
        let entry = {};
        entry['title'] = '';
        entry['link'] = '';
        entry['author'] = '';
        entry['released-date'] = '';
        entry['duration'] = '';
        entry['completed-date'] = '';
        entry['img'] = '';
        entry['linkedin'] = '';
        entry['details'] = '';
        entry['title'] = card_conts[i].querySelector('.lls-card-headline').innerText;
        entry['link'] = card_conts[i].querySelector('a.card-entity-link').href;
        temp = card_conts[i].querySelector('.lls-card-authors');
        if (temp) entry['author'] = temp.innerText.replace('By: ','');
        temp = card_conts[i].querySelector('.lls-card-released-on');
        if (temp) entry['released-date'] = temp.innerText.replace('Released ','');
        temp = card_conts[i].querySelector('.lls-card-duration-label');
        if (temp) entry['duration'] = temp.innerText;
        temp = card_conts[i].querySelector('.lls-card-completion-state--completed');
        if (temp) entry['completed-date'] = temp.innerText.replace('Completed ','');
        temp = card_conts[i].querySelector('img');
        if (temp) entry['img'] = temp.src;
        //entry["linkedin"] = newdata["linkedin"][i];
        //if (temp) entry['linkedin'] = temp;
        //temp = card_conts[i].querySelector('.meta-description');
        //if (temp) entry['details'] = temp.innerText;
        result['completed-courses'].push(entry);
      }
      return result;
    });

    if (options.preloadDetails) {
      for (i=0; i<newdata['completed-courses'].length; i++) {
        newdata['completed-courses'][i]['linkedin'] = sampleData['completed-courses'][i]['linkedin'];
        newdata['completed-courses'][i]['details'] = sampleData['completed-courses'][i]['details'];
      }
    }

    if (options.gatherDetails) {
      // HACK: found limit of 20-40 detail pages, will need to run this multiple times
      console.log("HACK: get next 10 detail pages");
      let newDetailCount = 0;
      for (i=0; i<newdata['completed-courses'].length && newDetailCount < 10; i++) {
        if (!newdata['completed-courses'][i]['details']) {
            [temp1, temp2] = await process_course_details(browser, options, newdata['completed-courses'][i]['link']);
            newdata['completed-courses'][i]['linkedin'] = temp1;
            newdata['completed-courses'][i]['details'] = temp2;
            newDetailCount += 1;
        }
      }
    }

    if (options.saveSampleData) {
      fs.writeFileSync(SAMPLE_FILE, JSON.stringify(newdata, null, 2));
    }
  }

   data['completed-courses'] = newdata['completed-courses'];
  //console.log("process_completed done");
};



exports.process_login = process_login;
exports.process_logout = process_logout;
exports.process_course_details = process_course_details;
exports.process_completed = process_completed;
